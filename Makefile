GOPATH:=$(shell go env GOPATH)
PROTO_FLAGS=--go_opt=paths=source_relative --micro_opt=paths=source_relative
PROTO_PATH=.
SRC_DIR=.
MODIFY=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,Mproto/imports/api.proto=github.com/micro/go-micro/v2/api/proto

.PHONY: proto
proto:
	find proto/ -name '*.proto' -exec protoc \
		--proto_path=${PROTO_PATH}:${GOPATH}/src $(PROTO_FLAGS) \
		--micro_out=${MODIFY}:${SRC_DIR} \
		--go_out=plugins=grpc:${SRC_DIR} {} \;

.PHONY: build
build: proto
	for x in $(shell ls ./server); do make -C ./server/$${x}; done;