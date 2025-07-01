package ngebut

const (
	statusMessageMin = 100
	statusMessageMax = 511
)

// HTTP status codes were borrowed from valyala/fasthttp
var (
	unknownStatusCode = "Unknown Status Code"

	statusMessages = []string{
		StatusContinue:           "Continue",
		StatusSwitchingProtocols: "Switching Protocols",
		StatusProcessing:         "Processing",
		StatusEarlyHints:         "Early Hints",

		StatusOK:                   "OK",
		StatusCreated:              "Created",
		StatusAccepted:             "Accepted",
		StatusNonAuthoritativeInfo: "Non-Authoritative Information",
		StatusNoContent:            "No Content",
		StatusResetContent:         "Reset Content",
		StatusPartialContent:       "Partial Content",
		StatusMultiStatus:          "Multi-Status",
		StatusAlreadyReported:      "Already Reported",
		StatusIMUsed:               "IM Used",

		StatusMultipleChoices:   "Multiple Choices",
		StatusMovedPermanently:  "Moved Permanently",
		StatusFound:             "Found",
		StatusSeeOther:          "See Other",
		StatusNotModified:       "Not Modified",
		StatusUseProxy:          "Use Proxy",
		StatusTemporaryRedirect: "Temporary Redirect",
		StatusPermanentRedirect: "Permanent Redirect",

		StatusBadRequest:                   "Bad Request",
		StatusUnauthorized:                 "Unauthorized",
		StatusPaymentRequired:              "Payment Required",
		StatusForbidden:                    "Forbidden",
		StatusNotFound:                     "Not Found",
		StatusMethodNotAllowed:             "Method Not Allowed",
		StatusNotAcceptable:                "Not Acceptable",
		StatusProxyAuthRequired:            "Proxy Authentication Required",
		StatusRequestTimeout:               "Request Timeout",
		StatusConflict:                     "Conflict",
		StatusGone:                         "Gone",
		StatusLengthRequired:               "Length Required",
		StatusPreconditionFailed:           "Precondition Failed",
		StatusRequestEntityTooLarge:        "Request Entity Too Large",
		StatusRequestURITooLong:            "Request URI Too Long",
		StatusUnsupportedMediaType:         "Unsupported Media Type",
		StatusRequestedRangeNotSatisfiable: "Requested Range Not Satisfiable",
		StatusExpectationFailed:            "Expectation Failed",
		StatusTeapot:                       "I'm a teapot",
		StatusMisdirectedRequest:           "Misdirected Request",
		StatusUnprocessableEntity:          "Unprocessable Entity",
		StatusLocked:                       "Locked",
		StatusFailedDependency:             "Failed Dependency",
		StatusUpgradeRequired:              "Upgrade Required",
		StatusPreconditionRequired:         "Precondition Required",
		StatusTooManyRequests:              "Too Many Requests",
		StatusRequestHeaderFieldsTooLarge:  "Request Header Fields Too Large",
		StatusUnavailableForLegalReasons:   "Unavailable For Legal Reasons",

		StatusInternalServerError:           "Internal Server Error",
		StatusNotImplemented:                "Not Implemented",
		StatusBadGateway:                    "Bad Gateway",
		StatusServiceUnavailable:            "Service Unavailable",
		StatusGatewayTimeout:                "Gateway Timeout",
		StatusHTTPVersionNotSupported:       "HTTP Version Not Supported",
		StatusVariantAlsoNegotiates:         "Variant Also Negotiates",
		StatusInsufficientStorage:           "Insufficient Storage",
		StatusLoopDetected:                  "Loop Detected",
		StatusNotExtended:                   "Not Extended",
		StatusNetworkAuthenticationRequired: "Network Authentication Required",
	}
)

// StatusText returns a text for the given HTTP status code. If the code is unknown, "Unknown Status Code" is returned.
func StatusText(statusCode int) string {
	if statusCode < statusMessageMin || statusCode > statusMessageMax {
		return unknownStatusCode
	}

	if s := statusMessages[statusCode]; s != "" {
		return s
	}
	return unknownStatusCode
}
