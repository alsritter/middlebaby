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

.PHONY: genpb
genpb:
	@protoc --go_out=./proto \
	--go-grpc_out=require_unimplemented_servers=false:./proto proto/*.proto