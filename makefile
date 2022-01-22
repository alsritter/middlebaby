ifeq ($(OS),Windows_NT)
# Customize for Windows
CP = copy
RM = del
else
# Customized for Linux
CP = cp
RM = rm -rf
endif

.PHONY: buildandrun install
BIN_FILE=middlebaby

buildandrun:
	@go build -o "${BIN_FILE}" main.go
	./"${BIN_FILE}"
	${RM} "${BIN_FILE}"

install:
	@go install .