BINARY=node
CLIENT=client
REGISTRY ?= localhost:5000
IMAGE_NAME=distributed-kv
TAG ?= latest

.PHONY: build run test docker up down logs clean deploy docker-tag docker-push

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
	docker build -t $(IMAGE_NAME):$(TAG) -f deployments/docker/Dockerfile .

docker-tag: docker
	docker tag $(IMAGE_NAME):$(TAG) $(REGISTRY)/$(IMAGE_NAME):$(TAG)

docker-push: docker-tag
	docker push $(REGISTRY)/$(IMAGE_NAME):$(TAG)

up:
	docker compose -f deployments/docker/docker-compose.yml up --build

down:
	docker compose -f deployments/docker/docker-compose.yml down

logs:
	docker compose -f deployments/docker/docker-compose.yml logs -f -t

clean:
	rm -rf bin

deploy:
	# Replaces ${IMAGE_REGISTRY} in manifests.yaml with the REGISTRY variable
	cat deployments/k8s/manifests.yaml | sed "s|\$${IMAGE_REGISTRY}|$(REGISTRY)|g" | kubectl apply -f -
