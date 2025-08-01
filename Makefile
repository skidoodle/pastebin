dev:
	@go tool templ generate
	@go run . -addr=":3000" -max-size=32768

gen:
	@go tool templ generate

test:
	@go test -v ./...

run:
	@go build .
	@./pastebin -addr=0.0.0.0:8080 -max-size=32768
