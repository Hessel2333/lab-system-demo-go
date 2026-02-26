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
| 试剂管理 — BPM-A | ✅ 完善 | 申购→团队长审批→采购员接单→BPM-A 闭环，支持订单凭证上传 |
| 试剂管理 — BPM-B | ✅ 完善 | Excel 导入→智能过滤→申购认领/直接指派→三段式物流收货→入库 |
| 试剂管理 — 领用审批 | ✅ 可用 | 研发申请→团队长审批→管控品双人双锁（24h 超时保护） |
| 试剂柜管理 | ✅ 完善 | 管控柜集中（F311），普通柜按团队分包 |
| 仪器管理 | ✅ 基本完善 | 采购审批、预约日历、维保记录 |
| 供应商管理 | ✅ 可用 | 基本 CRUD |
| 用户/部门管理 | ⚠️ 演示数据 | 有固定演示账号，无登录态鉴权 |
| 其他模块 | ⚠️ UI 框架 | 实验/原料/聚合物等有 UI，后端数据待完善 |

---

## 核心 API 速查表

### 试剂申购 (BPM-A)

| Method | 路径 | 说明 |
|---|---|---|
| GET | `/api/reagents/requests` | 获取申购单列表 |
| POST | `/api/reagents/requests` | 创建申购单 |
| POST | `/api/reagents/requests/:id/leader-approve` | 团队长审批管控品 |
| POST | `/api/reagents/requests/:id/approve` | 采购员「已接单」确认 |

### 采购批次导入 (BPM-B)

| Method | 路径 | 说明 |
|---|---|---|
| GET | `/api/reagents/procurement-batches` | 获取所有批次 |
| POST | `/api/reagents/procurement-batches` | 创建批次（上传 Excel 解析结果）|
| PUT | `/api/reagents/procurement-batches/:id/items/:item_id` | 手动修正明细行匹配关系 |
| POST | `/api/reagents/procurement-batches/:id/confirm` | 确认批次（生成待收货记录）|
| GET | `/api/reagents/pending-receives` | 获取待收货明细清单 |
| POST | `/api/reagents/pending-receives/:itemId/receive` | 物理签收，生成 `ReagentItem` |

### 库存与领用

| Method | 路径 | 说明 |
|---|---|---|
| GET | `/api/reagents/items` | 获取全量 ReagentItem 列表 |
| GET | `/api/reagents/items/:uuid` | 按 UUID 查单瓶 |
| PUT | `/api/reagents/items/:uuid/status` | 更新状态（上架/移库等）|
| PUT | `/api/reagents/items/:uuid/consume` | 记录领用消耗量 |
| GET | `/api/reagents/team-inventory` | 按团队聚合库存视图 |
| POST | `/api/reagents/dispense-requests` | 研发提交领用申请 |
| POST | `/api/reagents/dispense-requests/:id/leader-approve` | 团队长审批 |
| POST | `/api/reagents/dispense-requests/:id/key-holder-confirm` | 钥匙持有人双签确认 |

### 数据字典

| Method | 路径 | 说明 |
|---|---|---|
| GET/POST/PUT/DELETE | `/api/reagents/catalogs` | 试剂品目 CRUD |
| GET/POST/PUT/DELETE | `/api/reagents/cabinets` | 试剂柜 CRUD |
| GET | `/api/reagents/stats` | 大盘统计数据 |
| GET | `/api/reagents/stock-check` | 库存预警与 AI 采购建议 |

---

## 常见问题

**Q: 前端请求 API 出现 404 / CORS 错误？**
A: 确认后端已在 `:8080` 正常运行。前端通过 Vite 代理将 `/api/...` 转发到 `:8080`，无需修改任何代码。

**Q: AI 试剂解析功能无效？**
A: 检查 `backend/.env` 中的 `GEMINI_API_KEY` 是否已填写有效的 Google Gemini API Key。在非国内网络环境可删除 `HTTP_PROXY` 配置。

**Q: 如何重置数据库到演示状态？**
A: 删除 `backend/data/lab_system.db` 后重启后端，再依次执行上方的三个 Seed 命令。

**Q: Excel 采购明细上传后，为什么部分条目被自动「已忽略」？**
A: 系统根据 Excel 第22列「物资类别」和第24列「商品类别」自动识别耗材类目（化玻、劳保等），不属于化工/试剂类的条目会被自动过滤忽略，防止污染试剂台账。这是正常行为。

**Q: 没有申购记录，如何给紧急口头采购的物资入库？**
A: 在 BPM-B 的明细分配界面，下拉框下方分组「直接指派」中选择具体研发人员姓名，系统会自动在后台代填紧急申请单并完成台账确权，无需手动创建申购单。

---

## 试剂链路状态机（2026-02 更新）

### 角色动作闭环

- **采购员**：点验收货 → 生成 `ReagentItem(UUID)` → 批量打印标签 → 移交研发入库  
- **研发人员**：扫码识别 `UUID` → 选择库位/试剂柜 → 执行扫码入库（`已到货` → `在库`）

### ReagentItem 状态迁移

- `已到货` → `在库`（仅允许扫码入库/Check-in）
- `在库` → `已耗尽`（领用消耗归零或手动核销）
- 禁止跨级流转（例如 `已到货` 直接 `已耗尽`）

### 新增/强化 API

- `POST /api/reagents/pending-receives/:itemId/receive`
  - 返回 `created_uuids`（本次赋码生成的 UUID 列表）
- `POST /api/reagents/items/print-labels`
  - 入参：`{ "uuids": ["..."] }`
  - 用途：记录二维码打印动作并返回可打印标签数据
- `POST /api/reagents/items/:uuid/check-in`
  - 入参：`{ "location": "E309-01", "cabinet_id": 12 }`
  - 用途：研发扫码后执行正式入库
- `PUT /api/reagents/items/:uuid/status`
  - 增加状态迁移校验与入库必填库位校验
