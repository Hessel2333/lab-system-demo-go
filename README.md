# 🧪 实验室统一管理系统 (Lab System Manager)

![Vue](https://img.shields.io/badge/Vue.js-35495E?style=flat-square&logo=vuedotjs&logoColor=4FC08D)
![Vite](https://img.shields.io/badge/Vite-B73BFE?style=flat-square&logo=vite&logoColor=FFD62E)
![TailwindCSS](https://img.shields.io/badge/Tailwind_CSS-38B2AC?style=flat-square&logo=tailwind-css&logoColor=white)
![Go](https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white)
![SQLite](https://img.shields.io/badge/SQLite-07405E?style=flat-square&logo=sqlite&logoColor=white)

这是一个基于 **Go (Gin) + Vue 3** 构建的现代化、全生命周期实验室资产与物资管理系统。系统旨在为科研团队提供从申购审批、扫码入库、库区精细化管理（例如专柜双锁逻辑）到报废核销的端到端数字化解决方案。

---

## ✨ 核心特性

- **🚀 现代化的技术栈**：前后端分离架构，由 Vue 3 (Composition API) 驱动的极速响应前端，配合原生 Go 与 GORM 支撑的高并发后端逻辑。
- **📦 免配置持久化**：开箱即用的 SQLite 数据库设计。彻底抛弃 Mock 数据，每一次 UI 操作都会直接反应在物理存储上。
- **🔍 智能物资全链路追踪（试剂模块为例）**：
    - **BPM-A 需求采购闭环**：研发一键提需 → 管控品团队长审批 → 采购员填写易派客订单号并确认闭环，状态机精准驱动流转。
    - **BPM-B 到货资产流**：Excel 大明细上传 → 后端按物资类别自动过滤耗材 → 匹配历史申购单（或直接指派使用人回填紧急单）→ 物理签收生成 UUID 试剂瓶 → 贴码上架入库。
    - **订单号防重拦截**：自动读取易派客 Excel 订单编号并去重，杜绝同一批次反复录入。
    - **精细化余量管控**：支持单瓶试剂的反复领用、消耗追溯与空瓶核销，用量动态进度条（绿/黄/红）实时预警。
    - **合规性管理**：智能识别易制毒/易制爆/剧毒类管控品，强制集中存放管控专柜，领用须经团队长审批与钥匙持有人双签（24h 超时保护）。
- **🔌 模块化架构**：系统包含八大核心板块框架（仪器、实验、原料、试剂、耗材、聚合物、分析、AI 中心）。
- **🤖 AI 赋能**：内置由 Google Gemini API 驱动的智能解析中心（自然语言对话一键生成申购表单；AI 库存建议辅助采购员决策）。

---

## 📸 系统预览

*(此处可放置系统截图或动图)*
- 试剂全库台账视图
- 管控柜智能化选择
- 移动端适配的扫码核销卡片

---

## 🛠️ 快速开始

本项目针对开发人员友好，支持**一键拉取 + 零配置测试（Seed 数据直达）**。

您可以克隆代码后，在 5 分钟内将环境跑起来，并注入含有 40 余条极具真实感的实验品目与流转记录数据。

👉 **详细的启动命令、API 密钥配置与演示数据注入说明，请务必查看：**  
📚 **[DEVELOPER_GUIDE.md](./DEVELOPER_GUIDE.md)**

---

## 📝 迭代日志与逻辑细节

本项目在迭代过程中涉及大量针对实验室特定场景的“微小但致命”的业务逻辑优化（如基于房间号关联的库位对齐容错机制等）。

如果您想了解本系统在体验和业务合规性上做了哪些深度优化，请查阅：  
📜 **[CHANGELOG.md](./CHANGELOG.md)**

---

## 📄 架构与设计指南

有关系统模块拆分、技术选型决策以及内部 API 映射清单，请见：
- [项目架构文档](./项目架构文档.md)
- [交互与功能设计规范](./详细功能设计_更新版.md)

---
*由 AI Agent Driven Development 构建的现代化实验室标杆系统*
