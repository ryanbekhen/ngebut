package ngebut

import (
	"bytes"
	"context"
	"errors"
	"github.com/ryanbekhen/ngebut/internal/httpparser"
	"github.com/ryanbekhen/ngebut/log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/evanphx/wildcat"
	"github.com/panjf2000/gnet/v2"
)

type noopLogger struct{}

func (l *noopLogger) Debugf(format string, args ...interface{}) {}
func (l *noopLogger) Infof(format string, args ...interface{})  {}
func (l *noopLogger) Warnf(format string, args ...interface{})  {}
func (l *noopLogger) Errorf(format string, args ...interface{}) {}
func (l *noopLogger) Fatalf(format string, args ...interface{}) {}

// Server represents an HTTP server.
type Server struct {
	httpServer            *httpServer
	router                *Router
	disableStartupMessage bool
	errorHandler          Handler // Handler called when an error occurs during request processing
}

type httpServer struct {
	gnet.BuiltinEventEngine

	addr         string
	multicore    bool
	router       *Router
	eng          gnet.Engine
	errorHandler Handler // Handler called when an error occurs during request processing

	readTimeout  time.Duration // Read timeout for requests
	writeTimeout time.Duration // Write timeout for responses
	idleTimeout  time.Duration // Idle timeout for connections
}

// defaultErrorHandler is the default handler for errors.
// It returns a plain text response with the error message.
// If the error is an HttpError, it uses the status code from the HttpError.
// If the status code is already set to a 4xx or 5xx status code, it respects that.
func defaultErrorHandler(c *Ctx) {
	err := c.GetError()
	statusCode := c.StatusCode()

	// Check if the error is an HttpError
	var httpErr *HttpError
	if errors.As(err, &httpErr) {
		statusCode = httpErr.Code
	}

	c.Status(statusCode)
	c.String("%v", err)
}

// New creates a new server with the given configuration.
// This is the main entry point for creating a ngebut server instance.
//
// Parameters:
//   - config: The server configuration (use DefaultConfig() for sensible defaults)
//
// Returns:
//   - A new Server instance ready to be configured with routes and middleware
func New(config ...Config) *Server {
	r := NewRouter()

	// Use default config if none provided
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	hs := &httpServer{
		addr:         "",
		multicore:    true,
		router:       r,
		errorHandler: cfg.ErrorHandler,
		readTimeout:  cfg.ReadTimeout,
		writeTimeout: cfg.WriteTimeout,
		idleTimeout:  cfg.IdleTimeout,
	}

	return &Server{
		httpServer:            hs,
		router:                r,
		disableStartupMessage: cfg.DisableStartupMessage,
		errorHandler:          cfg.ErrorHandler,
	}
}

func (hs *httpServer) OnBoot(eng gnet.Engine) gnet.Action {
	hs.eng = eng
	return gnet.None
}

func (hs *httpServer) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	c.SetContext(&httpparser.Codec{Parser: wildcat.NewHTTPParser()})
	return nil, gnet.None
}

// requestPool is a pool of Request objects for reuse
var requestPool = sync.Pool{
	New: func() interface{} {
		return &Request{}
	},
}

// getRequest gets a Request from the pool and initializes it with the given http.Request
func getRequest(r *http.Request) *Request {
	// Use the existing NewRequest function which already handles all the initialization
	return NewRequest(r)
}

// releaseRequest returns a Request to the pool
func releaseRequest(r *Request) {
	// Reset all fields to zero values
	r.Method = ""
	r.URL = nil
	r.Proto = ""

	// Clear the header map
	for k := range r.Header {
		delete(r.Header, k)
	}

	// Clear the body
	r.Body = nil
	r.ContentLength = 0
	r.Host = ""
	r.RemoteAddr = ""
	r.RequestURI = ""
	r.ctx = nil

	// Return to the pool
	requestPool.Put(r)
}

func (hs *httpServer) OnTraffic(c gnet.Conn) gnet.Action {
	hc := c.Context().(*httpparser.Codec)
	buf, _ := c.Peek(-1)
	n := len(buf)
	var processed int

	// Get reusable readers from the pool
	reader := httpparser.GetReader()
	defer httpparser.ReleaseReader(reader)

	bytesReader := httpparser.GetBytesReader()
	defer httpparser.ReleaseBytesReader(bytesReader)

	for processed < n {
		// Parse the request
		nextOffset, body, err := hc.Parse(buf[processed:])
		hc.ResetParser()

		// Break early if we can't parse any more data
		if err != nil {
			// Don't log parse errors for normal HTTP requests
			// This prevents log spam when clients send valid requests that result in 405 responses

			// For form submissions that might be causing issues, send a 400 Bad Request response
			contentType := ""
			if processed+8 < n {
				// Look for Content-Type header in the buffer
				contentTypeIndex := bytes.Index(buf[processed:], []byte("Content-Type:"))
				if contentTypeIndex != -1 {
					endIndex := bytes.Index(buf[processed+contentTypeIndex:], []byte("\r\n"))
					if endIndex != -1 && endIndex < 100 { // Limit search to avoid buffer overruns
						contentType = string(buf[processed+contentTypeIndex+13 : processed+contentTypeIndex+endIndex])
						contentType = strings.TrimSpace(contentType)
					}
				}
			}

			// If this looks like a form submission, send a more helpful response
			if strings.Contains(contentType, "multipart/form-data") ||
				strings.Contains(contentType, "application/x-www-form-urlencoded") {
				parserHeaders := getParserHeaders()
				defer releaseParserHeaders(parserHeaders)
				errorMsg := []byte("Bad Request: Form data could not be processed. Please check your form submission.")
				hc.WriteResponse(http.StatusBadRequest, parserHeaders, errorMsg)
				c.Write(hc.Buf)
			}

			// Discard at least 1 byte to avoid getting stuck in a loop
			// This ensures we make progress even if we can't parse the request
			if processed < n {
				processed++
			}
			break
		}

		// Break if we don't have enough data
		if len(buf[processed:]) < nextOffset {
			break
		}

		// Reset and reuse the readers
		bytesReader.Reset(buf[processed : nextOffset+processed])
		reader.Reset(bytesReader)

		// Parse the HTTP request
		httpReq, err := http.ReadRequest(reader)
		if err != nil {
			// Discard at least 1 byte to avoid getting stuck in a loop
			// This ensures we make progress even if we can't parse the request
			if processed < n {
				processed++
			}
			break
		}

		// Set the body if it's not nil
		if body != nil {
			httpReq.Body = httpparser.GetBodyReader(body)
		}

		// Create a Request object from the *http.Request
		req := getRequest(httpReq)

		// Process the request
		processRequest(hs, hc, req, c)

		// Release the Request back to the pool
		releaseRequest(req)

		// Update processed count
		processed += nextOffset

		// If there's no more data to process, break
		if nextOffset == 0 {
			// Discard at least 1 byte to avoid getting stuck in a loop
			// This ensures we make progress even if we can't parse the request
			if processed < n {
				processed++
			}
			break
		}
	}

	// Write the response if there's data in the buffer
	if len(hc.Buf) > 0 {
		c.Write(hc.Buf)
	}

	// Reset the codec for the next request
	hc.Reset()

	// Discard processed data
	if processed > 0 {
		c.Discard(processed)
	}

	return gnet.None
}

// OnClose is called when a connection is closed
func (hs *httpServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	// Release the codec back to the pool
	if codec, ok := c.Context().(*httpparser.Codec); ok && codec != nil {
		httpparser.ReleaseCodec(codec)
	}
	return gnet.None
}

// dummyResponseWriter is used as a placeholder when creating a Ctx that will handle its own response writing
// but still needs to track headers correctly
type dummyResponseWriter struct {
	header http.Header
}

// dummyWriterPool is a pool of dummyResponseWriter objects for reuse
var dummyWriterPool = sync.Pool{
	New: func() interface{} {
		return &dummyResponseWriter{
			header: make(http.Header),
		}
	},
}

// getDummyWriter gets a dummyResponseWriter from the pool
func getDummyWriter() *dummyResponseWriter {
	return dummyWriterPool.Get().(*dummyResponseWriter)
}

// releaseDummyWriter returns a dummyResponseWriter to the pool
func releaseDummyWriter(d *dummyResponseWriter) {
	// Clear the header map
	for k := range d.header {
		delete(d.header, k)
	}
	dummyWriterPool.Put(d)
}

func (d *dummyResponseWriter) Header() http.Header {
	return d.header
}

func (d *dummyResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (d *dummyResponseWriter) WriteHeader(statusCode int) {
	// No-op
}

func (d *dummyResponseWriter) Flush() {
	// No-op
}

// headerPool is a pool of Header objects for reuse
var headerPool = sync.Pool{
	New: func() interface{} {
		return make(Header)
	},
}

// getHeader gets a Header from the pool
func getHeader() Header {
	return headerPool.Get().(Header)
}

// releaseHeader returns a Header to the pool
func releaseHeader(h Header) {
	for k := range h {
		delete(h, k)
	}
	headerPool.Put(h)
}

// parserHeadersPool is a pool of httpparser.Header objects for reuse
var parserHeadersPool = sync.Pool{
	New: func() interface{} {
		return make(httpparser.Header)
	},
}

// getParserHeaders gets a httpparser.Header from the pool
func getParserHeaders() httpparser.Header {
	return parserHeadersPool.Get().(httpparser.Header)
}

// releaseParserHeaders returns a httpparser.Header to the pool
func releaseParserHeaders(h httpparser.Header) {
	for k := range h {
		delete(h, k)
	}
	parserHeadersPool.Put(h)
}

func processRequest(hs *httpServer, hc *httpparser.Codec, req *Request, c gnet.Conn) {
	req.RemoteAddr = c.RemoteAddr().String()

	if req.ContentLength <= 0 && hc.ContentLength > 0 {
		req.ContentLength = int64(hc.ContentLength)
	}

	// Get a dummyWriter from the pool
	dummyWriter := getDummyWriter()
	defer releaseDummyWriter(dummyWriter)

	ctx := getContextFromRequest(dummyWriter, req)
	defer ReleaseContext(ctx)

	// Set server header directly in context header
	ctx.Set("Server", "ngebut")

	// Process the request
	hs.router.ServeHTTP(ctx, ctx.Request)

	// Handle errors
	if err := ctx.GetError(); err != nil {
		if hs.errorHandler != nil {
			hs.errorHandler(ctx)
		} else {
			defaultErrorHandler(ctx)
		}
	}

	// Ensure headers set after c.Next() in middleware are included in the response
	if ctx.Writer != nil {
		ctx.Writer.Flush()
	}

	// Get a parserHeader from the pool
	parserHeaders := getParserHeaders()
	defer releaseParserHeaders(parserHeaders)

	// Directly copy headers from dummyWriter to parserHeaders
	for k, values := range dummyWriter.header {
		if len(values) > 0 {
			parserHeaders[k] = values
		}
	}

	// Then copy headers from context (overriding any with same name)
	for k, values := range ctx.Request.Header {
		if len(values) > 0 {
			parserHeaders[k] = values
		}
	}

	// Handle HEAD requests specially per HTTP spec
	if ctx.Request.Method == MethodHead {
		if ctx.statusCode == StatusInternalServerError {
			ctx.statusCode = StatusOK
		}
		hc.WriteResponse(ctx.statusCode, parserHeaders, nil)
	} else {
		hc.WriteResponse(ctx.statusCode, parserHeaders, ctx.body)
	}
}

func (s *Server) Router() *Router {
	return s.router
}

// Listen starts the server and listens for incoming connections.
func (s *Server) Listen(addr string) error {
	// Clean up the address to ensure it is in the correct format
	if addr == "" {
		addr = ":3000" // Default address if none provided
	}

	// Set the address in the httpServer struct
	s.httpServer.addr = "tcp://" + addr

	// Initialize the logger
	initLogger(log.InfoLevel)

	// Display startup message if not disabled
	if !s.disableStartupMessage {
		displayStartupMessage(addr)
	}

	// Start the server directly
	return gnet.Run(
		s.httpServer,
		s.httpServer.addr,
		gnet.WithMulticore(s.httpServer.multicore),
		gnet.WithReuseAddr(true),
		gnet.WithReusePort(true),
		gnet.WithLogger(&noopLogger{}),
		gnet.WithTCPNoDelay(gnet.TCPNoDelay),
		gnet.WithTCPKeepAlive(s.httpServer.idleTimeout),
		gnet.WithReadBufferCap(int(s.httpServer.readTimeout.Seconds())*1024),
		gnet.WithWriteBufferCap(int(s.httpServer.writeTimeout.Seconds())*1024),
	)
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.eng.Stop(ctx)
}

// GET registers a new route with the GET method.
func (s *Server) GET(pattern string, handlers ...Handler) *Router {
	return s.router.GET(pattern, handlers...)
}

// HEAD registers a new route with the HEAD method.
func (s *Server) HEAD(pattern string, handlers ...Handler) *Router {
	return s.router.HEAD(pattern, handlers...)
}

// POST registers a new route with the POST method.
func (s *Server) POST(pattern string, handlers ...Handler) *Router {
	return s.router.POST(pattern, handlers...)
}

// PUT registers a new route with the PUT method.
func (s *Server) PUT(pattern string, handlers ...Handler) *Router {
	return s.router.PUT(pattern, handlers...)
}

// DELETE registers a new route with the DELETE method.
func (s *Server) DELETE(pattern string, handlers ...Handler) *Router {
	return s.router.DELETE(pattern, handlers...)
}

// CONNECT registers a new route with the CONNECT method.
func (s *Server) CONNECT(pattern string, handlers ...Handler) *Router {
	return s.router.CONNECT(pattern, handlers...)
}

// OPTIONS registers a new route with the OPTIONS method.
func (s *Server) OPTIONS(pattern string, handlers ...Handler) *Router {
	return s.router.OPTIONS(pattern, handlers...)
}

// TRACE registers a new route with the TRACE method.
func (s *Server) TRACE(pattern string, handlers ...Handler) *Router {
	return s.router.TRACE(pattern, handlers...)
}

// PATCH registers a new route with the PATCH method.
func (s *Server) PATCH(pattern string, handlers ...Handler) *Router {
	return s.router.PATCH(pattern, handlers...)
}

// Use adds middleware to the router.
func (s *Server) Use(middleware ...interface{}) {
	s.router.Use(middleware...)
}

// NotFound sets the handler for requests that don't match any route.
func (s *Server) NotFound(handler Handler) {
	s.router.NotFound = handler
}

// Group creates a new route group with the given prefix.
func (s *Server) Group(prefix string) *Group {
	return s.router.Group(prefix)
}
