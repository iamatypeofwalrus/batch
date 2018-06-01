package batch

import "net/http"

type request struct {
	contentID string
	request   *http.Request
}
