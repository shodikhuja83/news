FROM golang:alpine

WORKDIR /usr/local/go/src/News/cmd

COPY . /usr/local/go/src/News

RUN go build -o main .

EXPOSE 8080

ENTRYPOINT ["/usr/local/go/src/News/cmd/main"]