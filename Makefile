GOPATH = $(shell go env GOPATH)

make build:
	mkdir -p $(GOPATH)/bin/
	go build -o $(GOPATH)/bin/indexooor main.go
	@echo "Done building indexooor"
