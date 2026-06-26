all: compile

linux-build:
	@if kubectl get pods --ignore-not-found | grep -q linux-build; then \
		exit 0; \
	else \
		docker build --target linux-build-image -t linux_build:latest .; \
		kind load docker-image linux_build:latest --name my-node; \
		kubectl run -it linux-build --image=linux_build:latest --attach=false	\
			--overrides='{"spec":{"containers":[{"name":"linux-build","image":"linux_build:latest","imagePullPolicy":"IfNotPresent"}]}}'; \
	fi

compile:
	@if kubectl get pods --ignore-not-found | grep -q linux-build; then \
		for file in go.mod go.sum main.go Makefile pong.c vmlinux.h; do \
			kubectl cp ./$$file linux-build:/pong_ebpf/; \
		done; \
		kubectl exec linux-build -- make build; \
		kubectl cp linux-build:pong_ebpf ./pong_ebpf; \
		chmod +x ./pong_ebpf; \
	else \
		$(MAKE) build; \
	fi

build:
	go mod download
	go generate
	GOOS=linux GOARCH=arm64 go build -o pong_ebpf

destroy:
	kubectl delete pod linux-build --ignore-not-found
