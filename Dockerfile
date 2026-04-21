FROM golang:1.26.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o pastebin .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/pastebin /pastebin
COPY --from=builder /app/view /view
EXPOSE 3000
VOLUME /data
ENTRYPOINT ["/pastebin", "-addr", ":3000", "-db-path", "/data/pastebin.db"]
