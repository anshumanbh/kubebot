package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc/grpclog"
)

func init() {
	grpclog.SetLogger(log.New(ioutil.Discard, "", 0))
}

func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}
