# V1 生成约束（AI 代码生成）

## 1. 控件复用硬约束
- 必须优先使用仓库现有控件：`frontend/src/components/ui/*`、`frontend/src/components/workflow/*`、已存在业务表单组件。证据文件路径：`frontend/src/components/ui`、`frontend/src/components/workflow`
- 当场景已存在业务组件时，不得重复造轮子：
  - 申购录入：`ReagentRequestWizard`
  - 采购导入：`ProcurementBatchImport`
  - 领用审批：`ReagentDispensePanel`
  证据文件路径：`frontend/src/components/reagents/ReagentRequestWizard.vue`、`frontend/src/components/reagents/ProcurementBatchImport.vue`、`frontend/src/components/reagents/ReagentDispensePanel.vue`

## 2. 禁止绕过设计体系
- 不允许绕过设计体系直接写裸 HTML 输入控件，优先使用 `Input`/`Textarea`/`Select`。证据文件路径：`frontend/src/components/ui/Input.vue`、`frontend/src/components/ui/Textarea.vue`、`frontend/src/components/ui/Select.vue`
- 若临时使用 `native-*`（如第三方组件插槽限制场景），必须满足：
  - 在 FormSpec 中标注 `component=native-*`
  - 在代码注释或变更说明中写明“使用 native-* 的约束原因”
  - 样式与 `ui` 视觉基线一致（圆角、边框、focus ring）

## 3. 交互一致性约束
- 统一反馈机制：成功/失败优先使用 `toast` 或 `useActionFeedback`，避免页面内零散 `alert`。证据文件路径：`frontend/src/lib/feedback.ts`、`frontend/src/components/reagents/ReagentCabinetManager.vue`
- 提交按钮必须绑定防重复机制：`loading/saving/isPending/:disabled`。证据文件路径：`frontend/src/components/reagents/ResearcherArrivalList.vue`、`frontend/src/views/UserPermissionSettings.vue`
- 搜索与远程查询默认采用防抖（300ms~600ms）。证据文件路径：`frontend/src/components/reagents/ReagentCatalogManager.vue`、`frontend/src/components/reagents/ReagentRequestWizard.vue`

## 4. 校验约束
- 每个可提交表单至少包含：
  - 1 条“必填”规则
  - 1 条“提交前中断”规则（`if (...) return`）
  - 1 个“按钮禁用条件”
  证据文件路径：`frontend/src/components/instruments/InstrumentCreationDialog.vue`、`frontend/src/components/reagents/ReagentDispensePanel.vue`
- 校验文案统一输出中文业务语义（如“领取量超过当前余量”），禁止技术异常直接透传。证据文件路径：`frontend/src/components/reagents/ReagentDispensePanel.vue`

## 5. 联动约束
- 联动逻辑优先使用 `watch/computed`，禁止在模板内写复杂业务表达式。证据文件路径：`frontend/src/components/reagents/ReagentRequestWizard.vue`、`frontend/src/components/reagents/ResearcherArrivalList.vue`
- 涉及状态机/审批流时，必须产出：
  - `status`
  - `steps[]`
  - `actions[]`
  - `notes[]`
  并优先接入 `FlowDetailDialog`。证据文件路径：`frontend/src/components/reagents/ReagentDispensePanel.vue`、`frontend/src/components/workflow/FlowDetailDialog.vue`

## 6. 权限与只读约束
- Demo 阶段可基于 `sessionStore.currentRole/currentUserId` 或组件 `role/userId props`；生产阶段必须切换为“登录账户权限上下文（RBAC/ABAC）”判定，禁止以本地前端状态作为最终授权依据。证据文件路径：`frontend/src/stores/session.ts`、`frontend/src/components/organization/UserPermissionDialog.vue`
- 生产权限判定至少覆盖：页面可见性、字段可编辑性、动作可执行性、数据范围过滤（部门/团队/个人）。证据文件路径：`frontend/src/components/reagents/ReagentDispensePanel.vue`、`frontend/src/components/reagents/ReagentUnifiedInventory.vue`
- 只读态必须可见化：
  - disabled 输入
  - “仅查看”文本占位
  - 纯展示 Badge/Text
  证据文件路径：`frontend/src/components/reagents/ReagentDispensePanel.vue`、`frontend/src/components/instruments/ArrivalWizard.vue`

## 7. API 与提交流约束
- 不新增与现有流程冲突的 API 语义，优先复用现有端点：
  - `/api/reagents/requests`
  - `/api/reagents/procurement-batches`
  - `/api/reagents/dispense-requests`
  - `/api/reagents/items/:uuid/check-in`
  - `/api/users/:id/reagent-permissions`
  证据文件路径：`frontend/src/components/reagents/ReagentRequestWizard.vue`、`frontend/src/components/reagents/ProcurementBatchImport.vue`、`frontend/src/components/reagents/ReagentDispensePanel.vue`、`frontend/src/components/reagents/ResearcherArrivalList.vue`、`frontend/src/api/organization.ts`

## 8. 输出落地约束（给后续 Agent）
- 生成代码前先产出 FormSpec（遵循 `form-spec-v0.1.md`）。
- 生成代码后必须附：
  - 所用组件清单
  - 校验规则清单
  - 联动规则清单
  - 权限/只读规则清单
- 如无法满足约束，必须明确标注“阻塞项 + 需要人工确认项”。
