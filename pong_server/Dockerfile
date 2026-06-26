FROM debian:12-slim as pong-server-image

#RUN apt-get update && apt-get install strace

WORKDIR /app
COPY pong_server ./

EXPOSE 8090

CMD ["./pong_server"]
#CMD ["strace", "-f", "-o", "syscalls_dump.txt", "./pong_server"]

FROM debian:12-slim as pong-ebpf-sidecar-image

RUN apt-get update && apt-get install -y procps

WORKDIR /app
COPY --chmod=755 pong_ebpf ./

CMD ["./pong_ebpf"]
