
all: wolfengo test

wolfengo:
	go build -o bin/wolfengo ./src

errcheck:
	errcheck ./src

test:
	go test ./src

.PHONY: all wolfengo test errcheck
