FROM golang:1.19-bullseye as builder

WORKDIR /opt/app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go  ./

RUN go build main.go

FROM debian:11-slim

RUN apt-get update && apt-get install ca-certificates openssl
COPY --from=builder /opt/app/main /opt/app/

ENTRYPOINT ["/opt/app/main"]