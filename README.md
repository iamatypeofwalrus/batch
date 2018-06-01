[![Build Status](https://codebuild.us-west-2.amazonaws.com/badges?uuid=eyJlbmNyeXB0ZWREYXRhIjoicVhVc05mSW4xeW1vVzdadEk4VG1ldFFzUHFKekNjeUdXOHF6enVianRYNXMrZ01oamxuZHBWTnQvYmp0ay9lQzdzUU5sMkxhWUlqVjZwODRZcVFXeGZNPSIsIml2UGFyYW1ldGVyU3BlYyI6IjkvT3FtOTJGbnd3NEs3a3IiLCJtYXRlcmlhbFNldFNlcmlhbCI6MX0%3D&branch=master)](https://codebuild.us-west-2.amazonaws.com/badges?uuid=eyJlbmNyeXB0ZWREYXRhIjoicVhVc05mSW4xeW1vVzdadEk4VG1ldFFzUHFKekNjeUdXOHF6enVianRYNXMrZ01oamxuZHBWTnQvYmp0ay9lQzdzUU5sMkxhWUlqVjZwODRZcVFXeGZNPSIsIml2UGFyYW1ldGVyU3BlYyI6IjkvT3FtOTJGbnd3NEs3a3IiLCJtYXRlcmlhbFNldFNlcmlhbCI6MX0%3D&branch=master)
[![GoDoc](https://godoc.org/github.com/iamatypeofwalrus/batch?status.svg)](https://godoc.org/github.com/iamatypeofwalrus/batch)
[![Go Report Card](https://goreportcard.com/badge/github.com/iamatypeofwalrus/batch)](https://goreportcard.com/report/github.com/iamatypeofwalrus/batch)
# Batch

Batch is a simple Go library for handling multipart batch requests. It adheres to the [HTTP Multipart Batched Request Format draft spec](https://tools.ietf.org/id/draft-snell-http-batch-00.html).

## Note
The draft spec is not well specified when it comes to whether or not upstream requests use `http` or `https`. The `http.Client` expects either when it makes a request.

By default all requests made by Batch use `http`. However, individual requests can set the `x-use-https` header in the Multipart/Batch message to use `https`.

Here's a complete example of a Batch request where the batch message uses the `x-use-https` header.

```
POST / HTTP/1.1
Host: example.org
Content-Type: multipart/batch; type="application/http;version=1.1" boundary=batch
Mime-Version: 1.0

--batch
Content-Type: application/http;version=1.1
Content-Transfer-Encoding: binary
Content-ID: <df536860-34f9-11de-b418-0800200c9a66@example.org>
x-use-https: true

POST /example/application HTTP/1.1
Host: example.org
Content-Type: text/plain
Content-Length: 3

Foo
--batch--
```

## Examples

### Simple
Batch adheres to the `http.Handler` interface and it can be provided directly to `http.ListenAndServe`.

```go
package main

import (
	"net/http"

	"github.com/iamatypeofwalrus/batch"
)

func main() {
	b := batch.New()
	http.ListenAndServe(":8080", b)
}
```

### Custom
You can customize the HTTP Client that Batch uses to perform a single request by providing it with any struct that adheres to the `batch.HTTPClient` interface. The `http.Client` adheres to this interface.

You can provide Batch with a `batch.Logger` in order to capture any error messages that occur when processing requests. `log.Logger` adheres to this interface.

```go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/iamatypeofwalrus/batch"
)

func main() {
	l := log.New(os.Stdout, "", log.LstdFlags)
	b := &batch.Batch{
		Log:    l,
		Client: http.DefaultClient,
	}
	
	http.HandleFunc("/batch", b.ServeHTTP)

	log.Println("listening on :8080")
	http.ListenAndServe(":8080", nil)
}
```
