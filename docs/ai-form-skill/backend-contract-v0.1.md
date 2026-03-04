# Backend Contract v0.1（面向可定制系统）

## 1. 目标与适用范围
- 本规范用于与 `form-spec-v0.1.md`、`authz-contract.md` 配套，确保 AI 生成前端时，后端契约稳定可依赖。
- 适用于 BPM/OA/审批流/台账类系统，支持多组织、多角色、流程状态驱动页面。

证据文件路径：`backend/cmd/server/main.go`、`docs/ai-form-skill/form-spec-v0.1.md`、`docs/ai-form-skill/authz-contract.md`

## 2. 当前后端基线（来自现仓库）
- 路由前缀为 `/api`，资源按模块拆分（instrument/reagent/organization/supplier）。
- 返回风格当前以“成功直接返回数据、失败返回 `{error: ...}`”为主，尚未统一 envelope。
- 已存在资源-动作-范围的授权引擎（`Resource/Action/Scope` + `Evaluate`）。

证据文件路径：`backend/cmd/server/main.go`、`backend/internal/handlers/authz_helper.go`、`backend/internal/authz/policy.go`

## 3. 规范设计原则（MUST/SHOULD）
- MUST：后端是最终授权裁决点，前端权限仅做体验优化。
- MUST：流程状态必须可枚举、可校验、可审计。
- MUST：错误结构可机读（code/message/details/request_id）。
- SHOULD：接口向后兼容，分阶段迁移，不一次性推翻现有 `/api`。
- SHOULD：支持租户/组织/团队数据范围约束。

证据文件路径：`backend/internal/authz/policy.go`、`backend/internal/handlers/reagent_handler.go`

## 4. 统一认证与上下文契约

### 4.1 请求上下文（生产）
- `Authorization: Bearer <token>`（替代 Demo 的本地角色切换）。
- `X-Request-ID: <uuid>`（链路追踪）。
- `X-Tenant-ID: <tenant>`（多租户时启用）。

### 4.2 登录上下文接口
- `GET /api/auth/context`
- 返回 `user + authz + session`，字段建议复用 `authz-contract.md`。

证据文件路径：`docs/ai-form-skill/authz-contract.md`、`backend/internal/authz/policy.go`

## 5. 统一授权契约（RBAC + ABAC）

### 5.1 核心模型
```json
{
  "resource": "dispense_request",
  "action": "leader_approve",
  "scope": "team",
  "owner_id": 12,
  "department_id": 11,
  "assignee_id": 0
}
```

### 5.2 鉴权执行规则
- 前端可通过 `permissions[]` 先隐藏按钮。
- 后端必须在写操作前调用 `authorizeAction`。
- 拒绝时返回 `403`，并包含 `reason/resource/action/scope`。

证据文件路径：`backend/internal/authz/policy.go`、`backend/internal/handlers/authz_helper.go`

## 6. 通用 API 约定

### 6.1 命名与版本
- 推荐新增 `v1` 前缀：`/api/v1/...`。
- 现有 `/api/...` 作为兼容层保留一段时间。

### 6.2 列表查询参数（统一）
- `page`, `page_size`, `sort`, `order`, `q`, `filters`（JSON 或 key-value）。
- 数据范围参数不由前端直接决定，后端根据登录上下文附加。

### 6.3 统一响应 envelope（新接口 MUST）
```json
{
  "code": "OK",
  "message": "success",
  "data": {},
  "meta": {
    "request_id": "xxx",
    "page": 1,
    "page_size": 20,
    "total": 300
  }
}
```

### 6.4 兼容策略
- 老接口可继续返回“裸数据/`{error}`”。
- 新接口与重构接口必须使用 envelope。

证据文件路径：`backend/cmd/server/main.go`、`backend/internal/handlers/organization_handler.go`、`backend/internal/handlers/qualification_handler.go`

## 7. 错误码与错误结构规范

### 7.1 标准错误响应
```json
{
  "code": "AUTHZ_DENIED",
  "message": "permission denied",
  "details": {
    "reason": "team scope mismatch",
    "resource": "dispense_request",
    "action": "leader_approve",
    "scope": "team"
  },
  "request_id": "req-xxxx"
}
```

### 7.2 建议错误码
- `VALIDATION_ERROR` (400)
- `AUTHN_REQUIRED` (401)
- `AUTHZ_DENIED` (403)
- `NOT_FOUND` (404)
- `CONFLICT` (409)
- `IDEMPOTENCY_CONFLICT` (409)
- `INTERNAL_ERROR` (500)

证据文件路径：`backend/internal/handlers/authz_helper.go`、`backend/internal/handlers/reagent_handler.go`

## 8. 流程状态机契约（必须外显）

### 8.1 ReagentRequest（BPM-A）
- `待审批 -> 待采购 -> 已接单`
- 驳回分支：`待审批 -> 已驳回`

### 8.2 ReagentDispenseRequest（双签流程）
- `待审批 -> 待双签 -> 已完成`
- 驳回分支：`待审批|待双签 -> 已驳回`

### 8.3 ReagentItem（库存生命周期）
- `已到货 -> 在库 -> 已耗尽`

### 8.4 合同要求
- 每次状态变更必须写审计日志（操作者、时间、动作、备注）。
- 非法状态跃迁返回 `409 CONFLICT`。

证据文件路径：`backend/internal/handlers/reagent_handler.go`、`backend/internal/models/reagent.go`

## 9. 表单元数据契约（为 AI 生成前端服务）

### 9.1 建议新增接口
- `GET /api/v1/form-definitions/{form_key}`
- `GET /api/v1/dictionaries/{dict_key}`

### 9.2 form-definition 建议结构
```json
{
  "form_key": "reagent_request",
  "version": "1.0.0",
  "fields": [
    {
      "id": "quantity",
      "label": "申购数量",
      "type": "number",
      "required": true,
      "rules": [{ "type": "min", "value": 1 }],
      "component_hint": "Input"
    }
  ],
  "linkages": [
    {
      "when": "cas_number changed",
      "then": "fetch stock-check"
    }
  ]
}
```

### 9.3 要求
- 表单字段/校验/联动规则应尽量元数据化，减少前端硬编码。
- 字典项必须包含 `label/value/disabled/order`。

证据文件路径：`frontend/src/components/reagents/ReagentRequestWizard.vue`、`backend/internal/handlers/reagent_handler.go`

## 10. 幂等、并发与一致性
- 写接口 SHOULD 支持 `Idempotency-Key`（创建单据、批量确认、审批动作）。
- 状态更新 SHOULD 支持乐观锁（`version`/`updated_at` 比较）。
- 批量动作 MUST 可回放追踪（批次号、操作人、执行摘要）。

证据文件路径：`backend/internal/handlers/reagent_handler.go`（批次确认/审批类接口）

## 11. 审计与可观测性
- 必须记录：`who/when/what/before/after/request_id`。
- 业务日志与操作日志分离，便于合规审计。
- 关键动作（审批、双签、入库、耗尽）必须可追溯。

证据文件路径：`backend/internal/models/reagent.go`（`ReagentLog`）、`backend/internal/handlers/reagent_handler.go`

## 12. 与当前仓库的迁移路线（建议）
1. 新增 `GET /api/auth/context` 与统一错误码中间件。
2. 新增 `/api/v1` 路由组并优先在新功能使用 envelope。
3. 将审批与状态变更接口优先升级为幂等接口（先加 `Idempotency-Key`）。
4. 增加 `form-definitions`/`dictionaries` 元数据接口，减少前端硬编码规则。
5. 逐步废弃 Demo 角色切换入口，接入真实登录态。

证据文件路径：`backend/cmd/server/main.go`、`frontend/src/views/MainLayout.vue`、`docs/ai-form-skill/authz-contract.md`

## 13. 验收清单（后端）
- 401/403/409/500 错误结构统一且可机读。
- 任一敏感写接口未通过后端鉴权即不可执行。
- 状态跃迁均可被状态机校验，非法跃迁返回冲突。
- 列表接口支持分页和服务端数据范围约束。
- 审批/双签/入库/核销可完整追溯到用户与请求。

证据文件路径：`backend/internal/authz/policy_test.go`、`backend/internal/authz/policy.go`、`backend/internal/handlers/reagent_handler.go`
