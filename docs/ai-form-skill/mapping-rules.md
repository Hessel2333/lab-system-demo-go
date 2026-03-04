# 字段/场景到控件映射规则

## 1. 基础字段类型映射

| 字段类型/场景 | 推荐控件 | 备选控件 | 禁用控件 | 选择理由 | 反例 |
|---|---|---|---|---|---|
| 短文本（名称、编号、联系人） | `Input` | `native-input`（仅历史兼容） | 任意自定义裸 `<input>` 样式绕过体系 | `Input` 已统一尺寸、焦点、禁用态 | `SupplierManagement`、`UserDialog` 中裸 `<input>` 导致视觉/交互不统一 |
| 长文本（原因、备注） | `Textarea` | `native-textarea`（仅历史兼容） | 将长文本硬塞单行 `Input` | 已有统一 `Textarea`，可复用禁用态/focus 样式 | `StatusUpdateDialog`、`ReagentDispensePanel` |
| 数值（数量、阈值、评分） | `Input type=number + v-model.number` | `native-input type=number` | 字符串输入后手动 parse（无约束） | `v-model.number` 可直接获得数值语义并支持 min/max | `ReagentCatalogManager`、`ReagentRequestWizard` |
| 日期 | `Input type=date` | `native-input type=date` | 文本手输日期 | 当前仓库日期输入都依赖原生 date 控件 | `ReagentRequestWizard` |
| 布尔开关（是否管控） | `Switch` | 双按钮切换（仅固定枚举时） | 裸 checkbox（除历史页面） | `Switch` 支持统一交互与禁用样式 | `ReagentRequestWizard` 使用 `Switch`，优于裸 checkbox |
| 枚举（状态、类型） | `Select` 或分段按钮 | `native-select`（仅历史兼容） | 文本输入代替枚举 | 已有统一 `Select`，支持 `v-model.number` 和统一样式 | `StatusUpdateDialog`、`ReagentCabinetManager` |

证据文件路径：`frontend/src/components/ui/Input.vue`、`frontend/src/components/ui/Textarea.vue`、`frontend/src/components/ui/Select.vue`、`frontend/src/components/ui/switch/index.vue`、`frontend/src/components/reagents/ReagentRequestWizard.vue`、`frontend/src/components/reagents/ReagentCatalogManager.vue`、`frontend/src/views/SupplierManagement.vue`

## 2. 业务场景映射

| 业务场景 | 推荐控件/模式 | 备选 | 禁用 | 选择理由 | 反例 |
|---|---|---|---|---|---|
| 标准新增/编辑弹窗 | `Dialog + Input/Button/Label` | `ConfirmDialog`（仅确认型） | 页面内散落输入不包裹弹窗容器 | 统一弹窗骨架、页脚动作区和关闭行为 | `InstrumentCreationDialog` 是推荐模式 |
| 多步骤流程录入 | `ArrivalWizard`（step 驱动） | `Dialog + step state` 自行实现 | 单页面塞入全部字段不分阶段 | 减少认知负担，便于校验分层 | `ArrivalWizard` |
| 主子表批量导入/匹配 | `ProcurementBatchImport`（upload/match/done） | 拆分为多个页面 | 一次性提交无匹配校验 | 允许行级状态、批量忽略、确认前统计 | `ProcurementBatchImport` |
| 审批流动作面板 | `FlowDetailDialog + FlowActionPanel + FlowTimeline` | `FlowStatusPill`（仅状态展示） | 自定义临时按钮散落在表格行 | 可将动作、时间轴、元信息统一成“流转单”模型 | `ReagentDispensePanel` |
| 角色权限配置 | `UserPermissionSettings`（全局策略） | `UserPermissionDialog`（单用户仪器权限） | 在单用户弹窗里直接改全局双签 | 全局策略与单用户权限分离，避免误配置 | `UserPermissionDialog` 明确“前往权限策略中心” |
| 到货入库选柜 | `ResearcherArrivalList` 的“只读实验室 + 可选柜”模式 | 独立小弹窗复用该模式 | 手工输入实验室文本绕过柜位 | 柜位决定实验室，减少数据不一致 | `ResearcherArrivalList` |

证据文件路径：`frontend/src/components/instruments/ArrivalWizard.vue`、`frontend/src/components/reagents/ProcurementBatchImport.vue`、`frontend/src/components/reagents/ReagentDispensePanel.vue`、`frontend/src/components/workflow/FlowDetailDialog.vue`、`frontend/src/views/UserPermissionSettings.vue`、`frontend/src/components/reagents/ResearcherArrivalList.vue`

## 3. 禁用规则（生成时必须遵守）

- 禁止在已有 `ui` 控件可用时新增裸 HTML 输入控件（`<input>/<select>/<textarea>`）。证据文件路径：`frontend/src/components/ui/Input.vue`、`frontend/src/views/SupplierManagement.vue`
- 禁止跳过流程组件直接拼接审批按钮；审批/双签必须产出 `FlowDetailDialog` 对应的步骤与动作模型。证据文件路径：`frontend/src/components/workflow/FlowDetailDialog.vue`、`frontend/src/components/reagents/ReagentDispensePanel.vue`
- 禁止用前端硬编码角色字符串直接做权限判定；Demo 可走 `sessionStore`，生产必须切换为登录账户权限上下文（RBAC/ABAC）。证据文件路径：`frontend/src/stores/session.ts`、`frontend/src/components/organization/UserPermissionDialog.vue`

## 4. 选择优先级

1. 若存在业务封装组件（如 `ReagentRequestWizard`、`ProcurementBatchImport`），优先复用业务组件。
2. 若无业务封装但有 `ui` 组件，则组合 `ui` 组件。
3. 仅在仓库确实缺失对应组件时，允许 `native-*` 作为临时补位，并在 PR 中标注“待组件化”。

证据文件路径：`frontend/src/components/reagents/ReagentRequestWizard.vue`、`frontend/src/components/ui/Input.vue`、`frontend/src/components/ui/Textarea.vue`、`frontend/src/components/ui/Select.vue`、`frontend/src/views/SupplierManagement.vue`
