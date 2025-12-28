package main

import (
	"log"
	"os"

	"github.com/emersion/go-smtp"
	"github.com/goyal-aman/mailmux/src/backend"
	"github.com/goyal-aman/mailmux/src/types"

	"gopkg.in/yaml.v3"
)

func main() {
	// 1. Load Config
	var cfg types.Config
	data, _ := os.ReadFile("config.yaml")
	yaml.Unmarshal(data, &cfg)

	// 2. Setup SMTP Server
	// be := &backend.Backend{}
	be := backend.NewBackend(cfg)
	s := smtp.NewServer(be)
	s.Addr = ":1020"
	s.Domain = "localhost"

	// This is the critical line you need:
	s.AllowInsecureAuth = true // Allow auth over non-TLS (important for local testing)

	log.Println("Starting SMTP Proxy on :1020")
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
