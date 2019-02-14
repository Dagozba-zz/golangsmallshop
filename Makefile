
CLI_BINARY = cli
SERVER_BINARY = server
VET_REPORT = vet.report
TEST_REPORT = tests.xml
GOARCH = amd64

VERSION?=?

# Symlink into GOPATH
GITHUB_USERNAME=dagozba
BUILD_DIR=${GOPATH}/src/github.com/${GITHUB_USERNAME}/golangsmallshop/cmd/
CURRENT_DIR=$(shell pwd)

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=${VERSION}"

# Build the project
all: link proto fmt test vet linux darwin windows

link:
	BUILD_DIR=${BUILD_DIR}; \
	CURRENT_DIR=${CURRENT_DIR}; \
	CLI_BINARY=${CLI_BINARY}; \
	SERVER_BINARY=${SERVER_BINARY}

proto:
	echo Compiling proto files in the api folder; \
	cd ${CURRENT_DIR}; \
	protoc --go_out=plugins=grpc:internal/generated/ api/v1/*.proto; \
	cd - > /dev/null

linux:
	echo Building Linux binary; \
	cd ${BUILD_DIR}/cli; \
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${CLI_BINARY}-linux-${GOARCH} . ; \
	cd ${BUILD_DIR}/server; \
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${SERVER_BINARY}-linux-${GOARCH} . ; \
	cd - >/dev/null

darwin:
	echo Building MacOS binary; \
	cd ${BUILD_DIR}/cli; \
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ${CLI_BINARY}-darwin-${GOARCH} . ; \
	cd ${BUILD_DIR}/server; \
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ${SERVER_BINARY}-darwin-${GOARCH} . ; \
	cd - >/dev/null

windows:
	echo Building Windows binary; \
	cd ${BUILD_DIR}/cli; \
	GOOS=windows GOARCH=${GOARCH} go build ${LDFLAGS} -o ${CLI_BINARY}-windows-${GOARCH}.exe . ; \
	cd ${BUILD_DIR}/server; \
	GOOS=windows GOARCH=${GOARCH} go build ${LDFLAGS} -o ${SERVER_BINARY}-windows-${GOARCH}.exe . ; \
	cd - >/dev/null

test:
	if ! hash go2xunit 2>/dev/null; then go install github.com/tebeka/go2xunit; fi
	cd ${BUILD_DIR}../; \
	go test -v ./... 2>&1 | ${GOPATH}/bin/go2xunit -output ${TEST_REPORT} ; \
	cd - >/dev/null

vet:
	-cd ${BUILD_DIR}../; \
	go vet ./... > ${VET_REPORT} 2>&1 ; \
	cd - >/dev/null

fmt:
	cd ${BUILD_DIR}../; \
	go fmt $$(go list ./... | grep -v /vendor/) ; \
	cd - >/dev/null

clean:
	-rm -f ${TEST_REPORT}
	-rm -f ${VET_REPORT}

.PHONY: link linux darwin windows test vet fmt clean
