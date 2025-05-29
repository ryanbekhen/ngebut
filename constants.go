package ngebut

// HTTP protocol terminators
var (
	// crlf represents the HTTP header terminator in a byte slice
	crlf = []byte{0x0d, 0x0a, 0x0d, 0x0a}

	// lastChunk represents the end of a chunked HTTP response in a byte slice
	lastChunk = []byte{0x30, 0x0d, 0x0a, 0x0d, 0x0a} // "0\r\n\r\n"
)
