#!/bin/bash -e

go get -u ./...
go mod vendor
go mod tidy
