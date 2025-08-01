run:
	@go tool templ generate
	@go run .

gen:
	@go tool templ generate

test:
	@go test -v ./...

build:
	@go build -o bin/main main.go
