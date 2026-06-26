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

	const null byte = byte(0x00)
	ping := []byte{'p', 'i', 'n', 'g', null}
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
				b, err := reader.ReadBytes(null)
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
