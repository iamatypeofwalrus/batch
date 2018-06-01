package batch

import "net/http"

// Response tracks a particular HTTP response and the ContentID of the http request
// from the batch request
type Response struct {
	ContentID string
	Response  *http.Response
	Error     error
}
