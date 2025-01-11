package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

//go:embed web/dist
var webContent embed.FS

var (
	port int
	dir  string
	// Version information (set by build flag)
	Version = "dev"
)

// FileInfo represents a file or directory for the template
type FileInfo struct {
	Name string
	Size string
}

// DirectoryData represents the data passed to the directory listing template
type DirectoryData struct {
	Path  string
	Dirs  []FileInfo
	Files []FileInfo
}

// SecureFileSystem wraps http.FileSystem with path security checks
type secureFileSystem struct {
	root       string
	fileSystem http.FileSystem
}

func (fs *secureFileSystem) Open(name string) (http.File, error) {
	// Clean the path
	name = filepath.Clean(name)

	// Get absolute path
	absPath := filepath.Join(fs.root, name)

	// Check if the path is within allowed directory
	if !strings.HasPrefix(absPath, fs.root) {
		log.Warn().Str("path", name).Msg("Attempted to access unauthorized path")
		return nil, os.ErrPermission
	}

	return fs.fileSystem.Open(name)
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Write implements http.ResponseWriter
func (w *responseWriter) Write(b []byte) (int, error) {
	// If WriteHeader hasn't been called, we need to set the default status
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

// Logging middleware for HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture status
		wrapped := &responseWriter{
			ResponseWriter: w,
			status:         0, // Will be set when Write or WriteHeader is called
		}

		next.ServeHTTP(wrapped, r)

		// Log request details after completion
		log.Info().
			Str("Method", r.Method).
			Str("Path", r.URL.Path).
			Str("Remote", r.RemoteAddr).
			Int("Status", wrapped.status).
			Dur("Duration", time.Since(start)).
			Msg("Req")
	})
}

// DirectoryHandler creates a handler that serves directory listings
func DirectoryHandler(root http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Clean the URL path
		urlPath := filepath.Clean(r.URL.Path)

		// Try to open the file/directory
		f, err := root.Open(urlPath)
		if err != nil {
			if os.IsPermission(err) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			if os.IsNotExist(err) {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer f.Close()

		// Get file info
		fi, err := f.Stat()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// If it's not a directory, serve the file directly
		if !fi.IsDir() {
			http.FileServer(root).ServeHTTP(w, r)
			return
		}

		// For directories, generate the listing
		files, err := f.(http.File).Readdir(-1)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Sort files and directories
		sort.Slice(files, func(i, j int) bool {
			// If both are directories or both are files, sort by name
			if files[i].IsDir() == files[j].IsDir() {
				return files[i].Name() < files[j].Name()
			}
			// Directories come first
			return files[i].IsDir()
		})

		// Prepare data for template
		data := DirectoryData{
			Path:  urlPath,
			Dirs:  make([]FileInfo, 0),
			Files: make([]FileInfo, 0),
		}

		// Separate directories and files
		for _, file := range files {
			if file.IsDir() {
				data.Dirs = append(data.Dirs, FileInfo{
					Name: file.Name(),
					Size: "",
				})
			} else {
				data.Files = append(data.Files, FileInfo{
					Name: file.Name(),
					Size: humanize.Bytes(uint64(file.Size())),
				})
			}
		}

		// Load and execute template
		tmpl, err := template.ParseFS(webContent, "web/dist/directory.html")
		if err != nil {
			log.Error().Err(err).Msg("Error loading template")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Error().Err(err).Msg("Error executing template")
			// Since we've already started writing the response, we can't change the status code now
			return
		}
	})
}

func serve(cmd *cobra.Command, args []string) error {
	// Check if directory is provided
	if dir == "" {
		return fmt.Errorf("directory parameter is required")
	}

	// Get absolute path for the data directory
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", absPath)
	}

	// Create secure file system for user data
	dataFS := &secureFileSystem{
		root:       absPath,
		fileSystem: http.Dir(absPath),
	}

	// Create multiplexer for handling both embedded and data files
	mux := http.NewServeMux()

	// Serve user data under /data/ path with directory listing
	mux.Handle("/", DirectoryHandler(dataFS))

	// Add logging middleware
	handler := loggingMiddleware(mux)

	// Add CORS middleware
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "HEAD", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		MaxAge:         86400, // 24 hours
	}).Handler(handler)

	// Start server
	addr := fmt.Sprintf(":%d", port)
	log.Info().Msgf("Starting server... ")
	log.Info().
		Msgf("Server is running at http://localhost:%d", port)

	return http.ListenAndServe(addr, corsHandler)
}

func main() {
	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	})

	rootCmd := &cobra.Command{
		Use:     "revid-serve -d <directory>",
		Version: Version,
		Short:   "A simple static file server",
		Long: `revid-serve is a secure static file server.

Example:
  revid-serve -d ./media     Serve files from ./media directory
  revid-serve -d . -p 8080   Serve current directory on port 8080`,
		Example: `  revid-serve -d ./media
  revid-serve -d . -p 8080`,
		RunE: serve,
	}

	// Add command line flags
	rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port number (default: 8080)")
	rootCmd.Flags().StringVarP(&dir, "dir", "d", "", "Directory to serve (required)")
	rootCmd.MarkFlagRequired("dir")

	if err := rootCmd.Execute(); err != nil {
		if err.Error() == "required flag(s) \"dir\" not set" {
			log.Fatal().Msg("Error: Directory is required. Use -d flag to specify a directory to serve.\nExample: revid-serve -d ./media")
		} else {
			log.Fatal().Err(err).Msg("Failed to execute command")
		}
	}
}
