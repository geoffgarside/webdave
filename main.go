package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"syscall"

	"golang.org/x/net/webdav"
)

func main() {
	var (
		root   = lookupWithDefault("ROOT", "/dav")
		prefix = lookupWithDefault("PREFIX", "")
		user   = lookupWithDefault("USERNAME", "")
		pass   = lookupWithDefault("PASSWORD", "")
		puid   = lookupWithDefault("PUID", "")
		pgid   = lookupWithDefault("PGID", "")
	)

	if puid != "" && puid != "0" {
		id, err := strconv.Atoi(puid)
		if err != nil {
			log.Fatalf("Invalid PUID: %v", err)
		}
		if err = syscall.Setuid(id); err != nil {
			log.Fatalf("Failed to setuid to %v: %v", id, err)
		}
	}

	if pgid != "" && pgid != "0" {
		id, err := strconv.Atoi(pgid)
		if err != nil {
			log.Fatalf("Invalid PGID: %v", err)
		}
		if err = syscall.Setgid(id); err != nil {
			log.Fatalf("Failed to setgid to %v: %v", id, err)
		}
	}

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

	log.Print("starting webdave server on :5000")

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
