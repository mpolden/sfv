all: test

fmt:
	gofmt -w=true *.go

hack:
	@mkdir -p src/github.com/martinp
	@ln -sfn $(CURDIR) src/github.com/martinp/$(NAME)

test:
	go test
