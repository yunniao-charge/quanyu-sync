@echo off
cd /d %~dp0
echo Starting quanyu-battery-sync...
go run ./cmd/main.go -config config.yaml
pause
