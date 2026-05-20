#!/usr/bin/env bash
go generate ./... && go build ./... || {
    echo 'Failed to build project'
    exit 1
}
