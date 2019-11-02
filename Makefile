
ifeq ($(shell uname),Darwin)
GL_VER:=v4.1-core
else
GL_VER:=v2.1
endif

all: wolfengo test

wolfengo: gl
	go build -o bin/wolfengo ./src

gl:
	go run src/gl/generate/generate.go $(GL_VER) > src/gl/gl.go
	gofmt -w src/gl/gl.go

errcheck:
	errcheck ./src

test:
	go test ./src

.PHONY: all wolfengo test errcheck gl
