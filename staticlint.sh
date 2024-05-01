#!/bin/sh
( ( go build -o staticlint cmd/staticlint/*.go ) & ./staticlint ./... ) || echo "failed!"
