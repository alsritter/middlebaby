ifeq ($(OS),Windows_NT)
# Customize for Windows
CP = copy
RM = del
else
# Customized for Linux
CP = cp
RM = rm -rf
endif

BIN_FILE=testmb
DEBUG_DIR=cd ./examples/http/

.PHONY: debug
debug:
	@go build -o middlebaby -gcflags=all="-N -l" main.go
	@cp ./middlebaby ./examples/http/middlebaby
	@${DEBUG_DIR} && go build -o "${BIN_FILE}" main.go
	${DEBUG_DIR} && dlv --listen=:2345 --headless=true --api-version=2 \
		--accept-multiclient exec \
		./middlebaby serve -- --config.file=".middlebaby.yaml" --log.level=$(LEVEL) --target.path="./${BIN_FILE}"
	@${RM} testmb

# PROTO_FILES=$(shell find . -name *.proto)
PROTO_FILES=proto/task/task.proto

# use 'kratos proto add proto/task/task.proto'
# 'kratos proto client proto/task/task.proto'
.PHONY: proto
proto:
	$(call build_proto_files, $(PROTO_FILES))

# dirname: remove the non-directory part of the file name. (the 'pwd' command output)
define build_proto_files
@for file in $(1); do \
( 	echo "---\nbuilding: $$file" && \
 	protoc --proto_path=. \
  		--proto_path=./proto \
  		--grpc-gateway_out=./proto/task \
  		--go_out=paths=source_relative:. \
  		--go-grpc_out=paths=source_relative:. $$file)  \
done;
endef

