package main

import (
	"log"
	"net/http"
)

func main() {
	r := setupRouter()
	log.Println("svc-trading-api listening on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
