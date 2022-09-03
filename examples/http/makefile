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
LEVEL=trace

.PHONY: buildandrun 
buildandrun:
	@go build -o "${BIN_FILE}" main.go
	middlebaby serve --config.file=".middlebaby.yaml" --log.level=$(LEVEL) --target.path="./${BIN_FILE}"
	@${RM} testmb

.PHONY: genpb
genpb:
	@protoc --go_out=./proto \
	--go-grpc_out=require_unimplemented_servers=false:./proto proto/hello.proto


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
ps f `pidtree 9645`
endef

printTree: ; @$(value tree)
.ONESHELL: