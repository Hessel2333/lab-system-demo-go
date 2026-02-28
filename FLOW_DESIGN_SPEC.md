# 流转单 UI 设计规范 (Workflow Design Spec)

本规范定义了实验室管理系统中“流转详情”类界面的 UI 设计标准，核心目标是实现 **Apple 工业级质感** 与 **极致办公效率** 的统一。

---

## 1. 核心设计原则
- **一屏全览 (Single-Screen Fit)**：内容应通过弹性间距和物理高度扩容，确保 100% 覆盖首屏，严禁出现纵向滚动条。
- **呼吸感平衡**：在紧凑布局中，通过 14px 级字号和合理的 Padding（p-5/p-6）维持视觉韵律。
- **物理深度 (Apple Depth)**：利用阴影和背景材质区分虚实，强化界面层级。

---

## 2. 布局模型 (Grid & Layout)
- **容器结构**：采用 `8:4` 比例的双栏布局。
    - **左侧 (Main)**：执行节点、核心单据信息（悬浮态）。
    - **右侧 (Side)**：附件、备注、操作区（嵌入态）。
- **空间参数**：
    - 全局 Margin: `px-6 py-4`。
    - 模块间距 (Gap): `gap-4`。
    - 弹窗基础高度: `min-h-[550px]`。

---

## 3. 像素级对齐 (Precision Alignment)
- **时间轴基准线**：
    - 节点圆圈 (H-8/W-8) 与右侧卡片标题必须实现 **1.5px 级** 的水平轴心对齐（使用 `mt-[1.5px]` 微调）。
    - 连接线起始点应为 `top-[34px]`，确保几何连接严丝合缝。
- **间距黄金比例**：时间轴节点步进间距固定为 **`pb-4`**。

---

## 4. 材质与视觉特征 (Visual Identity)
- **阴影层级**：
    - 主内容卡片：`shadow-md` + `border-slate-200/40`。
    - 侧边辅助栏：`bg-slate-50/40` + `shadow-none` + `border-slate-200/30`。
- **交互反馈**：卡片需配置 `apple-card-hover` 样式（微升效果）。

---

## 5. 状态反馈与动画 (State & Animation)
- **进行中状态 (In Progress)**：
    - 必须植入 **`pulse-soft`（微光脉冲）** 呼吸效果。
    - 边框加亮并配合蓝色的投影，引导视觉焦点。
- **动画规范**：使用 `cubic-bezier(0.4, 0, 0.2, 1)` 过渡曲线，周期宜设为 `3s` 以保证静谧感。

---

## 6. 标准字号配置
- **标题**：`text-2xl` / `font-bold` / `tracking-tight`。
- **正文**：`text-sm` (14px) / `font-bold`（重要内容）或 `text-slate-500`（次要内容）。
- **标签/小字**：`text-[10px]` / `font-bold` / `uppercase` / `tracking-widest`。

---
*注：修改本规范涉及的代码时，请参考 `FlowDetailDialog.vue` 与 `FlowTimeline.vue` 的实现模型。*
