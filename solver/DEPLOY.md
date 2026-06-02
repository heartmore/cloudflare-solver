# Linux 部署指南

## 1. 环境要求

- Go 1.22+ (编译) 或直接使用预编译二进制
- FlareSolverr v3.5.0 (需 Chrome/Chromium)
- Redis (可选，用于多实例)

## 2. 快速部署

### Docker Compose

```bash
docker compose up -d --scale solver=2 --scale flaresolverr=3
```

### 裸机

```bash
# 编译
GOOS=linux GOARCH=amd64 go build -o solver-linux .

# 运行
PORT=7979 FLARESOLVERR_URL=http://localhost:8191 ./solver-linux
```

## 3. 性能调优

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| WORKERS | 2 | 并发 worker 数 |
| PORT | 8080 | 监听端口 |
| FLARESOLVERR_URL | localhost:8191 | FlareSolverr 地址 |
| REDIS_URL | - | Redis 连接串 |
| SOLVER_API_KEY | - | API 鉴权密钥 |
