FROM golang:1.25.1-alpine AS builder
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go generate ./...
RUN go build -ldflags="-w -s" -o /pastebin .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN mkdir /data && chown appuser:appgroup /data
USER appuser
COPY --from=builder /pastebin /pastebin
COPY --from=builder /app/view/style.css /view/style.css
EXPOSE 3000
VOLUME /data
ENTRYPOINT ["/pastebin", "-db-path=/data/pastebin.db"]
