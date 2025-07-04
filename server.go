package ngebut

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ryanbekhen/ngebut/internal/httpparser"
	"github.com/ryanbekhen/ngebut/log"

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
		return &Request{
			Header: NewHeader(),
			ctx:    context.Background(),
		}
	},
}

// getRequest gets a Request from the pool and initializes it with the given http.Request
func getRequest(r *http.Request) *Request {
	// Get a Request from the pool
	req := requestPool.Get().(*Request)

	if r == nil {
		// Just return the initialized Request from the pool
		return req
	}

	// Initialize the Request with the given http.Request
	// Read the request body if it exists
	var body []byte
	if r.Body != nil {
		// Get a buffer from the pool
		buf := requestBodyBufferPool.Get().(*bytes.Buffer)
		buf.Reset()

		// Read the body into the buffer
		_, err := io.Copy(buf, r.Body)
		if err == nil {
			// Close the original body
			_ = r.Body.Close()

			// Get the bytes from the buffer
			body = buf.Bytes()

			// Create a new ReadCloser so the body can be read again if needed
			// Use the same buffer to avoid allocation
			r.Body = io.NopCloser(bytes.NewReader(body))
		}

		// Return the buffer to the pool
		requestBodyBufferPool.Put(buf)
	}

	// Initialize the Request fields
	req.Method = r.Method
	req.URL = r.URL
	req.Proto = r.Proto

	// Handle headers more efficiently
	if req.Header == nil {
		// If the request doesn't have a header map, create a new one
		req.Header = NewHeader()
	}

	// Update the existing header map with values from the request
	// This avoids allocating a new map
	UpdateHeaderFromMap(req.Header, r.Header)

	req.Body = body
	req.ContentLength = r.ContentLength
	req.Host = r.Host
	req.RemoteAddr = r.RemoteAddr
	req.RequestURI = r.RequestURI
	req.ctx = r.Context()

	return req
}

// releaseRequest returns a Request to the pool
func releaseRequest(r *Request) {
	// Reset all fields to zero values
	r.Method = ""
	r.URL = nil
	r.Proto = ""

	// Clear the header map
	if r.Header != nil {
		for k := range *r.Header {
			delete(*r.Header, k)
		}
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
				contentTypeIndex := bytes.Index(buf[processed:], []byte(HeaderContentType+":"))
				if contentTypeIndex != -1 {
					endIndex := bytes.Index(buf[processed+contentTypeIndex:], []byte("\r\n"))
					if endIndex != -1 && endIndex < 100 { // Limit search to avoid buffer overruns
						contentType = string(buf[processed+contentTypeIndex+len(HeaderContentType)+1 : processed+contentTypeIndex+endIndex])
						contentType = strings.TrimSpace(contentType)
					}
				}
			}

			// If this looks like a form submission, send a more helpful response
			if strings.Contains(contentType, MIMEMultipartForm) ||
				strings.Contains(contentType, MIMEApplicationForm) {
				parserHeaders := getParserHeaders()
				defer releaseParserHeaders(parserHeaders)
				errorMsg := []byte("Bad Request: Form data could not be processed. Please check your form submission.")
				hc.WriteResponse(StatusBadRequest, parserHeaders, errorMsg)
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

// responseRecorder is an optimized response writer for benchmarks and internal use
// It implements http.ResponseWriter with minimal allocations
// Similar to httptest.ResponseRecorder but optimized for our use case
type responseRecorder struct {
	header http.Header
	body   []byte
	code   int
}

// Pre-allocate a shared empty header for zero-allocation initialization
var emptyHeader = make(http.Header)

// responseRecorderPool is a pool of responseRecorder objects for reuse
var responseRecorderPool = sync.Pool{
	New: func() interface{} {
		return &responseRecorder{
			header: make(http.Header, 16), // Pre-allocate with larger capacity for headers
			body:   make([]byte, 0, 4096), // 4KB initial capacity to reduce reallocations
			code:   StatusOK,
		}
	},
}

// getResponseRecorder gets a responseRecorder from the pool
func getResponseRecorder() *responseRecorder {
	return responseRecorderPool.Get().(*responseRecorder)
}

// releaseResponseRecorder returns a responseRecorder to the pool
func releaseResponseRecorder(w *responseRecorder) {
	// For small header maps, clear them
	// For larger ones, replace with a new one to avoid expensive iteration
	if len(w.header) > 32 {
		w.header = make(http.Header, 16)
	} else if len(w.header) > 0 {
		// Clear the header map efficiently
		for k := range w.header {
			delete(w.header, k)
		}
	}

	// Reset the body efficiently
	// If the capacity is too large, replace it to avoid holding onto memory
	if cap(w.body) > 16384 { // 16KB
		w.body = make([]byte, 0, 4096) // 4KB
	} else {
		w.body = w.body[:0]
	}

	// Reset the status code
	w.code = StatusOK

	// Return to pool
	responseRecorderPool.Put(w)
}

func (w *responseRecorder) Header() http.Header {
	return w.header
}

func (w *responseRecorder) Write(b []byte) (int, error) {
	// For benchmarks, we can skip storing the data
	// But for actual requests, we need to store it
	if len(b) > 0 {
		// Use append directly which will handle capacity growth efficiently
		// This avoids the manual capacity check and buffer creation
		w.body = append(w.body, b...)
	}
	return len(b), nil
}

func (w *responseRecorder) WriteHeader(statusCode int) {
	w.code = statusCode
}

func (w *responseRecorder) Flush() {
	// No-op for benchmarks
}

// headerPool is a pool of Header objects for reuse
var headerPool = sync.Pool{
	New: func() interface{} {
		return NewHeader()
	},
}

// getHeader gets a Header from the pool
func getHeader() *Header {
	return headerPool.Get().(*Header)
}

// releaseHeader returns a Header to the pool
func releaseHeader(h *Header) {
	for k := range *h {
		delete(*h, k)
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

	// Get a responseRecorder from the pool
	recorder := getResponseRecorder()
	defer releaseResponseRecorder(recorder)

	ctx := getContextFromRequest(recorder, req)
	defer ReleaseContext(ctx)

	// Set server header directly in context header
	ctx.Set(HeaderServer, "ngebut")

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

	// Directly copy headers from responseRecorder to parserHeaders
	for k, values := range recorder.header {
		if len(values) > 0 {
			parserHeaders[k] = values
		}
	}

	// Then copy headers from context (overriding any with same name)
	if ctx.Request.Header != nil {
		for k, values := range *ctx.Request.Header {
			if len(values) > 0 {
				parserHeaders[k] = values
			}
		}
	}

	// Handle HEAD requests specially per HTTP spec
	if ctx.Request.Method == MethodHead {
		if ctx.statusCode == StatusInternalServerError {
			ctx.statusCode = StatusOK
		}
		hc.WriteResponse(ctx.statusCode, parserHeaders, nil)
	} else {
		hc.WriteResponse(ctx.statusCode, parserHeaders, recorder.body)
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

// STATIC registers a new route with the GET method.
func (s *Server) STATIC(prefix, root string, config ...Static) *Router {
	return s.router.STATIC(prefix, root, config...)
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
