FROM golang:1.17-alpine

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN go mod download

RUN go build -o /web ./cmd/main.go

EXPOSE 7575

#ENV HTTP_PORT=7575

CMD ["/web"]