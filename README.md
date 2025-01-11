# revid-serve

When creating numerous Remotion templates, bundling a large amount of media in `./public` directory to assets can become cumbersome. This is because each time you bundle, all the media in the public folder is included in the final bundle assets, causing the bundle size to grow continuously.

To address this issue and improve the developer experience (DX), we've created a simple tool that offers a solution similar to the existing workflow.

With this tool, you can:

1. Store your media files in the `./media` folder instead of `./public`.
2. Use the `mediaFile` utility function in place of `staticFile`.

This approach allows for more efficient media management and helps keep your bundle size under control.

(Please ensure you maintain the proper indentation when implementing this in your code.)



```ts
function mediaFile(path: string) {
    // change to your media server base url
    return `http://localhost:8080/${path}`;
}
```

## Usage

1. Start the media server:
```bash
revid-serve -d ./media
```

2. Use the `mediaFile` function in your code:
```ts
// Example
const video = mediaFile('videos/intro.mp4');