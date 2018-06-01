package batch

import "net/http"

// HTTPClient is any interface that can perform an HTTP request and return a response and an error6t
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}
