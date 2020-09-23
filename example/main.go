package main

import (
	"context"
	"log"
	"net/http"

	"github.com/caddyserver/certmagic"
	crdb "github.com/maragudk/certmagic-storage-crdb"
)

func main() {
	// Use the staging CA
	certmagic.DefaultACME.CA = certmagic.LetsEncryptStagingCA

	// These settings are for connecting in development. In production, use the keys and certs.
	s := crdb.New(crdb.Options{
		User:     "certmagic",
		Host:     "localhost",
		Port:     26257,
		Database: "certmagic",
	})
	certmagic.Default.Storage = s
	if err := s.Connect(context.Background()); err != nil {
		log.Fatalln("Error connecting to storage:", err)
	}

	if err := certmagic.HTTPS([]string{"example.com"}, handler()); err != nil {
		log.Fatalln(err)
	}
}

func handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.Path))
	}
}
