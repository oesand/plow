package specs

type HttpMethod string

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

func (method HttpMethod) IsPostable() bool {
	return method == HttpMethodPost || method == HttpMethodPut ||
		method == HttpMethodDelete || method == HttpMethodPatch
}

func (method HttpMethod) CanHaveResponseBody() bool {
	return !(method == HttpMethodHead || method == HttpMethodConnect || method == HttpMethodOptions)
}
