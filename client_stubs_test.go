package batch

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type NoOpClient struct{}

func (n NoOpClient) Do(req *http.Request) (*http.Response, error) {
	body := "hello, world"
	resp := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
		ContentLength: int64(len(body)),
		Request:       req,
		Header:        make(http.Header, 0),
	}
	return resp, nil
}
