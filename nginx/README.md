# Nginx 配置部署指南

域名: `elk-helper.slileisure.com`

## 📋 部署步骤

### 1. 获取 SSL 证书（Let's Encrypt）

```bash
# 安装 Certbot
sudo apt update
sudo apt install certbot python3-certbot-nginx -y

# 获取证书
sudo certbot certonly --nginx -d elk-helper.slileisure.com

# 或者使用 standalone 模式（如果 nginx 未运行）
sudo certbot certonly --standalone -d elk-helper.slileisure.com
```

### 2. 部署 Nginx 配置

```bash
# 复制配置文件
sudo cp nginx/simple.conf /etc/nginx/sites-available/elk-helper.conf

# 创建符号链接
sudo ln -s /etc/nginx/sites-available/elk-helper.conf /etc/nginx/sites-enabled/

# 测试配置
sudo nginx -t

# 重载 Nginx
sudo systemctl reload nginx
```

### 3. 启动 Docker 服务

```bash
# 确保在项目根目录
docker compose -f docker-compose-prod.yml up -d

# 检查服务状态
docker compose -f docker-compose-prod.yml ps
```

### 4. 验证部署

```bash
# 测试 HTTPS
curl https://elk-helper.slileisure.com/

# 验证证书
openssl s_client -connect elk-helper.slileisure.com:443 -servername elk-helper.slileisure.com < /dev/null
```

## 🔐 证书自动续期

Let's Encrypt 证书有效期 90 天，需要定期续期：

```bash
# 测试自动续期
sudo certbot renew --dry-run

# Certbot 会自动设置 cron/systemd timer
# 查看续期定时任务
sudo systemctl list-timers | grep certbot
```

## 🛠️ 常用命令

```bash
# 重新加载 Nginx（不中断服务）
sudo nginx -s reload

# 测试配置
sudo nginx -t

# 查看日志
sudo tail -f /var/log/nginx/elk-helper.log
sudo tail -f /var/log/nginx/elk-helper-error.log

# 查看证书信息
sudo certbot certificates
```

## 🌐 DNS 配置

确保域名 `elk-helper.slileisure.com` 的 DNS A 记录指向服务器 IP：

```
elk-helper.slileisure.com  →  A记录  →  服务器IP地址
```

## 📊 架构

```
Internet
    ↓
Nginx (:443)
    ↓
Docker Frontend (:3000)
    ↓
Docker Backend (:8080)
```

## 🔍 故障排查

### 证书错误

```bash
# 检查证书是否存在
sudo ls -l /etc/letsencrypt/live/elk-helper.slileisure.com/

# 查看证书有效期
sudo openssl x509 -in /etc/letsencrypt/live/elk-helper.slileisure.com/fullchain.pem -noout -dates
```

### 502 错误

```bash
# 检查 Docker 容器
docker compose ps

# 测试本地访问
curl http://localhost:3000
```

### DNS 解析问题

```bash
# 检查 DNS 解析
nslookup elk-helper.slileisure.com
dig elk-helper.slileisure.com
```

## 🔒 安全建议

1. ✅ 已启用 HTTPS
2. ✅ 已配置 HTTP 自动重定向
3. ✅ 使用 TLS 1.2/1.3
4. 建议配置防火墙只开放 80/443 端口

```bash
# UFW 防火墙配置
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

