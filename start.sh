#!/bin/bash
cd "$(dirname "$0")"
echo "Starting quanyu-battery-sync..."
go run ./cmd/main.go -config config.yaml
