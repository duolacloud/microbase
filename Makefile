GOPATH:=$(shell go env GOPATH)
PROTO_FLAGS=--go_opt=paths=source_relative --micro_opt=paths=source_relative
PROTO_PATH=$(GOPATH)/src:.
SRC_DIR=$(GOPATH)/src



.PHONY: proto
proto:
	find proto/ -name '*.proto' -exec protoc --proto_path=$(PROTO_PATH) $(PROTO_FLAGS) --micro_out=$(SRC_DIR) --go_out=plugins=grpc:$(SRC_DIR) {} \;
