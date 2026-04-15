#!/bin/bash
set -e

# 全裕电池数据同步服务 - 重启脚本
# 用法: bash scripts/restart.sh

APP_NAME="quanyu-sync"

echo "重启 ${APP_NAME}..."
systemctl restart "${APP_NAME}"
sleep 2
systemctl status "${APP_NAME}" --no-pager -l
echo ""
echo "查看日志: journalctl -u ${APP_NAME} -f"
