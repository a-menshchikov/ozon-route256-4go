#!/usr/bin/env sh

protoc -I. \
  -I./third_party \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  --validate_out="lang=go:." --validate_opt=paths=source_relative \
  report.proto

