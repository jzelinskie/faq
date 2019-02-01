#!/bin/bash
set -e

export GO111MODULE=on
go test -v ./...
