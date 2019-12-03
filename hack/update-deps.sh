#!/bin/bash -e

export GO111MODULE=on
go get -u ./...
go mod vendor
go mod tidy
