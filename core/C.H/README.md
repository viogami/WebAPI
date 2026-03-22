# C.H backend (Go + PostgreSQL)

auto-memories-doll最小可用后端，名称来源于C.H邮局，用于上传数据并云端保存：

- 用户注册/登录（最简单校验）
- 动漫收藏历史记录上传与查询
- Rank 快照上传与查询
- 批量同步接口

## 1. 启动 PostgreSQL

使用docker，pull 最新pg镜像，这里直接启动：

```bash
cd C.H
docker compose up -d
```

> 首次启动会自动执行 `migrations/001_init.sql` 建表。

## 2. 配置环境变量

```bash
cp .env.example .env
```

默认值即可本地开发。

## 3. 启动后端

```bash
go mod tidy
go run main.go
```

默认监听：`http://localhost:8088`

## 4. API 概览

### 认证

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/me` (Bearer token)

### 历史记录

- `POST /api/v1/history` (Bearer token)
- `GET /api/v1/history?limit=50` (Bearer token)

`POST /api/v1/history` body:

```json
{
  "items": [
    {
      "anime_id": 1,
      "name": "Attack on Titan",
      "cover": "https://...",
      "added_at": "2026-03-20T10:00:00Z"
    }
  ]
}
```

### Rank

- `POST /api/v1/rank` (Bearer token)
- `GET /api/v1/rank?limit=20` (Bearer token)
- `GET /api/v1/rank/latest` (Bearer token)

`POST /api/v1/rank` body:

```json
{
  "title": "我的三月榜单",
  "tier_board_name": "Tier Board",
  "grid_board_name": "九宫格",
  "payload": {
    "tiers": [],
    "history": []
  }
}
```

### 批量同步

- `POST /api/v1/sync` (Bearer token)

```json
{
  "history": [],
  "rank": {
    "title": "我的榜单",
    "tier_board_name": "Tier Board",
    "grid_board_name": "九宫格",
    "payload": {}
  }
}
```
