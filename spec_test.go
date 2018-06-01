// This is a high level test file to ensure that the library conforms to the specifications listed
// here: https://tools.ietf.org/id/draft-snell-http-batch-00.html
package batch

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var validBody = `
--batch
Content-Type: application/http;version=1.1
Content-Transfer-Encoding: binary
Content-ID: df536860-34f9-11de-b418-0800200c9a66
x-use-https: true

GET / HTTP/1.1
Host: www.google.com
Content-Type: text/plain
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.12; rv:60.0) Gecko/20100101 Firefox/60.0
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Language: en-US,en;q=0.5


--batch--
`

var invalidContentTypeBody = `
--batch
Content-Type: application/json
Content-Transfer-Encoding: binary
Content-ID: df536860-34f9-11de-b418-0800200c9a66

GET / HTTP/1.1
Host: www.google.com
Content-Type: text/plain
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.12; rv:60.0) Gecko/20100101 Firefox/60.0
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Language: en-US,en;q=0.5


--batch--
`

var contentTypeTests = []struct {
	contentType string
	body        string
	statusCode  int
}{
	{
		contentType: `multipart/batch; type="application/http"; boundary=batch`,
		body:        validBody,
		statusCode:  http.StatusOK,
	},
	{
		contentType: "multipart/upload",
		body:        validBody,
		statusCode:  http.StatusBadRequest,
	},
	{
		contentType: `multipart/batch; type="application/json"`,
		body:        validBody,
		statusCode:  http.StatusBadRequest,
	},
	{
		contentType: `multipart/batch; type"application/json"`,
		body:        invalidContentTypeBody,
		statusCode:  http.StatusBadRequest,
	},
}

func TestContentType(t *testing.T) {
	b := &Batch{
		Client: NoOpClient{},
	}

	ts := httptest.NewServer(b)
	defer ts.Close()

	for i, test := range contentTypeTests {
		resp, err := http.Post(
			ts.URL,
			test.contentType,
			strings.NewReader(test.body),
		)
		if err != nil {
			t.Errorf("test %v failed: %v", i, test)
		}

		if resp.StatusCode != test.statusCode {
			t.Errorf("expected HTTP Status code to be %v but was %v", test.statusCode, resp.StatusCode)
			body, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			t.Logf("body: %v\n", string(body))
		}
	}
}
