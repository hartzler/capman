TEST?=$$(glide nv)
NAME = $(shell awk -F\" '/^const Name/ { print $$2 }' main.go)
VERSION = $(shell awk -F\" '/^const Version/ { print $$2 }' main.go)
DEPS = $(shell go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

all: build

build:
	@mkdir -p bin/
	go build -o bin/$(NAME)

deps:
	glide up

test:
	go test $(TEST) $(TESTARGS) -timeout=30s -parallel=4
	go vet $(TEST)

package: test
	$(eval FILES := $(shell ls build))
	@mkdir -p build/tgz
	for f in $(FILES); do \
		(cd $(shell pwd)/build && tar -zcvf tgz/$$f.tar.gz $$f); \
		echo $$f; \
	done

.PHONY: all deps updatedeps build test package
