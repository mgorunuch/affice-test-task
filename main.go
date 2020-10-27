package main

import (
	"context"
	"github.com/mgorunuch/test-task-affise/router"
	"github.com/mgorunuch/test-task-affise/service"
	"github.com/mgorunuch/test-task-affise/utils"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())

	limiter := utils.NewRateLimiter(100)
	downloader := service.NewUrlDownloader(4, time.Second)

	s := http.Server{
		Addr:    ":8080",
		Handler: router.New(&limiter, downloader),
	}

	s.BaseContext = func(listener net.Listener) context.Context {
		return ctx
	}

	go func() {
		osCall := <-c
		log.Printf("Got system call: %+v", osCall)

		cancel()
		err := s.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("error when trying to serve: %s", err)
	}
}
