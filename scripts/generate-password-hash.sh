#!/bin/bash
# 生成 bcrypt 密码哈希
# 使用方法: ./scripts/generate-password-hash.sh [password]

cd "$(dirname "$0")/../backend" || exit 1
go run cmd/tools/generate-password-hash.go "$@"

