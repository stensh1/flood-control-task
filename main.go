package main

import (
	"context"
	"flood-control/pkg/server"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Init loads environment variables from .env file
func init() {
	if err := godotenv.Load("cfg/.env"); err != nil {
		log.Fatal("init(): Cannot read env vars:", err)
	} else {
		log.Println("init(): Env vars have been successfully loaded")
	}
}

func main() {
	// New server object
	var s = server.S{}

	// Channel for tracking interrupts
	var ch = make(chan os.Signal, 1)

	// New context to stop server
	ctx := context.Background()

	// Trace interrupts ctrl+c
	signal.Notify(ch, os.Interrupt, syscall.SIGTSTP)

	// Starting the server
	s.Start()

	// Catching keyboard combination
	interrupt := <-ch
	s.LogInfo.Println("Server is shutting down by:", interrupt)
	if err := s.S.Shutdown(ctx); err != nil {
		s.LogFatal.Fatal("Server shutting down status:", err)
	} else {
		s.Stop()
		s.LogInfo.Println("Server shutting down status: OK")
	}
}
