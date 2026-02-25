# 实验室管理系统 — 开发者快速上手指南

## 技术栈

| 层级 | 技术 |
|---|---|
| **前端** | Vue 3 + TypeScript + Vite + Tailwind CSS |
| **后端** | Go 1.21+ (Gin 框架) |
| **数据库** | SQLite（开箱即用，数据文件：`backend/data/lab_system.db`）|
| **AI 服务** | Google Gemini API（仅试剂申购 AI 解析功能需要）|

---

## 环境依赖

- **Go** >= 1.21
- **Node.js** >= 18

---

## 快速启动

### 1. 配置后端环境变量

```bash
cd backend
cp .env.example .env
# 用编辑器打开 .env，填入你自己的 Gemini API Key
# 如不需要 AI 功能，留空也可正常运行
```

`.env` 文件说明：

```env
GEMINI_API_KEY=your_gemini_api_key_here   # Gemini API Key（可在 Google AI Studio 免费获取）
GEMINI_MODEL_NAME=gemini-2.5-flash        # 使用的模型
HTTP_PROXY=http://127.0.0.1:7890          # 可选：代理（中国大陆环境需要）
HTTPS_PROXY=http://127.0.0.1:7890
```

### 2. 启动后端

```bash
cd backend
go run cmd/server/main.go
# 或编译后运行：
go build -o server cmd/server/main.go && ./server
```

后端监听地址：http://localhost:8080

### 3. 启动前端

```bash
cd frontend
npm install   # 首次需要安装依赖
npm run dev
```

前端访问地址：http://localhost:5173

---

## 数据初始化（Seed Data）

系统首次启动数据库为空，可以通过以下 API 快速注入演示数据：

```bash
# 1. 初始化试剂品目与库存数据（按 admin 用户）
curl -X POST http://localhost:8080/api/debug/seed_reagents

# 2. 初始化各团队的在库试剂（需先执行第1步）
curl -X POST http://localhost:8080/api/debug/seed_team_inventory

# 3. 初始化试剂柜（需先执行第1步，会自动匹配分配柜位）
curl -X POST http://localhost:8080/api/debug/seed_cabinets
```

> **⚠️ 注意**：Seed 命令会先清空旧数据再重新写入，生产环境请勿执行。

---

## 功能模块完成度

| 模块 | 状态 | 说明 |
|---|---|---|
| 试剂管理 | ✅ 完善 | 申购→审批→入库→库存→领用→核销 全链路 |
| 试剂柜管理 | ✅ 完善 | 管控柜集中（F311），普通柜按团队分包 |
| 仪器管理 | ✅ 基本完善 | 采购审批、预约日历、维保记录 |
| 供应商管理 | ✅ 可用 | 基本 CRUD |
| 用户/部门管理 | ⚠️ 演示数据 | 有固定演示账号，无登录态鉴权 |
| 其他模块 | ⚠️ UI 框架 | 实验/原料/聚合物等有 UI，后端数据待完善 |

---

## 项目结构

```
lab-system-demo-go/
├── backend/
│   ├── cmd/server/main.go      # 程序入口
│   ├── internal/
│   │   ├── handlers/           # API 处理器（业务逻辑）
│   │   ├── models/             # 数据模型（GORM）
│   │   └── database/           # 数据库初始化
│   ├── data/lab_system.db      # SQLite 数据库文件
│   └── .env.example            # 环境变量模板
├── frontend/
│   ├── src/
│   │   ├── views/              # 页面级组件
│   │   ├── components/         # 业务组件
│   │   │   ├── reagents/       # 试剂管理组件
│   │   │   ├── instruments/    # 仪器管理组件
│   │   │   └── ui/             # 通用 UI 基础组件
│   │   └── router/             # 路由配置
│   └── vite.config.ts          # 构建配置（API 代理：/api → :8080）
├── docs/                       # 功能文档
└── DEVELOPER_GUIDE.md          # 本文档
```

---

## 常见问题

**Q: 前端请求 API 出现 404 / CORS 错误？**
A: 确认后端已在 `:8080` 正常运行。前端通过 Vite 代理将 `/api/...` 转发到 `:8080`，无需修改任何代码。

**Q: AI 试剂解析功能无效？**
A: 检查 `backend/.env` 中的 `GEMINI_API_KEY` 是否已填写有效的 Google Gemini API Key。在非国内网络环境可删除 `HTTP_PROXY` 配置。

**Q: 如何重置数据库到演示状态？**
A: 删除 `backend/data/lab_system.db` 后重启后端，再依次执行上方的三个 Seed 命令。
