#!/bin/bash
# 重启 ELK Helper 服务

cd /root/elk-helper || cd "$(dirname "$0")/.." || exit 1

echo "=== 停止服务 ==="
docker compose -f docker-compose-prod.yml down

echo ""
echo "=== 拉取最新镜像 ==="
docker compose -f docker-compose-prod.yml pull

echo ""
echo "=== 启动服务 ==="
docker compose -f docker-compose-prod.yml up -d

echo ""
echo "=== 等待服务启动（60秒）==="
sleep 60

echo ""
echo "=== 服务状态 ==="
docker compose -f docker-compose-prod.yml ps

echo ""
echo "✓ 重启完成"
echo "访问: https://elk-helper.slileisure.com"

