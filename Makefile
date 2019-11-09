# meta info
Name := golang

export GO111MODULE=on

.PHONY: test
test: go test ./...

.PHONY: lint
lint: go vet ./...
	golint -set_exist_status ./...

.PHONY: build
build: go build -o hogehoge


