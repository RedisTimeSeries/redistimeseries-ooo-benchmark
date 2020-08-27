# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

.PHONY: all test coverage
all: test coverage build

build:
	$(GOBUILD) .

checkfmt:
	@echo 'Checking gofmt';\
 	bash -c "diff -u <(echo -n) <(gofmt -d .)";\
	EXIT_CODE=$$?;\
	if [ "$$EXIT_CODE"  -ne 0 ]; then \
		echo '$@: Go files must be formatted with gofmt'; \
	fi && \
	exit $$EXIT_CODE

get:
	$(GOGET) -t -v ./...

test: get
	$(GOFMT) ./...
	$(GOTEST) -race -covermode=atomic ./...

coverage: get test
	$(GOTEST) -race -coverprofile=coverage.txt -covermode=atomic .


release:
	$(GOGET) github.com/mitchellh/gox
	$(GOGET) github.com/tcnksm/ghr
	GO111MODULE=on gox  -osarch "linux/amd64 darwin/amd64" -output "dist/redistimeseries-ooo-benchmark_{{.OS}}_{{.Arch}}" .
	aws s3 cp dist/redistimeseries-ooo-benchmark_darwin_amd64 s3://benchmarks.redislabs/redistimeseries/tools/redistimeseries-ooo-benchmark/ --acl=public-read
	aws s3 cp dist/redistimeseries-ooo-benchmark_linux_amd64 s3://benchmarks.redislabs/redistimeseries/tools/redistimeseries-ooo-benchmark/ --acl=public-read

fmt:
	$(GOFMT) ./...

