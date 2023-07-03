package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pico-db/pico/cli/picod/server"
)

const (
	name   = "picod"
	banner = `
       _               _ 
 _ __ (_) ___ ___   __| |
| '_ \| |/ __/ _ \ / _. |	The lightweight, distributed
| |_) | | (_| (_) | (_| |	document database
| .__/|_|\___\___/ \__,_|	
|_|	github.com/pico-db/pico
`
)

func init() {
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stderr)
	log.SetPrefix(fmt.Sprintf("[%s] ", name))
}

// Entrypoint of the database
func main() {
	defer log.Println("pico server stopped")
	fmt.Print(banner)
	s := server.NewServer()
	graceful(s)
}

// Start the server and handle shutdown on exit signals
func graceful(t server.Server) {
	wait := time.Second * 10
	sigs := make(chan os.Signal, 1)
	ssigs := make(chan error, 1)
	sstop := make(chan bool, 1)
	go func() {
		err := t.Start()
		if err != nil {
			log.Printf("unable to start database server: %s", err.Error())
			ssigs <- err
		}
		sstop <- true
	}()
	signal.Notify(
		sigs,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	select {
	case <-sigs:
		log.Println("received shutdown signal")
		ctx, cancel := context.WithTimeout(context.Background(), wait)
		defer cancel()
		err := t.Stop(ctx)
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("context exceeded")
		}
		if err != nil {
			log.Printf("unable to stop database server: %s", err.Error())
		}
	case err := <-ssigs:
		log.Printf("unable to start database server: %s", err.Error())
	case <-sstop:
		return
	}
}
