package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
)

func contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("Content-Type", "application/wasm")
		}
		next.ServeHTTP(w, r)
	})
}

var (
	addr      = flag.String("a", "127.0.0.1:8080", "listen address")
	staticDir = flag.String("d", "./static", "static dir")
)

func main() {
	flag.Parse()

	http.Handle("/", contentTypeMiddleware(http.FileServer(http.Dir(*staticDir))))
	log.Println("Listening on:", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
