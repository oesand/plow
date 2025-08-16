package specs

type HttpMethod string

// HttpMethod constants represent the standard HTTP methods as defined in RFC 7231 and related specifications.
// These methods are used to indicate the desired action to be performed on a resource identified by a URL.
const (
	HttpMethodGet     HttpMethod = "GET"
	HttpMethodPost    HttpMethod = "POST"
	HttpMethodPut     HttpMethod = "PUT"
	HttpMethodDelete  HttpMethod = "DELETE"
	HttpMethodOptions HttpMethod = "OPTIONS"
	HttpMethodHead    HttpMethod = "HEAD"
	HttpMethodConnect HttpMethod = "CONNECT"
	HttpMethodPatch   HttpMethod = "PATCH"
	HttpMethodTrace   HttpMethod = "TRACE"

	MethodPreface HttpMethod = "PRI"
)

// IsValid checks if the HttpMethod is one of the standard HTTP methods.
func (method HttpMethod) IsValid() bool {
	return method == HttpMethodGet ||
		method == HttpMethodPost ||
		method == HttpMethodPut ||
		method == HttpMethodDelete ||
		method == HttpMethodOptions ||
		method == HttpMethodHead ||
		method == HttpMethodConnect ||
		method == HttpMethodPatch ||
		method == HttpMethodTrace
}

// IsPostable checks if the HttpMethod is suitable for sending a request body.
func (method HttpMethod) IsPostable() bool {
	return method == HttpMethodPost || method == HttpMethodPut ||
		method == HttpMethodDelete || method == HttpMethodPatch
}

// IsReplyable checks if the HttpMethod can have a response.
func (method HttpMethod) IsReplyable() bool {
	return !(method == HttpMethodHead || method == HttpMethodConnect || method == HttpMethodOptions)
}
