all: build
GOBIN := /home/isucon/local/go/bin/go

deps:
	cd src/bench; ${GOBIN} mod vendor
	cd src/cmd; ${GOBIN} mod vendor

.PHONY: build
build:
	cd src/cmd; ${GOBIN} install ./...

.PHONY: build-linux
build-linux:
	cd src/cmd; GOOS=linux GOARCH=amd64 ${GOBIN} install ./...

.PHONY: race
race:
	GOPATH=`pwd`:`pwd`/vendor ${GOBIN} install -race ./src/cmd/...

clean:
	rm -f isucon8q-initial-dataset.sql.gz
