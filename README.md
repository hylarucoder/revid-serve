# revid-serve

A lightweight media server for Remotion projects that helps manage media assets outside of the bundle.

## Why?

- Keep your Remotion bundle size small
- Manage media files separately from your bundle
- Simple and efficient media serving
- No need to bundle large media files in `./public`

## Install

### Manual Install

If you prefer not to use the automatic installation script, you can download and install manually:

1. Visit the [Releases](https://github.com/hylarucoder/revid-serve/releases/latest) page
2. Download the version for your system
3. Rename and make it executable:
   ```bash
   mv revid-serve-* revid-serve
   chmod +x revid-serve
   ```

## Usage

1. Start the media server:

   ```bash
   ./revid-serve -d ./media
   ```

2. Add the `mediaFile` utility function:

   ```ts
   function mediaFile(path: string) {
     // change to your media server base url
     return `http://localhost:8080/${path}`;
   }
   ```

3. Use it in your code:
   ```ts
   const video = mediaFile("videos/intro.mp4");
   const image = mediaFile("images/background.png");
   ```
