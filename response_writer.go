package ngebut

// ResponseWriter is the interface used by an HTTP handler to construct an HTTP response.
type ResponseWriter interface {
	// Header returns the header map that will be sent by WriteHeader.
	// The Header map also is the mechanism with which
	// Handlers can set HTTP trailers.
	Header() Header

	// Write writes the data to the connection as part of an HTTP reply.
	Write([]byte) (int, error)

	// WriteHeader sends an HTTP response header with the provided
	// status code.
	WriteHeader(statusCode int)

	// Flush writes the buffered response to the underlying writer.
	Flush()
}
