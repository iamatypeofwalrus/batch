package batch

import "net/http"

type response struct {
	contentID string
	response  *http.Response
	err       error
}
