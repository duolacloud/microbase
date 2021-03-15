GOPATH:=$(shell go env GOPATH)
SERVICE_NAME:=search
BIN_NAME:=${SERVICE_NAME}-srv
IMAGE_NAME:=microbase/${BIN_NAME}

all: build

.PHONY: run
run: build
	MICRO_SERVICE_NAME=com.microbase.srv.${SERVICE_NAME} \
	APOLLO_NAMESPACE=application \
	APOLLO_ADDRESS=http://apollo-dev.dev.lucfish.com:8080 \
	APOLLO_APP_ID=microbase-${BIN_NAME} \
	APOLLO_CLUSTER=dev \
	REGISTRY=consul \
	REGISTRY_ADDRESS=localhost:8500 \
	PROMETHEUS_ADDR= \
	./build/${BIN_NAME}
	  
.PHONY: build
build:
	go build -o build/${BIN_NAME} *.go

.PHONY: build-linux
build-linux:
	GOOS="linux" GOARCH="amd64" CGO_ENABLED=0 go build -o build/${BIN_NAME}-linux main.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build -t ${IMAGE_NAME}:latest .

.PHONY: publish-image
publish-image: docker
	docker tag ${IMAGE_NAME}:latest ccr.ccs.tencentyun.com/mofanshow_services/${IMAGE_NAME}:latest
	docker push ccr.ccs.tencentyun.com/mofanshow_services/${IMAGE_NAME}:latest
