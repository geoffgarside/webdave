package main

import (
	"log"
	"net/http"
	"os"
	"syscall"

	"golang.org/x/net/webdav"
)

func main() {
	var (
		root   = lookupWithDefault("ROOT", "/dav")
		prefix = lookupWithDefault("PREFIX", "")
		user   = lookupWithDefault("USERNAME", "")
		pass   = lookupWithDefault("PASSWORD", "")
	)

	var handler http.Handler = &webdav.Handler{
		Prefix:     prefix,
		FileSystem: webdav.Dir(root),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			log.Printf("webdave: %v %v (%v) (%#v): %v",
				r.Method, r.URL.Path, r.ContentLength, r.Header, err)
		},
	}

	if user != "" && pass != "" {
		handler = authHandler(handler, user, pass)
	}

	log.Printf("starting webdave server on :5000 [uid=%v/gid=%v]",
		syscall.Getuid(), syscall.Getgid())

	if err := http.ListenAndServe(":5000", handler); err != http.ErrServerClosed {
		log.Println(err)
	}
}

func lookupWithDefault(key, defaultValue string) string {
	s, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return s
}

func authHandler(handler http.Handler, user, pass string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok {
			log.Printf("webdave: authentication failed, missing basic authentication")
			w.Header().Set("WWW-Authenticate", "Basic realm=\"WebDAV\", charset=\"UTF-8\"")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if u != user || p != pass {
			log.Print("webdave: authentication failed, invalid username or password")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		log.Print("webdave: authentication succeeded")

		handler.ServeHTTP(w, r)
	})
}
