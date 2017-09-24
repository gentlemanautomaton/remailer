#!/bin/bash

env > /tmp/ENV
go get -v github.com/dustywilson/remailer github.com/flashmob/go-guerrilla
go install -ldflags "-X main.commitVersion=$SOURCE_COMMIT" github.com/dustywilson/remailer/cmd/server
