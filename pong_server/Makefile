.DEFAULT_GOAL := build

.PHONY: clean build docker deploy attach

clean:
	rm -f pong_server pong_ebpf
	kubectl delete -f deployment.yaml --ignore-not-found

build:
	GOOS=linux GOARCH=arm64 go build -o pong_server

docker:
	@if [ ! -f "../pong_ebpf/pong_ebpf" ]; then \
		echo "pong_ebpf must be compiled"; \
		exit 1; \
	fi
	cp ../pong_ebpf/pong_ebpf .

	docker build --target pong-server-image -t pong_server:latest .
	kind load docker-image pong_server:latest --name my-node

	docker build --target pong-ebpf-sidecar-image -t pong_ebpf_sidecar:latest .
	kind load docker-image pong_ebpf_sidecar:latest --name my-node

deploy:
	kubectl apply -f deployment.yaml
	$(MAKE) attach

attach:
	@until kubectl get pods | grep -q "pong-server"; do \
		sleep 1; \
	done
	kubectl wait --for=condition=Ready pods -l app=pong-server --timeout=120s
	kubectl port-forward service/pong-service 8090:8090
