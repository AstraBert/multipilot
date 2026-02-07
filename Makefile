BIN="./bin"
BIN_NAME="multipilot"
MAIN_PKG="."
SRC=$(shell find . -name "*.go")

ifeq (, $(shell which golangci-lint))
$(warning "could not find golangci-lint in $(PATH), run: curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh")
endif

.PHONY: fmt lint test install_deps clean

all: fmt lint test build

fmt:
	$(info ******************** checking formatting ********************)
	@test -z $(shell gofmt -l $(SRC)) || (gofmt -d $(SRC); exit 1)

lint:
	$(info ******************** running lint tools ********************)
	golangci-lint run -v

test: install_deps
	$(info ******************** running tests ********************)
	go test -v ./...

install_deps:
	$(info ******************** downloading dependencies ********************)
	go get -v ./...

build: install_deps
	$(info ******************** building project ********************)
	@mkdir -p ${BIN}
	GOARCH=amd64 GOOS=darwin go build -o ${BIN}/${BIN_NAME}-darwin-amd64 ${MAIN_PKG}
	GOARCH=amd64 GOOS=linux go build -o ${BIN}/${BIN_NAME}-linux-amd64 ${MAIN_PKG}
	GOARCH=amd64 GOOS=windows go build -o ${BIN}/${BIN_NAME}-windows-amd64.exe ${MAIN_PKG}
	GOARCH=arm64 GOOS=darwin go build -o ${BIN}/${BIN_NAME}-darwin-arm64 ${MAIN_PKG}
	GOARCH=arm64 GOOS=linux go build -o ${BIN}/${BIN_NAME}-linux-arm64 ${MAIN_PKG}
	GOARCH=arm64 GOOS=windows go build -o ${BIN}/${BIN_NAME}-windows-arm64.exe ${MAIN_PKG}

clean:
	@rm -rf ${BIN}
