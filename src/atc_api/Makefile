DEPS = \
	   github.com/gorilla/mux

.PHONY: build
build: test
	go build -o atc_api

.PHONY: test
test:
	go test -v

.PHONY: get
get:
	go get $(DEPS)
