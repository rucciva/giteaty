.PHONY: generate build test clean

all: generate build

generate:
	go get -v ./... && go generate -v ./...
    
build:
	for dir in `find ./cmd -name main.go -type f` ; do\
    	go build -v -o "bin/$$( basename $$(dirname $$dir))" "$$(dirname $$dir)" ;\
	done

test: 
	go test --parallel 1 --count 1 -coverpkg ./... -coverprofile  coverage.out  ./... 
    
clean: 
	rm -rf bin