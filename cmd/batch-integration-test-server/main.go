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
