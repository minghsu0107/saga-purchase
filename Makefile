# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get

all: test build

test: pretest runtest
build: dep
	$(GOBUILD) -o server -v ./cmd/main.go
build-linux: dep
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o server -v ./cmd/main.go

pretest: mockgen
	$(shell $(GOCMD) env GOPATH)/bin/mockgen -source=repo/auth.go -destination=mock/repo/auth.go -package=mock_repo
	$(shell $(GOCMD) env GOPATH)/bin/mockgen -source=service/purchase/interface.go -destination=mock/service/purchase.go -package=mock_service
	$(shell $(GOCMD) env GOPATH)/bin/mockgen -source=service/result/interface.go -destination=mock/service/result.go -package=mock_service
runtest:
	$(GOTEST) -v ./...
dep: wire
	$(shell $(GOCMD) env GOPATH)/bin/wire ./dep

mockgen:
	GO111MODULE=on $(GOGET) github.com/golang/mock/mockgen@v1.4.4
wire:
	GO111MODULE=on $(GOGET) -u github.com/google/wire/cmd/wire@v0.4.0

clean:
	$(GOCLEAN)
	rm -f server