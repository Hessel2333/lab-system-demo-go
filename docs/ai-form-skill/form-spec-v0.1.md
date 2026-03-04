# FormSpec v0.1（BPM/OA 表单生成）

## 1. 目标
- 统一将当前仓库中的表单页面归一化为可被 AI 稳定消费的描述结构，覆盖字段、布局、校验、联动、权限、提交流。证据文件路径：`frontend/src/views/ReagentManagement.vue`、`frontend/src/views/Instruments.vue`
- 规范必须优先复用现有组件体系（`ui` + `workflow` + 业务封装组件），不直接抽象出仓库外控件。证据文件路径：`frontend/src/components/ui/Input.vue`、`frontend/src/components/workflow/FlowDetailDialog.vue`

## 2. 适用范围
- 页面入口：`frontend/src/views/*`（共 9 个视图）。证据文件路径：`frontend/src/views`
- 组件来源：`frontend/src/components/*`（共 46 个 Vue 组件，含 ui/workflow/instruments/reagents/organization/common）。证据文件路径：`frontend/src/components`
- 本仓库无集中式 `schema/validator/rules` 目录，校验逻辑主要分散在组件 `submit`/`save` 函数与按钮禁用条件中。证据文件路径：`frontend/src/components/reagents/ReagentRequestWizard.vue`、`frontend/src/components/reagents/ReagentCabinetManager.vue`、`frontend/src/views/UserPermissionSettings.vue`

## 3. FormSpec 顶层结构
```json
{
  "formId": "string",
  "scene": "string",
  "fields": [
    {
      "fieldId": "string",
      "label": "string",
      "component": "string",
      "valueType": "string",
      "required": true,
      "readonly": false,
      "default": null,
      "options": []
    }
  ],
  "layout": {
    "type": "dialog|page|step-page|split-layout|calendar+dialog",
    "grid": "string",
    "steps": []
  },
  "validations": [
    {
      "rule": "string",
      "trigger": "submit|save|change",
      "level": "error|warning",
      "implementation": "string"
    }
  ],
  "linkages": [
    {
      "trigger": "string",
      "target": "string",
      "effect": "string"
    }
  ],
  "permissions": {
    "visibility": "string",
    "readonlyRules": ["string"]
  },
  "submitFlow": {
    "action": "string",
    "endpoint": "string",
    "payload": [],
    "onSuccess": []
  },
  "evidencePaths": ["string"]
}
```

## 4. 字段规范
- `fieldId`：直接使用源码中的 `form.xxx` / `formFields.xxx` / `quickApply.xxx` 路径。证据文件路径：`frontend/src/components/instruments/InstrumentCreationDialog.vue`、`frontend/src/components/reagents/ReagentRequestWizard.vue`、`frontend/src/components/reagents/ReagentDispensePanel.vue`
- `component`：优先填已存在组件名（`Input`/`Switch`/`Dialog` 等）；仅当源码本身使用原生控件时标记 `native-input/native-select/native-textarea`。证据文件路径：`frontend/src/components/ui/Input.vue`、`frontend/src/views/SupplierManagement.vue`
- `valueType`：按实际绑定方式确定（`v-model.number` -> `number`，`Switch` -> `boolean`，`Input type=date` -> `date`）。证据文件路径：`frontend/src/components/reagents/ReagentCatalogManager.vue`、`frontend/src/components/reagents/ReagentRequestWizard.vue`

## 5. 布局规范
- `dialog`：弹窗表单（新增/编辑/审批）。证据文件路径：`frontend/src/components/instruments/StatusUpdateDialog.vue`、`frontend/src/components/organization/UserDialog.vue`
- `step-page/step-dialog`：分阶段流程输入（upload/match/done 或多步骤向导）。证据文件路径：`frontend/src/components/reagents/ProcurementBatchImport.vue`、`frontend/src/components/instruments/ArrivalWizard.vue`
- `split-layout`：左台账 + 右固定申请面板。证据文件路径：`frontend/src/components/reagents/ReagentDispensePanel.vue`
- `calendar+dialog`：时段画布 + 确认弹窗。证据文件路径：`frontend/src/components/instruments/InstrumentReservationCalendar.vue`

## 6. 校验规范
- 优先提取显式逻辑：`if (...) return`、`toast.error`、`alert`、`:disabled`。证据文件路径：`frontend/src/components/reagents/ReagentCabinetManager.vue`、`frontend/src/components/instruments/InstrumentCreationDialog.vue`
- 允许并行记录两类校验：
  - 运行时校验（submit/save 中断）。
  - 交互时校验（按钮禁用、只读字段）。
  证据文件路径：`frontend/src/components/reagents/ReagentDispensePanel.vue`、`frontend/src/components/reagents/ResearcherArrivalList.vue`

## 7. 联动规范
- 使用 `watch/computed` 作为联动的标准表达。证据文件路径：`frontend/src/components/reagents/ReagentRequestWizard.vue`、`frontend/src/components/instruments/InstrumentReservationCalendar.vue`
- 典型联动类型：
  - 远程联动：输入 -> 防抖查询库存 -> 字段自动回填/禁用。证据文件路径：`frontend/src/components/reagents/ReagentRequestWizard.vue`
  - 权限联动：角色/状态 -> 动作按钮集合。证据文件路径：`frontend/src/components/reagents/ReagentDispensePanel.vue`
  - 业务联动：试剂是否管控 -> 可选柜类型过滤。证据文件路径：`frontend/src/components/reagents/ResearcherArrivalList.vue`

## 8. 权限与只读规范
- 当前仓库的 `sessionStore.currentRole/currentUserId` 仅作为 Demo 运行时适配层；生产系统应改为“登录账户上下文 + 权限策略（RBAC/ABAC）”驱动。证据文件路径：`frontend/src/stores/session.ts`、`frontend/src/components/organization/UserPermissionDialog.vue`
- 生产权限模型建议最小字段：`userId`、`departmentId`、`roles[]`、`permissions[]`、`dataScopes[]`，并由后端签发/校验。证据文件路径：`frontend/src/api/organization.ts`
- 只读通常通过三种方式体现：
  - `Input disabled`
  - 纯文本显示
  - 无权限时“仅查看”占位
  证据文件路径：`frontend/src/components/instruments/ArrivalWizard.vue`、`frontend/src/components/reagents/ReagentDispensePanel.vue`

## 9. 提交流规范
- `submitFlow` 至少包含：触发动作、后端接口、成功后 UI 行为（关闭弹窗/刷新列表/切换步骤）。证据文件路径：`frontend/src/components/reagents/ProcurementBatchImport.vue`、`frontend/src/components/reagents/ReagentRequestWizard.vue`
- 若组件通过 `emit` 上抛提交，则记录为“父组件承接接口”。证据文件路径：`frontend/src/components/organization/UserDialog.vue`、`frontend/src/views/UserManagement.vue`

## 10. 兼容约束（V0.1）
- 允许 `native-*` 组件存在（仓库已有历史代码），但新生成页面应优先使用 `ui` 组件。证据文件路径：`frontend/src/views/SupplierManagement.vue`、`frontend/src/components/ui/Input.vue`
- 不引入仓库不存在的表单库（如 `el-form`/`react-hook-form`/`zod` runtime schema）。证据文件路径：`frontend/src`（全局检索未发现对应依赖与目录）
