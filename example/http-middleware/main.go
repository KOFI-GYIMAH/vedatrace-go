package main

import (
	"fmt"
	"log"
	"net/http"

	vedatrace "github.com/KOFI-GYIMAH/vedatrace-go"
	vnethttp "github.com/KOFI-GYIMAH/vedatrace-go/middleware/nethttp"
)

func main() {
	logger, err := vedatrace.New(vedatrace.Config{
		APIKey:  "vt_your_api_key_here",
		Service: "http-example",
	})
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Stop()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from vedatrace-go!")
	})
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	})

	handler := vnethttp.Middleware(logger)(mux)

	logger.Info("server listening", vedatrace.LogMetadata{"addr": ":8080"})
	if err := http.ListenAndServe(":8080", handler); err != nil {
		logger.Error("server error", err)
	}
}
