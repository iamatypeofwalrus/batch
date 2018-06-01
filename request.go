package batch

import "net/http"

// Request is a simple struct used to track a specific request as a part of a larger
// batch request and it's ContentID
type Request struct {
	ContentID string
	Request   *http.Request
}
