.PHONY: all deps test validate vet lint fmt

all: deps test validate

deps:
	go get -t ./...
	go get github.com/golang/lint/golint

test:
	go test -timeout 10s -v -race -cover ./...

validate: vet lint fmt

vet:
	go vet ./...

lint:
	out="$$(golint ./...)"; \
	if [ -n "$$(golint ./...)" ]; then \
		echo "$$out"; \
		exit 1; \
	fi

fmt:
	test -z "$(gofmt -s -l . | tee /dev/stderr)"
