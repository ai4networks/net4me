FROM golang:1.23 AS builder

WORKDIR /src
COPY go.md .
COPY go.sum .
RUN go mod download

RUN go build -o out/net4me cmd/net4me/...


FROM alpine:latest

COPY --from=builder /src/out/net4me /usr/local/bin

ENTRYPOINT ["net4me"]
