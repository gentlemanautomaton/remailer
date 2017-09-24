#!/bin/bash

go get -v github.com/dustywilson/remailer github.com/flashmob/go-guerrilla
go install -ldflags "-X main.commitVersion=$(git log | head -n1 | cut -d' ' -f2)" github.com/dustywilson/remailer/cmd/server
