BINARY=node
CLIENT=client

.PHONY: build run test docker up down logs clean deploy

build:
	go build -o bin/$(BINARY) ./cmd/node
	go build -o bin/$(CLIENT) ./cmd/client

run:
	go run ./cmd/node

client:
	go run ./cmd/client

test:
	go test ./...

docker:
	docker build -t distributed-kv -f deployments/docker/Dockerfile .

up:
	docker compose -f deployments/docker/docker-compose.yml up --build

down:
	docker compose -f deployments/docker/docker-compose.yml down

logs:
	docker compose -f deployments/docker/docker-compose.yml logs -f -t

clean:
	rm -rf bin

deploy:
	kubectl apply -f deployments/k8s/manifests.yaml
