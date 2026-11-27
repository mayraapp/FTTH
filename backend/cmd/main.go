package main

import (
	"log"
	"net/http"
	"time"

	"ftth/backend/config"
	"ftth/backend/handler"
	"ftth/backend/middleware"
	"ftth/backend/router"
)

func main() {
	log.SetPrefix("ftth ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC)

	appConfig := config.Load()
	api, err := handler.New(appConfig)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	router.SetupRoutes(mux, api)
	mux.Handle("/", router.SPAHandler(http.Dir(appConfig.PublishedWebDir)))
	server := &http.Server{
		Addr:              ":" + appConfig.Port,
		Handler:           middleware.RequestLogger(middleware.SecurityHeaders(mux), log.Default()),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("event=server_started address=%q", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("event=server_stopped error=%q", err)
	}
}
