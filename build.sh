#!/usr/bin/env bash

go run ./cmd/makedec/makedec.go ./spec.toml -p vm -o ./pkg/vm/spec -t --stringer && \
    go generate ./... && \
    go build ./... || {
    echo 'Failed to build project'
    exit 1
}
