all: compile

compile:
	go generate
	GOOS=linux GOARCH=arm64 go build -o pong_ebpf
