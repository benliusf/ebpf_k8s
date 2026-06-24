package main

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf pong pong.c

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
)

const (
	sys_pidfd_open  = 434
	sys_pidfd_getfd = 438
)

type connectionEvent struct {
	Pid uint32
	Fd  int32
}

type writeEvent struct {
	Fd int32
}

func main() {
	out, err := exec.Command("pgrep", "pong_server").Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			log.Fatal("pong_server not found")
		}
		log.Fatal(err)
	}
	tmp := strings.Fields(string(out))[0]
	serverPid, _ := strconv.Atoi(tmp)

	serverFd, _, errno := syscall.Syscall(sys_pidfd_open, uintptr(serverPid), 0, 0)
	if errno != 0 {
		log.Fatal(fmt.Errorf("failure to get pong_server fd: %v", errno))
	}
	defer syscall.Close(int(serverFd))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal("Removing memlock:", err)
	}

	var objs pongObjects
	if err := loadPongObjects(&objs, nil); err != nil {
		log.Fatal(err)
	}
	defer objs.Close()

	programs := map[string]*ebpf.Program{
		"sys_enter_accept4": objs.HandleSysEnterAccept4,
		"sys_exit_accept4":  objs.HandleSysExitAccept4,
		"sys_enter_read":    objs.HandleSysEnterRead,
		"sys_exit_read":     objs.HandleSysExitRead,
		"sys_enter_write":   objs.HandleSysEnterWrite,
	}
	for name, program := range programs {
		tp, err := link.Tracepoint("syscalls", name, program, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer tp.Close()
	}

	var fds sync.Map

	connectionFds, err := ringbuf.NewReader(objs.ConnectionFds)
	if err != nil {
		log.Fatal(err)
	}
	defer connectionFds.Close()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				record, err := connectionFds.Read()
				if err != nil {
					log.Println(err)
					continue
				}
				var event connectionEvent
				if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
					log.Println(err)
					continue
				}
				fmt.Printf("connection event: pid=%d, fd=%d\n", event.Pid, event.Fd)
				if event.Fd > 0 {
					fds.Store(event.Fd, struct{}{})
				}
			}
		}
	}()

	writeFds, err := ringbuf.NewReader(objs.WriteFds)
	if err != nil {
		log.Fatal(err)
	}
	defer writeFds.Close()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				record, err := writeFds.Read()
				if err != nil {
					log.Println(err)
					continue
				}
				var event writeEvent
				if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
					log.Println(err)
					continue
				}
				fmt.Printf("write event: fd=%d\n", event.Fd)

				if _, ok := fds.Load(event.Fd); ok {
					newFd, _, errno := syscall.Syscall(sys_pidfd_getfd, uintptr(serverFd), uintptr(event.Fd), 0)
					if errno != 0 {
						log.Println(fmt.Errorf("failure to get new fd: %v", errno))
						continue
					}
					if fi := os.NewFile(uintptr(newFd), fmt.Sprintf("%d", newFd)); fi != nil {
						defer fi.Close()
						if _, err := fi.WriteString("pong from ebpf\x00"); err != nil {
							log.Println(err)
						}
					}
				}
			}
		}
	}()

	log.Println("eBPF program attached")
	<-ctx.Done()
	log.Println("received signal, exiting...")
}
