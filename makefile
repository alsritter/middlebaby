
.PHONY: install
install:
	sudo cp mitm-proxy.crt /usr/local/share/ca-certificates
	sudo update-ca-certificates
	go install



.PHONY: debug-http
debug-http:
	$(call debug_template, ./examples/http)

.PHONY: debug-grpc
debug-grpc:
	$(call debug_template, examples/grpc)

.PHONY: run-http 
run-http:
	@go build -o middlebaby -gcflags=all="-N -l" main.go
	@cp ./middlebaby ./examples/http/middlebaby
	@cd ./examples/http && go build -o "${BIN_FILE}" main.go
	./middlebaby serve --config.file=".middlebaby.yaml" --log.level=$(LEVEL) --target.path="./${BIN_FILE}"
	@${RM} "${BIN_FILE}"

.PHONY: proto
proto:
	$(call build_proto_files, $(PROTO_FILES))

# print process PID tree
printTree: ; @$(value tree)
.ONESHELL:

ifeq ($(OS),Windows_NT)
# Customize for Windows
CP = copy
RM = del
else
# Customized for Linux
CP = cp
RM = rm -rf
endif

LEVEL=trace
BIN_FILE=testmb

# PROTO_FILES=$(shell find . -name *.proto)
PROTO_FILES=proto/task/task.proto

define debug_template
	@go build -o middlebaby -gcflags=all="-N -l" main.go
	@cp ./middlebaby $(1)/middlebaby
	@cd $(1) && go build -o "${BIN_FILE}" main.go
	dlv --listen=:2345 --headless=true --api-version=2 \
		--accept-multiclient exec \
		./middlebaby serve -- --config.file=".middlebaby.yaml" --log.level=$(LEVEL) --target.path="./${BIN_FILE}"
	@${RM} testmb
endef

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

# print child process tree and important-taskserver.
# reference:  
# https://superuser.com/questions/363169/ps-how-can-i-recursively-get-all-child-process-for-a-given-pid
# https://unix.stackexchange.com/questions/270778/how-to-write-exactly-bash-scripts-into-makefiles
define tree =
#!/bin/bash
pidtree() { 
	echo -n $1 " "
	for _child in $(ps -o pid --no-headers --ppid $1); do
		echo -n $_child `pidtree $_child` " "
	done
}

# PID 
ps f `pidtree 11958`
endef