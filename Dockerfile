FROM golang:1.24.5-alpine AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go generate ./...
RUN go build -ldflags="-w -s" -o /pastebin .

FROM gcr.io/distroless/static-debian12

COPY --from=builder --chown=nonroot:nonroot /pastebin /pastebin
COPY --from=builder --chown=nonroot:nonroot /app/view/style.css /view/style.css

USER nonroot:nonroot
EXPOSE 3000

ENTRYPOINT ["/pastebin"]
