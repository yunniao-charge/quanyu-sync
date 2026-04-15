#!/bin/bash
set -e

# 全裕电池数据同步服务 - 一键部署脚本
# 用法:
#   bash scripts/deploy.sh          # 更新部署
#   bash scripts/deploy.sh init     # 首次安装

APP_NAME="quanyu-sync"
APP_DIR="/opt/${APP_NAME}"
SERVICE_NAME="${APP_NAME}"
BINARY="quanyu-battery-sync"

if [ "$1" = "init" ]; then
    echo "=== 首次安装 ${APP_NAME} ==="

    # 检查 config.yaml
    if [ ! -f "${APP_DIR}/config.yaml" ]; then
        echo "[1/4] 创建配置文件..."
        cp "${APP_DIR}/config.yaml.example" "${APP_DIR}/config.yaml"
        echo "请编辑 ${APP_DIR}/config.yaml 填入实际配置后再运行 deploy.sh"
        exit 1
    fi

    # 创建日志目录
    mkdir -p "${APP_DIR}/logs"

    # 安装 systemd 服务
    echo "[2/4] 安装 systemd 服务..."
    cp "${APP_DIR}/scripts/${SERVICE_NAME}.service" /etc/systemd/system/
    systemctl daemon-reload
    systemctl enable "${SERVICE_NAME}"

    echo "[3/4] 编译..."
    cd "${APP_DIR}"
    go build -o "${BINARY}" ./cmd/

    echo "[4/4] 启动服务..."
    systemctl start "${SERVICE_NAME}"
    sleep 2
    systemctl status "${SERVICE_NAME}" --no-pager -l

    echo ""
    echo "=== 安装完成 ==="
    echo "查看日志: journalctl -u ${SERVICE_NAME} -f"
    exit 0
fi

echo "=== ${APP_NAME} 更新部署 ==="

cd "${APP_DIR}"

echo "[1/4] 拉取代码..."
git pull

echo "[2/4] 编译..."
go build -o "${BINARY}" ./cmd/

echo "[3/4] 重启服务..."
systemctl restart "${SERVICE_NAME}"

echo "[4/4] 检查状态..."
sleep 2
systemctl status "${SERVICE_NAME}" --no-pager -l

echo ""
echo "=== 部署完成 ==="
echo "查看日志: journalctl -u ${SERVICE_NAME} -f"
