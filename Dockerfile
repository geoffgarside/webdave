# syntax = docker/dockerfile:1.3

FROM golang:alpine AS builder
RUN addgroup -S -g 1000 app && adduser -S -u 1000 -g app app
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags '-s -w -extldflags "-static"' -o /webdave

FROM scratch
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /webdave /webdave
EXPOSE 5000
USER app
ENTRYPOINT ["/webdave"]
