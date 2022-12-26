FROM golang:1.18-alpine as builder

WORKDIR /usr/app/src
COPY . ./
RUN go mod download
RUN go build main.go

FROM alpine
WORKDIR /usr/local/bin
COPY --from=builder /usr/app/src/main .
ENTRYPOINT ["main"]

