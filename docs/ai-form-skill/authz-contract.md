# 登录账户权限契约（AuthZ Contract v0.1）

## 1. 背景与目标
- 当前前端 `sessionStore` 的 `currentRole/currentUserId` 仅用于 Demo 视角切换，不是生产授权来源。证据文件路径：`frontend/src/stores/session.ts`
- 生产系统应由“登录账户上下文 + 后端授权决策”统一控制页面、字段、动作和数据范围。证据文件路径：`backend/internal/authz/policy.go`、`backend/internal/handlers/authz_helper.go`

## 2. 现有模型对齐（仓库真实字段）

### 2.1 用户主数据（数据库）
- `id`
- `username`
- `real_name`
- `department_id`
- `role`
- `is_dispense_key_holder_a`
- `is_dispense_key_holder_b`

证据文件路径：`backend/internal/models/user.go`

### 2.2 鉴权 Actor（后端运行态）
- `user_id`
- `department_id`
- `raw_role`
- `role`（已归一化）

证据文件路径：`backend/internal/authz/policy.go`

### 2.3 后端已支持的资源/动作/范围
- `resource`: `dispense_request` / `procurement_batch` / `reagent_item`
- `action`: `apply` / `leader_approve` / `keyholder_confirm` / `confirm_batch` / `receive` / `checkin`
- `scope`: `global` / `self` / `team` / `assigned`

证据文件路径：`backend/internal/authz/policy.go`

## 3. 生产契约（建议）

### 3.1 登录上下文接口（新增）
`GET /api/auth/context`

```json
{
  "user": {
    "id": 101,
    "username": "zhangsan",
    "real_name": "张三",
    "department_id": 11,
    "department_name": "分析组",
    "role": "team_leader",
    "role_normalized": "leader",
    "is_dispense_key_holder_a": true,
    "is_dispense_key_holder_b": false
  },
  "authz": {
    "roles": ["team_leader"],
    "permissions": [
      "dispense_request:leader_approve:team",
      "reagent_item:checkin:team",
      "reagent_request:read:team"
    ],
    "data_scopes": {
      "department_ids": [11],
      "allow_global_inventory": false
    }
  },
  "session": {
    "tenant_id": "default",
    "issued_at": "2026-03-04T10:30:00+08:00",
    "expires_at": "2026-03-04T18:30:00+08:00"
  }
}
```

说明：
- `role_normalized` 与后端 `NormalizeRole` 保持一致（`team_leader -> leader` 等）。证据文件路径：`backend/internal/authz/policy.go`
- `permissions` 是前端展示优化字段，最终是否允许仍以后端授权为准。证据文件路径：`backend/internal/handlers/authz_helper.go`

### 3.2 前端统一状态契约（替换 sessionStore）
建议新增 `useAuthStore`（或在现有 `sessionStore` 上升级）：

```ts
interface AuthContext {
  user: {
    id: number
    username: string
    real_name: string
    department_id: number
    role: string
    role_normalized: string
    is_dispense_key_holder_a: boolean
    is_dispense_key_holder_b: boolean
  }
  authz: {
    roles: string[]
    permissions: string[]
    data_scopes: {
      department_ids: number[]
      allow_global_inventory: boolean
    }
  }
}
```

前端最小能力函数：
- `hasPermission(code: string): boolean`
- `can(action: string, resource: string, scope?: string): boolean`
- `inScope(deptId: number): boolean`

证据文件路径：`frontend/src/views/ReagentManagement.vue`、`frontend/src/components/reagents/ReagentDispensePanel.vue`

### 3.3 后端动作校验契约（保持权威）
所有敏感写操作必须通过后端 `authorizeAction`：
- 请求中带业务参数（owner/department/assignee/keyHolderA/B）
- 后端构建 `PermissionRequest`
- 返回 403 + 明确 reason（前端仅展示）

证据文件路径：`backend/internal/handlers/authz_helper.go`、`backend/internal/authz/policy.go`

## 4. 页面权限判定标准（生产）

### 4.1 页面可见性
- 依据 `permissions` 决定模块入口可见（例如采购导入、双签审批）。
- 不再用本地切换角色作为可见性依据。

证据文件路径：`frontend/src/views/MainLayout.vue`、`frontend/src/views/ReagentManagement.vue`

### 4.2 字段可编辑性
- 字段编辑权限由 `permissions + data_scopes` 决定（例如双签角色全局策略页可编辑，单用户弹窗只读）。

证据文件路径：`frontend/src/components/organization/UserPermissionDialog.vue`、`frontend/src/views/UserPermissionSettings.vue`

### 4.3 动作可执行性
- 前端先按权限隐藏/禁用按钮；
- 后端再做最终校验（双保险）。

证据文件路径：`frontend/src/components/reagents/ReagentDispensePanel.vue`、`backend/internal/handlers/authz_helper.go`

### 4.4 数据范围过滤
- 列表接口必须支持并应用数据范围（部门/团队/本人）。

证据文件路径：`backend/internal/handlers/organization_handler.go`、`backend/internal/handlers/reagent_handler.go`

## 5. 与现有系统的映射规则

### 5.1 角色归一化（建议复用后端逻辑）
- `admin -> admin`
- `team_leader -> leader`
- `researcher/member -> researcher`
- `procurement/procurement_specialist -> procurement`
- `director -> director`

证据文件路径：`backend/internal/authz/policy.go`

### 5.2 双签角色映射
- `is_dispense_key_holder_a/is_dispense_key_holder_b` 作为显式权限位，不由前端猜测。

证据文件路径：`backend/internal/models/user.go`、`frontend/src/views/UserPermissionSettings.vue`

## 6. 迁移计划（从 Demo 到生产）

1. 增加 `GET /api/auth/context`，前端启动时拉取并缓存。
2. 在前端新增 `auth store`，逐步替换 `sessionStore.currentRole/currentUserId` 读法。
3. 页面层改造顺序建议：
   - `ReagentManagement`（tab 可见性）
   - `ReagentDispensePanel`（动作权限）
   - `ReagentUnifiedInventory`（数据范围）
4. 保留后端 `authorizeAction` 作为最终裁决，不因前端隐藏而省略后端校验。
5. 移除 Demo 角色切换 UI（MainLayout 中的角色切换按钮）。

证据文件路径：`frontend/src/views/ReagentManagement.vue`、`frontend/src/components/reagents/ReagentDispensePanel.vue`、`frontend/src/components/reagents/ReagentUnifiedInventory.vue`、`frontend/src/views/MainLayout.vue`

## 7. 最小验收清单
- 任意用户无法通过改前端变量绕过后端动作授权（返回 403）。
- 同一账号在不同部门范围下，列表数据可见范围正确变化。
- 双签流程仅 A/B 持有人可执行确认动作。
- 页面按钮、字段可编辑态与后端授权结果一致。

证据文件路径：`backend/internal/authz/policy_test.go`、`backend/internal/authz/policy.go`、`frontend/src/components/reagents/ReagentDispensePanel.vue`
