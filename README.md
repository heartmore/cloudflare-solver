# Cloudflare Solver

自建 Cloudflare 验证绕过服务 — 一键获取 `cf_clearance` + `turnstile_token`。

## 架构

```
Client → Go Gateway (:7979) → FlareSolverr v3.5.0 (:8191) → Chrome → Target Site
                                  │
                                  ├── cf_clearance cookies
                                  └── turnstile_token (auto widget injection)
```

## 快速开始

### API

```bash
# 传 url + sitekey，自动返回 cf_clearance + turnstile_token
curl -X POST http://localhost:7979/v1/solve \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","sitekey":"0x4AAAAAA..."}'

# 查询结果
curl http://localhost:7979/v1/result/{task_id} \
  -H "X-API-Key: your-api-key"
```

### 部署

**Docker**

```bash
docker compose up -d
```

**裸机**

```bash
# 1. 启动 FlareSolverr（需要 Chrome）
cd flaresolverr-v35/src && PORT=8191 python3 flaresolverr.py

# 2. 启动网关
PORT=7979 ./solver-linux
```

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/v1/solve` | 提交打码任务 |
| `GET` | `/v1/result/{id}` | 查询任务结果 |
| `GET` | `/v1/health` | 健康检查 |
| `GET` | `/v1/stats` | 统计数据 |

详细文档见 `API文档.md`。

## 项目结构

```
solver/                       # Go 网关
├── main.go                   # 入口
├── internal/
│   ├── api/                  # HTTP 路由、鉴权、限流
│   ├── flaresolverr/         # FlareSolverr 客户端
│   ├── model/                # 数据模型
│   ├── store/                # 任务存储（内存/Redis）
│   ├── redisq/               # Redis 消息队列
│   ├── ratelimit/            # 令牌桶限流
│   └── web/                  # 内嵌 Web UI + OpenAPI 文档
├── DEPLOY.md                 # 部署指南
├── Dockerfile
└── solver.service            # systemd 配置

turnstile_service_self.py     # Python SDK
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | `8080` | 监听端口 |
| `FLARESOLVERR_URL` | `http://localhost:8191` | FlareSolverr 地址 |
| `REDIS_URL` | — | Redis 连接串（可选，默认内存存储） |
| `WORKERS` | `2` | 并发 worker 数 |
| `SOLVER_API_KEY` | — | API 鉴权密钥 |

## 许可

MIT
