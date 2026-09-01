# C.H backend

C.H 是根 WebAPI 服务可选挂载的 PostgreSQL 数据接口，用于 auto-memories-doll 的用户、历史记录、榜单和同步数据。它作为根项目的本地 Go 模块使用，不再提供单独的 Docker Compose 部署入口。

启用方式：在根目录 `conf/config.yaml` 中设置 `ch.enabled: true`，并在部署环境中设置 `CH_API_DATABASE_URL`、`CH_API_PASSWORD_PEPPER` 与可选的 `CH_API_ALLOWED_ORIGIN`。迁移脚本位于 `migrations/001_init.sql`，应由受管 PostgreSQL 或部署流程执行。

## API 概览

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
