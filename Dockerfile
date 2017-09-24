FROM golang:latest

EXPOSE 25

RUN env > /tmp/ENV
RUN pwd > /tmp/PWD
RUN git rev-parse HEAD > /tmp/HEAD
WORKDIR /go/src/github.com/dustywilson/remailer
COPY . .
RUN bash cmd/server/build.sh

WORKDIR /config
CMD ["/go/bin/server"]
