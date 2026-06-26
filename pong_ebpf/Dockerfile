FROM golang:1.23 as linux-build-image

RUN dpkg --add-architecture i386 && \
    apt-get update && \
    apt-get install -y build-essential \
                       llvm \
                       libelf-dev \
                       libbpf-dev \
                       bpftool

RUN apt-get install -y gcc-multilib
RUN apt-get install -y clang

WORKDIR /pong_ebpf

CMD ["tail", "-f", "/dev/null"]
