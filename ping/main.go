package main

import (
	"bufio"
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

	ping := []byte{'p', 'i', 'n', 'g', 0}
	if _, err = cl.Write(ping); err != nil {
		log.Fatal(err)
	}
	log.Printf("sent: %q\n", ping)

	reader := bufio.NewReader(cl)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				b, err := reader.ReadBytes(0)
				if err != nil {
					defer stop()
					if errors.Is(err, net.ErrClosed) || err == io.EOF {
						return
					}
					log.Fatal(err)
				}
				log.Printf("received: %q\n", b)
			}
		}
	}()
	<-ctx.Done()
}
