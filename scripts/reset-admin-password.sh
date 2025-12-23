#!/bin/bash
# 重置 admin 用户密码为 admin123
# 使用方法: ./scripts/reset-admin-password.sh

set -e

echo "正在重置 admin 用户密码..."

# 方法1：删除现有用户，让代码自动创建（推荐）
echo "删除现有的 admin 用户..."
docker-compose exec -T postgres psql -U postgres -d elk_helper -c "DELETE FROM users WHERE username = 'admin' AND deleted_at IS NULL;" || {
  echo "错误：无法连接到数据库或执行 SQL"
  exit 1
}

echo "已删除 admin 用户"
echo "正在重启后端服务以自动创建新的 admin 用户..."
docker-compose restart backend

echo ""
echo "✅ 重置完成！"
echo "默认账号密码："
echo "  用户名: admin"
echo "  密码: admin123"
echo ""
echo "等待服务启动后即可使用新密码登录..."

