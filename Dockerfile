FROM golang:latest

EXPOSE 25

WORKDIR /go/src/github.com/dustywilson/remailer
COPY . .

WORKDIR /go/src/github.com/dustywilson/remailer/cmd/server
RUN bash build.sh

WORKDIR /config
CMD ["/go/bin/server"]
