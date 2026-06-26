.DEFAULT_GOAL := build

.PHONY: clean destroy build linux-build

OBJS := pong_bpf.go pong_bpf.o pong_ebpf

clean:
	rm -f $(OBJS)

destroy:
	$(MAKE) clean
	kubectl delete pod linux-build --ignore-not-found

build:
	go generate
	GOOS=linux GOARCH=arm64 go build -o pong_ebpf

linux-build: SRCS := go.mod go.sum main.go Makefile pong.c vmlinux.h
linux-build:
	@if ! kubectl get pods --ignore-not-found | grep -q linux-build; then \
		docker build --target linux-build-image -t linux_build:latest .; \
		kind load docker-image linux_build:latest --name my-node; \
		\
		kubectl run -it linux-build \
			--image=linux_build:latest \
			--attach=false \
			--overrides='{"spec":{"containers":[{"name":"linux-build","image":"linux_build:latest","imagePullPolicy":"IfNotPresent"}]}}'; \
	fi

	@until kubectl get pods | grep -q "linux-build"; do \
	        sleep 1; \
	done
	kubectl wait --for=condition=Ready pod/linux-build --timeout=120s
	kubectl exec linux-build -- bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h

	@for file in $(SRCS); do \
		kubectl cp ./$$file linux-build:/pong_ebpf/; \
	done

	kubectl exec linux-build -- make

	@for file in $(OBJS); do \
		kubectl cp linux-build:$$file ./$$file; \
	done
