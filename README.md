# Cloudflare Solver — 自建打码服务

一键获取 Cloudflare `cf_clearance` + `turnstile_token`。

## 架构

```
客户端 → Go 网关 (:7979) → FlareSolverr v3.5.0 (:8191) → Chrome → 目标站
                              │
                              ├── cf_clearance cookies
                              └── turnstile_token（widget 注入自解）
```

## 快速开始

### API（像 YesCaptcha 一样简单）

```bash
# 一条命令 = cf_clearance + turnstile_token
curl -X POST http://localhost:7979/v1/solve \
  -H "X-API-Key: your-key" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://目标站","sitekey":"站点sitekey"}'

# 查结果
curl http://localhost:7979/v1/result/{task_id} -H "X-API-Key: your-key"
```

### 部署

```bash
# Docker
docker compose up -d

# 裸机
cd solver && PORT=7979 ./solver-linux
# 需要 FlareSolverr v3.5.0 在 :8191 运行
```

## 文件结构

```
solver/                      # Go 网关
├── main.go                  # 入口
├── internal/
│   ├── api/                 # HTTP 路由 + 鉴权 + 限流
│   ├── flaresolverr/        # FlareSolverr 客户端（含 widget 注入）
│   ├── model/               # 数据模型
│   ├── store/               # 内存/Redis 存储
│   ├── ratelimit/           # 限流器
│   └── web/                 # 内嵌 Web UI + OpenAPI 文档
├── DEPLOY.md
└── Dockerfile

turnstile_service_self.py    # Python SDK（供注册机调用）
```

## 许可

MIT
