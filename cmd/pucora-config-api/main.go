package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pucora/pucora-configurator/internal/api"
	"github.com/pucora/pucora-configurator/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	var origins []string
	if raw := os.Getenv("ALLOWED_ORIGINS"); raw != "" {
		for _, o := range strings.Split(raw, ",") {
			if t := strings.TrimSpace(o); t != "" {
				origins = append(origins, t)
			}
		}
	}

	storeDir := os.Getenv("CONFIG_STORE_PATH")
	if storeDir == "" {
		storeDir = "./data"
	}
	st, err := store.New(storeDir)
	if err != nil {
		log.Fatal(err)
	}

	srv := api.NewServer(origins, st, os.Getenv("CONFIG_API_KEY"), os.Getenv("PUBLIC_BASE_URL"))
	log.Printf("pucora-config-api listening on :%s (store: %s)", port, storeDir)
	if err := http.ListenAndServe(":"+port, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
