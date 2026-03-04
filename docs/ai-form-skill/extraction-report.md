# BPM/OA 表单技能提炼报告（Extraction Report）

## 1. 扫描范围

- 前端主范围：`frontend/src/views`、`frontend/src/components`、`frontend/src/api`、`frontend/src/lib`、`frontend/src/stores`。证据文件路径：`frontend/src/router/index.ts`
- 组件总量：`48` 个 Vue 组件。证据文件路径：`frontend/src/components`
- 视图总量：`9` 个 Vue 视图。证据文件路径：`frontend/src/views`
- 识别到与表单相关（含 `v-model`）的页面/组件文件：`27` 个。证据文件路径：`frontend/src/components`、`frontend/src/views`
- 代码库不存在集中式 `schema/validator/rules/form` 目录，表单规则主要内嵌在组件逻辑。证据文件路径：`frontend/src`（目录检索结果为空）、`frontend/src/components/reagents/ReagentRequestWizard.vue`、`frontend/src/components/reagents/ReagentCabinetManager.vue`

## 2. 发现的控件数量和分类统计

### 2.1 可复用控件入库结果

- 本次登记到 `component-registry.json` 的可复用控件：`30` 个。证据文件路径：`docs/ai-form-skill/component-registry.json`
- 控件分层清晰，包含基础 UI、流程组件、业务表单组件。证据文件路径：`frontend/src/components/ui`、`frontend/src/components/workflow`、`frontend/src/components/reagents`

### 2.2 分类统计（按 registry.category）

| 分类 | 数量 |
|---|---:|
| base-input | 4 |
| base-overlay | 2 |
| base-action | 1 |
| base-form-meta | 1 |
| status-display | 1 |
| layout | 1 |
| workflow-display | 2 |
| workflow-action | 1 |
| workflow-container | 1 |
| business-form-dialog | 5 |
| business-permission-dialog | 1 |
| business-step-form | 1 |
| business-time-form | 2 |
| business-ai-form | 1 |
| business-masterdata-form | 2 |
| business-master-detail-form | 1 |
| business-approval-form | 1 |
| business-checkin-form | 1 |
| business-lifecycle-form | 1 |

证据文件路径：`docs/ai-form-skill/component-registry.json`

### 2.3 关键观察

- 当前项目已经有统一基础组件层（`Input/Textarea/Select/Button/Dialog/Switch/TableSection`），具备复用基础。证据文件路径：`frontend/src/components/ui/Input.vue`、`frontend/src/components/ui/Textarea.vue`、`frontend/src/components/ui/Select.vue`、`frontend/src/components/ui/Button.vue`、`frontend/src/components/ui/Dialog.vue`
- 审批流组件链条完整（状态胶囊 + 时间轴 + 动作面板 + 流转单容器），可直接作为 BPM 表单标准骨架。证据文件路径：`frontend/src/components/workflow/FlowStatusPill.vue`、`frontend/src/components/workflow/FlowTimeline.vue`、`frontend/src/components/workflow/FlowActionPanel.vue`、`frontend/src/components/workflow/FlowDetailDialog.vue`
- 业务表单中仍存在较多裸 HTML 表单元素（52 处），会导致新老页面风格不一致，需要在生成约束中限制新增裸控件。证据文件路径：`frontend/src/views/SupplierManagement.vue`、`frontend/src/components/organization/UserDialog.vue`、`frontend/src/components/reagents/ReagentCabinetManager.vue`

## 3. 示例页面来源（FormSpec 12 个）

1. `frontend/src/components/instruments/InstrumentCreationDialog.vue`
2. `frontend/src/components/instruments/StatusUpdateDialog.vue`
3. `frontend/src/components/instruments/ArrivalWizard.vue`
4. `frontend/src/components/instruments/InstrumentReservationCalendar.vue`
5. `frontend/src/components/organization/UserDialog.vue`
6. `frontend/src/views/UserPermissionSettings.vue`
7. `frontend/src/components/reagents/ReagentRequestWizard.vue`
8. `frontend/src/components/reagents/ReagentCatalogManager.vue`
9. `frontend/src/components/reagents/ReagentCabinetManager.vue`
10. `frontend/src/components/reagents/ProcurementBatchImport.vue`
11. `frontend/src/components/reagents/ReagentDispensePanel.vue`
12. `frontend/src/components/reagents/ResearcherArrivalList.vue`

证据文件路径：`docs/ai-form-skill/form-spec-examples.json`

## 4. 规则提炼结论

- 已形成“字段类型/业务场景 -> 控件”映射，覆盖简单录入、审批流、主子表、动态联动、权限只读。证据文件路径：`docs/ai-form-skill/mapping-rules.md`
- 已形成 V1 生成硬约束，明确“优先复用现有控件、禁止绕过设计体系、统一校验与提交规范”。证据文件路径：`docs/ai-form-skill/generation-constraints.md`
- 已定义统一 FormSpec 结构，可作为后续 Agent 的输入协议。证据文件路径：`docs/ai-form-skill/form-spec-v0.1.md`

## 5. 不确定项与待人工确认项

### 5.1 不确定项

- `Dialog` 组件定义使用 `size`，但部分业务组件仍传 `maxWidth`（未在 `Dialog` props 中声明），依赖 `v-bind="$attrs"` 间接透传，需确认是否为有意兼容。证据文件路径：`frontend/src/components/ui/Dialog.vue`、`frontend/src/components/instruments/InstrumentCreationDialog.vue`
- 一些业务流程仍使用 `alert` 而非 `toast`，需要确认是否作为过渡实现（影响统一交互反馈）。证据文件路径：`frontend/src/components/instruments/InstrumentCreationDialog.vue`、`frontend/src/components/instruments/InstrumentReservationCalendar.vue`
- 已确认策略：`sessionStore` 仅用于 Demo 切换视角；生产系统权限应由登录账户属性与后端授权策略决定（角色+权限点+数据范围）。证据文件路径：`frontend/src/stores/session.ts`、`frontend/src/components/organization/UserPermissionDialog.vue`

### 5.2 待人工确认项

- 是否要继续将 `ReagentCabinetManager`、`ReagentDispensePanel` 等模块中的裸控件改造为 `ui` 体系（`Textarea/Select` 已补齐并在部分页面替换）。
- 是否要引入统一表单校验抽象层（当前完全分散，后续可考虑 composable 方案）。

证据文件路径：`frontend/src/views/SupplierManagement.vue`、`frontend/src/components/organization/UserDialog.vue`、`frontend/src/components/reagents/ReagentCabinetManager.vue`
