.PHONY: gen deploy

all: help

gen: ## generate certificates, K8s TLS secret, and webhook configuration
	docker build -t generate-certs:v1 ./tls
	docker run --rm -it -v ${PWD}/tls:/tls -v ${PWD}/deploy:/deploy generate-certs:v1

build: ## build Docker image for image cloner
	docker build -t image-cloner:v1 .

deploy:	## register and deploy webhook in K8s cluster
	kubectl apply -f deploy/image-cloner-tls.yaml
	kubectl apply -f deploy/image-cloner-svc.yaml
	kubectl apply -f deploy/image-cloner-deploy.yaml
	kubectl apply -f deploy/image-cloner-webhook.yaml

test: ## run tests
	go clean -testcache
	go test ./... -v

lint: ## run lint and go mod tidy
	golint ./...
	go mod tidy

help: ## print this help
	@grep --no-filename -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sed 's/:.*##/·/' | sort | column -ts '·' -c 120

