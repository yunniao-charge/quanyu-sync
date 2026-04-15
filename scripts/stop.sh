#!/bin/bash
set -e

# 全裕电池数据同步服务 - 停止脚本
# 用法: bash scripts/stop.sh

APP_NAME="quanyu-sync"

echo "停止 ${APP_NAME}..."
systemctl stop "${APP_NAME}"
systemctl status "${APP_NAME}" --no-pager -l
