package batch

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

const (
	multipartBatch = "multipart/batch"
	keyBoundary    = "boundary"
	keyType        = "type"

	headerContentID               = "Content-ID"
	headerContentType             = "Content-Type"
	headerContentTransferEncoding = "Content-Transfer-Encoding"
	headerInReplyTo               = "In-Reply-To"
	headerUseHTTPS                = "x-use-https"

	contentTypeHTTP         = "application/http"
	contentTransferEncoding = "binary"

	schemeHTTP  = "http"
	schemeHTTPS = "https"

	badRequest = "400 Bad Request"

	split = ";"
)

// New returns an initialized Batch struct that uses the http.DefaultClient
func New() *Batch {
	return &Batch{Client: http.DefaultClient}
}

// Batch conforms to the http.Handler interface and can be used natively or with the HandlerFunc
// methods to assign the ServeHTTP method to a particular route.
type Batch struct {
	Log    Logger
	Client HTTPClient
}

// ServeHTTP accepts and performs batch HTTP requests per the following specification:
// 	https://tools.ietf.org/id/draft-snell-http-batch-00.html.
func (b *Batch) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if req.Method != http.MethodPost {
		http.Error(w, "404 route not found", http.StatusNotFound)
		return
	}

	mediaType, params, err := mime.ParseMediaType(req.Header.Get(headerContentType))
	if err != nil {
		msg := fmt.Sprintf("could not parse Content-Type header: %v", err)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if mediaType != multipartBatch {
		msg := fmt.Sprintf(
			"expected Content-Type to be %v but was %v",
			multipartBatch,
			mediaType,
		)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Per the spec:
	//   The type parameter of the Multipart/Batch content type and each part of the Multipart/Batch
	//   request MUST use the application/http Content-Type as specified by Section 19.1 of [RFC2616].
	if applicationType := params[keyType]; !isApplicationHTTP(applicationType) {
		msg := fmt.Sprintf(
			"expected Content-Type param 'type' to be %v but was %v",
			contentTypeHTTP,
			applicationType,
		)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	var boundary string
	if boundary = params[keyBoundary]; boundary == "" {
		http.Error(w, "expected boundary field to be present in Content-Type", http.StatusBadRequest)
		return
	}

	requests, err := parseBatchRequests(boundary, req.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	responses := performHTTPRequests(b.Client, requests)

	var body bytes.Buffer
	err = generateResponseBody(&body, boundary, responses)
	if err != nil {
		http.Error(w, "something went wrong while processing the batch request", http.StatusInternalServerError)
		b.log("encountered an error while processing batch request:", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body.Bytes())
}

func (b *Batch) log(v ...interface{}) {
	if b.Log != nil {
		b.Log.Print(v...)
	}
}

func parseBatchRequests(boundaryKey string, body io.Reader) ([]*Request, error) {
	requests := make([]*Request, 0)
	mr := multipart.NewReader(body, boundaryKey)
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		defer p.Close()

		if err != nil {
			return requests, fmt.Errorf("encountered error while parsing multipart message: %v", err)
		}

		// Per the spec:
		// 	Any message parts included in the multipart package that do not specify a Content-Type
		// 	equivalent to that specified in the type parameter MUST be ignored.
		ct := p.Header.Get(headerContentType)
		if ct == "" {
			continue
		}

		// Per the spec:
		//   The type parameter of the Multipart/Batch content type and each part of the Multipart/Batch
		//   request MUST use the application/http Content-Type as specified by Section 19.1 of [RFC2616].
		if !isApplicationHTTP(ct) {
			err = fmt.Errorf(
				"expected multipart message Content-Type header to be %v but was %v",
				contentTypeHTTP,
				ct,
			)
			return requests, err
		}

		rdr := bufio.NewReader(p)
		httpReq, err := http.ReadRequest(rdr)
		if err != nil {
			return requests, fmt.Errorf("encountered error while parsing multipart request: %v", err)
		}

		id := p.Header.Get(headerContentID)
		if id == "" {
			return requests, fmt.Errorf("expected each request to have a present and unique value in the %v header", headerContentID)
		}

		// Prevents an error like "http: Request.RequestURI can't be set in client requests." that is
		// caused by naively creating an HTTP request from the body of the Part.
		httpReq.RequestURI = ""

		// The spec doesn't say how to handle individual batch requests and whether or not they
		// use HTTP or HTTPs. Introducing a convention that if the Part includes the "x-use-https"
		// in the headers with any value we will use HTTPS.
		var scheme string
		if useHTTPS := p.Header.Get(headerUseHTTPS); useHTTPS != "" {
			scheme = schemeHTTPS
		} else {
			scheme = schemeHTTP
		}

		url := &url.URL{
			Scheme: scheme,
			Host:   httpReq.Host,
			Path:   httpReq.URL.Path,
		}
		httpReq.URL = url

		req := &Request{
			ContentID: id,
			Request:   httpReq,
		}

		requests = append(requests, req)
	}

	if len(requests) == 0 {
		return requests, fmt.Errorf("no batch requests present")
	}

	return requests, nil
}

func performHTTPRequests(doer HTTPClient, requests []*Request) []*Response {
	respChan := make(chan *Response, len(requests))

	for _, req := range requests {
		go func(req *Request, ch chan *Response) {
			httpResp, err := http.DefaultClient.Do(req.Request)
			resp := &Response{
				ContentID: req.ContentID,
				Response:  httpResp,
				Error:     err,
			}
			ch <- resp
		}(req, respChan)
	}

	responses := make([]*Response, 0, len(requests))

	// TODO rather than range for requests wait on a channel until we're done
	for range requests {
		resp := <-respChan
		responses = append(responses, resp)
	}
	close(respChan)

	return responses
}

func generateResponseBody(w io.Writer, boundary string, responses []*Response) error {
	mw := multipart.NewWriter(w)
	mw.SetBoundary(boundary)

	for _, resp := range responses {
		h := make(textproto.MIMEHeader)
		h.Set(headerContentType, contentTypeHTTP)
		h.Set(headerContentTransferEncoding, contentTransferEncoding)
		h.Set(headerInReplyTo, resp.ContentID)

		part, err := mw.CreatePart(h)
		if err != nil {
			return err
		}

		// The error associated with the response here means there was something wrong with the
		// request such that the client couldn't make the request at all. Any application level
		// errors would be surfaced in the HTTP response headers and body.
		//
		// Given that we're assuming that the request itself was bad.
		if resp.Error != nil {
			resp.Response = &http.Response{
				Status:        badRequest,
				StatusCode:    http.StatusBadRequest,
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				Body:          ioutil.NopCloser(bytes.NewBufferString(resp.Error.Error())),
				ContentLength: int64(len(resp.Error.Error())),
				Request:       &http.Request{},
				Header:        make(http.Header, 0),
			}
		}

		err = resp.Response.Write(part)
		if err != nil {
			return err
		}
	}

	return mw.Close()
}

func isApplicationHTTP(val string) bool {
	if val == contentTypeHTTP {
		return true
	}

	vals := strings.Split(val, split)

	if vals[0] == contentTypeHTTP {
		return true
	}

	return false
}
