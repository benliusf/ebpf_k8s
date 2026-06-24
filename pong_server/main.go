package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	listener, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	log.Print("started pong server")

	pong := []byte{'p', 'o', 'n', 'g', 0}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if conn, err := listener.Accept(); err != nil {
					if ctx.Err() == nil {
						log.Print(err)
					}
				} else {
					defer conn.Close()
					buf := make([]byte, 1024)
					if n, err := conn.Read(buf); err != nil {
						log.Print(err)
					} else {
						log.Printf("received: %q\n", buf[:n])
						if _, err := conn.Write(pong); err != nil {
							log.Print(err)
						} else {
							log.Printf("sent: %q\n", pong)
						}
					}
				}
			}
		}
	}()
	<-ctx.Done()
	log.Print("stopped pong server")
}
