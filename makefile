ifeq ($(OS),Windows_NT)
# Customize for Windows
CP = copy
RM = del
else
# Customized for Linux
CP = cp
RM = rm -rf
endif

BIN_FILE=middlebaby
PROTO_FILES=$(shell find . -name *.proto)

.PHONY: buildandrun
buildandrun:
	@go build -o "${BIN_FILE}" main.go
	./"${BIN_FILE}"
	${RM} "${BIN_FILE}"

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
  		--proto_path=$(shell dirname $(shell pwd)) \
  		--grpc-gateway_out=. \
  		--go_out=paths=source_relative:. \
  		--go-grpc_out=paths=source_relative:. \
  		--go-errors_out=paths=source_relative:. $$file)  \
done;
endef

