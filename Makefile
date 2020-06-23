export ROOT:=$(realpath $(dir $(firstword $(MAKEFILE_LIST))))

build: gobindata install

example: build
	$(ROOT)/bin/cardx $(ROOT)/example
bin/go-bindata:
	GOBIN=$(ROOT)/bin go get -u github.com/go-bindata/go-bindata/...

gobindata: bin/go-bindata
	cd $(ROOT) && $(ROOT)/bin/go-bindata -nometadata -o templates.go templates/...

install:
	GOBIN=$(ROOT)/bin go install .

mod: 
	go mod tidy
