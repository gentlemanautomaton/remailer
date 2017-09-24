FROM golang:latest

EXPOSE 25

WORKDIR /go/src/github.com/dustywilson/remailer
COPY . .
RUN bash cmd/server/build.sh

WORKDIR /config
CMD ["/go/bin/server"]
