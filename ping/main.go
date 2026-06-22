package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	cl, err := net.Dial("tcp", ":8090")
	if err != nil {
		log.Fatal(err)
	}
	defer cl.Close()

	ping := []byte("ping")
	if _, err = cl.Write(ping); err != nil {
		log.Fatal(err)
	}
	log.Printf("sent: %q\n", ping)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				buf := make([]byte, 1024)
				n, err := cl.Read(buf)
				if err != nil {
					if errors.Is(err, net.ErrClosed) || err == io.EOF {
						stop()
						return
					}
					log.Fatal(err)
				}
				log.Printf("received: %q\n", buf[:n])
			}
		}
	}()
	<-ctx.Done()
}
