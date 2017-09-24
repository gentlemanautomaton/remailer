FROM golang:latest

EXPOSE 25

WORKDIR /go/src/github.com/dustywilson/remailer/cmd/server
COPY . .

RUN ./build.sh

WORKDIR /config
CMD ["/go/bin/server"]
