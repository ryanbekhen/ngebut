package ngebut

// HTTP methods as registered with IANA.
// See: https://www.iana.org/assignments/http-methods/http-methods.xhtml
const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH"
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
)

// MIME types that are commonly used
const (
	MIMETextXML               = "text/xml"
	MIMETextHTML              = "text/html"
	MIMETextPlain             = "text/plain"
	MIMETextJavaScript        = "text/javascript"
	MIMETextCSS               = "text/css"
	MIMEApplicationXML        = "application/xml"
	MIMEApplicationJSON       = "application/json"
	MIMEApplicationCBOR       = "application/cbor"
	MIMEApplicationForm       = "application/x-www-form-urlencoded"
	MIMEOctetStream           = "application/octet-stream"
	MIMEMultipartForm         = "multipart/form-data"
	MIMEApplicationJavaScript = "application/javascript"

	MIMETextXMLCharsetUTF8               = "text/xml; charset=utf-8"
	MIMETextHTMLCharsetUTF8              = "text/html; charset=utf-8"
	MIMETextPlainCharsetUTF8             = "text/plain; charset=utf-8"
	MIMETextJavaScriptCharsetUTF8        = "text/javascript; charset=utf-8"
	MIMETextCSSCharsetUTF8               = "text/css; charset=utf-8"
	MIMEApplicationXMLCharsetUTF8        = "application/xml; charset=utf-8"
	MIMEApplicationJSONCharsetUTF8       = "application/json; charset=utf-8"
	MIMEApplicationJavaScriptCharsetUTF8 = "application/javascript; charset=utf-8"
)

// Headers were borrowed from valyala/fasthttp
const (
	// Authentication.
	HeaderAuthorization      = "Authorization"
	HeaderProxyAuthenticate  = "Proxy-Authenticate"
	HeaderProxyAuthorization = "Proxy-Authorization"
	HeaderWWWAuthenticate    = "WWW-Authenticate"

	// Caching.
	HeaderAge           = "Age"
	HeaderCacheControl  = "Cache-Control"
	HeaderClearSiteData = "Clear-Site-Data"
	HeaderExpires       = "Expires"
	HeaderPragma        = "Pragma"
	HeaderWarning       = "Warning"

	// Client hints.
	HeaderAcceptCH         = "Accept-CH"
	HeaderAcceptCHLifetime = "Accept-CH-Lifetime"
	HeaderContentDPR       = "Content-DPR"
	HeaderDPR              = "DPR"
	HeaderEarlyData        = "Early-Data"
	HeaderSaveData         = "Save-Data"
	HeaderViewportWidth    = "Viewport-Width"
	HeaderWidth            = "Width"

	// Conditionals.
	HeaderETag              = "ETag"
	HeaderIfMatch           = "If-Match"
	HeaderIfModifiedSince   = "If-Modified-Since"
	HeaderIfNoneMatch       = "If-None-Match"
	HeaderIfUnmodifiedSince = "If-Unmodified-Since"
	HeaderLastModified      = "Last-Modified"
	HeaderVary              = "Vary"

	// Connection management.
	HeaderConnection      = "Connection"
	HeaderKeepAlive       = "Keep-Alive"
	HeaderProxyConnection = "Proxy-Connection"

	// Content negotiation.
	HeaderAccept         = "Accept"
	HeaderAcceptCharset  = "Accept-Charset"
	HeaderAcceptEncoding = "Accept-Encoding"
	HeaderAcceptLanguage = "Accept-Language"

	// Controls.
	HeaderCookie      = "Cookie"
	HeaderExpect      = "Expect"
	HeaderMaxForwards = "Max-Forwards"
	HeaderSetCookie   = "Set-Cookie"

	// CORS.
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderOrigin                        = "Origin"
	HeaderTimingAllowOrigin             = "Timing-Allow-Origin"
	HeaderXPermittedCrossDomainPolicies = "X-Permitted-Cross-Domain-Policies"

	// Do Not Track.
	HeaderDNT = "DNT"
	HeaderTk  = "Tk"

	// Downloads.
	HeaderContentDisposition = "Content-Disposition"

	// Message body information.
	HeaderContentEncoding = "Content-Encoding"
	HeaderContentLanguage = "Content-Language"
	HeaderContentLength   = "Content-Length"
	HeaderContentLocation = "Content-Location"
	HeaderContentType     = "Content-Type"

	// Proxies.
	HeaderForwarded       = "Forwarded"
	HeaderVia             = "Via"
	HeaderXForwardedFor   = "X-Forwarded-For"
	HeaderXForwardedHost  = "X-Forwarded-Host"
	HeaderXForwardedProto = "X-Forwarded-Proto"

	// Redirects.
	HeaderLocation = "Location"

	// Request context.
	HeaderFrom           = "From"
	HeaderHost           = "Host"
	HeaderReferer        = "Referer"
	HeaderReferrerPolicy = "Referrer-Policy"
	HeaderUserAgent      = "User-Agent"

	// Response context.
	HeaderAllow  = "Allow"
	HeaderServer = "Server"

	// Range requests.
	HeaderAcceptRanges = "Accept-Ranges"
	HeaderContentRange = "Content-Range"
	HeaderIfRange      = "If-Range"
	HeaderRange        = "Range"

	// Security.
	HeaderContentSecurityPolicy           = "Content-Security-Policy"
	HeaderContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only"
	HeaderCrossOriginResourcePolicy       = "Cross-Origin-Resource-Policy"
	HeaderExpectCT                        = "Expect-CT"
	HeaderFeaturePolicy                   = "Feature-Policy"
	HeaderPublicKeyPins                   = "Public-Key-Pins"
	HeaderPublicKeyPinsReportOnly         = "Public-Key-Pins-Report-Only"
	HeaderStrictTransportSecurity         = "Strict-Transport-Security"
	HeaderUpgradeInsecureRequests         = "Upgrade-Insecure-Requests"
	HeaderXContentTypeOptions             = "X-Content-Type-Options"
	HeaderXDownloadOptions                = "X-Download-Options"
	HeaderXFrameOptions                   = "X-Frame-Options"
	HeaderXPoweredBy                      = "X-Powered-By"
	HeaderXXSSProtection                  = "X-XSS-Protection"

	// Server-sent event.
	HeaderLastEventID = "Last-Event-ID"
	HeaderNEL         = "NEL"
	HeaderPingFrom    = "Ping-From"
	HeaderPingTo      = "Ping-To"
	HeaderReportTo    = "Report-To"

	// Transfer coding.
	HeaderTE               = "TE"
	HeaderTrailer          = "Trailer"
	HeaderTransferEncoding = "Transfer-Encoding"

	// WebSockets.
	HeaderSecWebSocketAccept     = "Sec-WebSocket-Accept"
	HeaderSecWebSocketExtensions = "Sec-WebSocket-Extensions" // #nosec G101
	HeaderSecWebSocketKey        = "Sec-WebSocket-Key"
	HeaderSecWebSocketProtocol   = "Sec-WebSocket-Protocol"
	HeaderSecWebSocketVersion    = "Sec-WebSocket-Version"

	// Other.
	HeaderAcceptPatch         = "Accept-Patch"
	HeaderAcceptPushPolicy    = "Accept-Push-Policy"
	HeaderAcceptSignature     = "Accept-Signature"
	HeaderAltSvc              = "Alt-Svc"
	HeaderDate                = "Date"
	HeaderIndex               = "Index"
	HeaderLargeAllocation     = "Large-Allocation"
	HeaderLink                = "Link"
	HeaderPushPolicy          = "Push-Policy"
	HeaderRetryAfter          = "Retry-After"
	HeaderServerTiming        = "Server-Timing"
	HeaderSignature           = "Signature"
	HeaderSignedHeaders       = "Signed-Headers"
	HeaderSourceMap           = "SourceMap"
	HeaderUpgrade             = "Upgrade"
	HeaderXDNSPrefetchControl = "X-DNS-Prefetch-Control"
	HeaderXPingback           = "X-Pingback"
	HeaderXRequestedWith      = "X-Requested-With"
	HeaderXRobotsTag          = "X-Robots-Tag"
	HeaderXUACompatible       = "X-UA-Compatible"
)

// HTTP status codes as registered with IANA.
// See: https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml
const (
	StatusContinue           = 100 // RFC 7231, 6.2.1
	StatusSwitchingProtocols = 101 // RFC 7231, 6.2.2
	StatusProcessing         = 102 // RFC 2518, 10.1
	StatusEarlyHints         = 103 // RFC 8297

	StatusOK                   = 200 // RFC 7231, 6.3.1
	StatusCreated              = 201 // RFC 7231, 6.3.2
	StatusAccepted             = 202 // RFC 7231, 6.3.3
	StatusNonAuthoritativeInfo = 203 // RFC 7231, 6.3.4
	StatusNoContent            = 204 // RFC 7231, 6.3.5
	StatusResetContent         = 205 // RFC 7231, 6.3.6
	StatusPartialContent       = 206 // RFC 7233, 4.1
	StatusMultiStatus          = 207 // RFC 4918, 11.1
	StatusAlreadyReported      = 208 // RFC 5842, 7.1
	StatusIMUsed               = 226 // RFC 3229, 10.4.1

	StatusMultipleChoices   = 300 // RFC 7231, 6.4.1
	StatusMovedPermanently  = 301 // RFC 7231, 6.4.2
	StatusFound             = 302 // RFC 7231, 6.4.3
	StatusSeeOther          = 303 // RFC 7231, 6.4.4
	StatusNotModified       = 304 // RFC 7232, 4.1
	StatusUseProxy          = 305 // RFC 7231, 6.4.5
	_                       = 306 // RFC 7231, 6.4.6 (Unused)
	StatusTemporaryRedirect = 307 // RFC 7231, 6.4.7
	StatusPermanentRedirect = 308 // RFC 7538, 3

	StatusBadRequest                   = 400 // RFC 7231, 6.5.1
	StatusUnauthorized                 = 401 // RFC 7235, 3.1
	StatusPaymentRequired              = 402 // RFC 7231, 6.5.2
	StatusForbidden                    = 403 // RFC 7231, 6.5.3
	StatusNotFound                     = 404 // RFC 7231, 6.5.4
	StatusMethodNotAllowed             = 405 // RFC 7231, 6.5.5
	StatusNotAcceptable                = 406 // RFC 7231, 6.5.6
	StatusProxyAuthRequired            = 407 // RFC 7235, 3.2
	StatusRequestTimeout               = 408 // RFC 7231, 6.5.7
	StatusConflict                     = 409 // RFC 7231, 6.5.8
	StatusGone                         = 410 // RFC 7231, 6.5.9
	StatusLengthRequired               = 411 // RFC 7231, 6.5.10
	StatusPreconditionFailed           = 412 // RFC 7232, 4.2
	StatusRequestEntityTooLarge        = 413 // RFC 7231, 6.5.11
	StatusRequestURITooLong            = 414 // RFC 7231, 6.5.12
	StatusUnsupportedMediaType         = 415 // RFC 7231, 6.5.13
	StatusRequestedRangeNotSatisfiable = 416 // RFC 7233, 4.4
	StatusExpectationFailed            = 417 // RFC 7231, 6.5.14
	StatusTeapot                       = 418 // RFC 7168, 2.3.3
	StatusMisdirectedRequest           = 421 // RFC 7540, 9.1.2
	StatusUnprocessableEntity          = 422 // RFC 4918, 11.2
	StatusLocked                       = 423 // RFC 4918, 11.3
	StatusFailedDependency             = 424 // RFC 4918, 11.4
	StatusTooEarly                     = 425 // RFC 8470, 5.2.
	StatusUpgradeRequired              = 426 // RFC 7231, 6.5.15
	StatusPreconditionRequired         = 428 // RFC 6585, 3
	StatusTooManyRequests              = 429 // RFC 6585, 4
	StatusRequestHeaderFieldsTooLarge  = 431 // RFC 6585, 5
	StatusUnavailableForLegalReasons   = 451 // RFC 7725, 3

	StatusInternalServerError           = 500 // RFC 7231, 6.6.1
	StatusNotImplemented                = 501 // RFC 7231, 6.6.2
	StatusBadGateway                    = 502 // RFC 7231, 6.6.3
	StatusServiceUnavailable            = 503 // RFC 7231, 6.6.4
	StatusGatewayTimeout                = 504 // RFC 7231, 6.6.5
	StatusHTTPVersionNotSupported       = 505 // RFC 7231, 6.6.6
	StatusVariantAlsoNegotiates         = 506 // RFC 2295, 8.1
	StatusInsufficientStorage           = 507 // RFC 4918, 11.5
	StatusLoopDetected                  = 508 // RFC 5842, 7.2
	StatusNotExtended                   = 510 // RFC 2774, 7
	StatusNetworkAuthenticationRequired = 511 // RFC 6585, 6
)
