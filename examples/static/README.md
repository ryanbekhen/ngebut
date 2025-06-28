# ngebut Examples

This directory contains examples demonstrating various features of the ngebut HTTP framework.

## Static File Server Example

### Overview

The `static_server.go` example demonstrates the complete static file serving functionality of ngebut, including:

- ✅ Basic static file serving
- ✅ Custom URL prefixes
- ✅ Directory browsing
- ✅ File download mode
- ✅ Conditional serving with Next function
- ✅ Custom headers and response modification
- ✅ MIME type detection
- ✅ Byte range requests (for video/audio streaming)
- ✅ Index file support

### How to Run

```bash
# From the project root directory
go run examples/static/static_server.go
```

The server will start on `http://localhost:3000` and automatically create the necessary static files for demonstration.

### Available Endpoints

| Endpoint                             | Description                                       |
| ------------------------------------ | ------------------------------------------------- |
| `http://localhost:3000/`             | Basic static files (serves index.html)            |
| `http://localhost:3000/public/`      | Static files with `/public` prefix                |
| `http://localhost:3000/assets/`      | Advanced static with directory browsing enabled   |
| `http://localhost:3000/conditional/` | Conditional static serving (skips .private files) |
| `http://localhost:3000/downloads/`   | Force download mode                               |
| `http://localhost:3000/api/info`     | API endpoint for testing                          |

### Features Demonstrated

#### 1. Basic Static File Serving

```go
app.STATIC("/", "./examples/static/assets")
```

#### 2. Advanced Configuration

```go
app.STATIC("/assets/", "./examples/static/assets", ngebut.Static{
    Browse:    true,         // Enable directory browsing
    Download:  false,        // Don't force downloads
    Index:     "index.html", // Default index file
    MaxAge:    3600,         // Cache for 1 hour
    ByteRange: true,         // Enable byte range requests
    ModifyResponse: func(c *ngebut.Ctx) {
        // Add custom headers
        c.Set("X-Static-Server", "ngebut-example")
    },
})
```

#### 3. Conditional Serving

```go
app.STATIC("/conditional/", "./examples/static/assets", ngebut.Static{
    Browse: true,
    Next: func(c *ngebut.Ctx) bool {
        // Skip files with .private extension
        if filepath.Ext(c.Path()) == ".private" {
            return true // Skip this middleware
        }
        return false // Process with static serving
    },
})
```

#### 4. Download Mode

```go
app.STATIC("/downloads/", "./examples/static/assets", ngebut.Static{
    Download: true, // Forces Content-Disposition: attachment
    Browse:   true,
})
```

### Static Configuration Options

| Option           | Type              | Description                     | Default            |
| ---------------- | ----------------- | ------------------------------- | ------------------ |
| `Compress`       | `bool`            | Enable file compression         | `false`            |
| `ByteRange`      | `bool`            | Enable byte range requests      | `false`            |
| `Browse`         | `bool`            | Enable directory browsing       | `false`            |
| `Download`       | `bool`            | Force file downloads            | `false`            |
| `Index`          | `string`          | Index file name                 | `"index.html"`     |
| `CacheDuration`  | `time.Duration`   | Cache duration for files        | `10 * time.Second` |
| `MaxAge`         | `int`             | Cache-Control max-age (seconds) | `0`                |
| `ModifyResponse` | `Handler`         | Custom response modifier        | `nil`              |
| `Next`           | `func(*Ctx) bool` | Skip middleware condition       | `nil`              |

### File Structure

The example automatically creates the following structure:

```
examples/
├── static_server.go          # Main example file
├── static/                   # Static files directory
│   ├── index.html           # Homepage with feature showcase
│   ├── sample.txt           # Sample text file for downloads
│   ├── secret.private       # Private file (blocked by conditional serving)
│   ├── css/
│   │   └── style.css        # Stylesheet
│   └── js/
│       └── app.js           # JavaScript file
└── README.md                # This file
```

### Testing the Features

1. **Basic File Serving**: Visit `http://localhost:3000/` to see the homepage
2. **Directory Browsing**: Visit `http://localhost:3000/assets/` to browse files
3. **CSS/JS Serving**: Check browser dev tools to see CSS and JS files loading
4. **Download Mode**: Visit `http://localhost:3000/downloads/sample.txt` - file will download
5. **Conditional Serving**: Try accessing `http://localhost:3000/conditional/secret.private` (should fail)
6. **Custom Headers**: Check response headers on `/assets/` files for `X-Static-Server` header
7. **MIME Types**: Different file types should have appropriate `Content-Type` headers

### Performance Features

- **Object Pooling**: Reuses buffers and objects to reduce garbage collection
- **Efficient Routing**: Fast pattern matching for static routes
- **Range Requests**: Supports partial content for streaming media
- **Caching Headers**: Proper cache control headers for better client-side caching

### Security Features

- **Path Traversal Prevention**: Prevents access to files outside the root directory
- **Conditional Access**: Next function allows custom access control
- **Safe File Serving**: Only serves files, not directories (unless browsing is enabled)

## Running Multiple Examples

To avoid port conflicts when running multiple examples, you can modify the port in each example file or run them on different terminals one at a time.

## Development

Feel free to modify the examples to test different configurations and features. The static file server is highly configurable and supports many use cases from simple file serving to complex content delivery scenarios.
