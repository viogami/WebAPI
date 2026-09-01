# Viogami's WebAPI

基于 Go 和 Gin 的个人 API 服务，提供 p5r 预告信生成、AI 回复、GitHub 内容跳转代理，以及可选的 C.H 数据服务。

## 部署

项目固定使用 Go 1.26.4。SaaS 平台直接从 GitHub 仓库构建即可，无需 Docker 或 Compose：

```bash
go build -o webapi .
./webapi
```

服务默认监听 `8080`；多数平台会自动注入 `PORT`，该变量优先于 `conf/config.yaml`。

`conf/config.yaml` 已提交且可直接使用，仅保存不敏感的服务和 p5cc 默认值。复制 [`.env.example`](.env.example) 为本地 `.env` 后，将其中的变量导入终端；SaaS 平台则在环境变量面板中设置它们。不要提交真实 `.env` 文件。

| 变量 | 用途 |
| --- | --- |
| `OPENAI_BASE_URL` | `/gpt` 使用的 OpenAI 兼容服务地址 |
| `OPENAI_API_KEY` | `/gpt` 所需的 OpenAI 或兼容服务密钥 |
| `DEEPSEEK_BASE_URL` | `/deepseek` 使用的 DeepSeek 兼容服务地址 |
| `DEEPSEEK_API_KEY` | `/deepseek` 所需的 DeepSeek 密钥 |
| `CH_API_DATABASE_URL` | 启用 C.H 时的 PostgreSQL 连接串 |
| `CH_API_PASSWORD_PEPPER` | 启用 C.H 时的密码 pepper |
| `CH_API_ALLOWED_ORIGIN` | 启用 C.H 时允许的前端来源 |

默认 `ch.enabled: false`。需要 C.H 时，将其改为 `true`，并同时配置数据库 URL 与密码 pepper。

## 路由

- `GET /`：服务首页文本
- `GET /healthz`：健康检查，返回 `{"status":"ok"}`
- `GET /p5cc/:text`、`POST /p5cc`：p5r 风格预告信生成
- `POST /gpt`：OpenAI 回复，表单字段 `message`
- `POST /deepseek`：DeepSeek 锐评回复，表单字段 `message`
- `GET /jump/github/*proxyPath`：允许列表内的 GitHub 内容跳转代理
- `/CH/*`：可选 C.H 数据接口

`/gpt` 与 `/deepseek` 均为公开接口；请在部署平台侧配合限流、访问控制或网关策略管理访问量与成本。
