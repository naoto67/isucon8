all: build
GOBIN := /home/isucon/local/go/bin/go

.PHONY: clean
clean:
	rm -rf torb

deps:
	# ${GOBIN} mod init torb
	${GOBIN} mod vendor

.PHONY: build
build:
	${GOBIN} build -o torb -v ./src/torb/**/*.go
