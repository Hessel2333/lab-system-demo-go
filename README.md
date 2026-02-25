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
- **🔍 智能物资流转 (以试剂模块为例)**：
    - **全生命周期跟踪**：申购进度、物流到货、扫码入库（入库自动匹配柜位与存放房间）。
    - **精细化余量管控**：支持单瓶试剂的反复领用与消耗追溯。用量动态进度条监控（绿/黄/红预警提示）。
    - **合规性管理**：智能识别“易制毒/易制爆/剧毒”类管控品，入库时强制分发至集中安全柜（如 F311 专柜）并启用双人双锁记录。
- **🔌 模块化架构**：系统包含八大核心板块框架（仪器、实验、原料、试剂、耗材、聚合物、分析、AI 中心）。
- **🤖 AI 赋能**：内置由 Google Gemini API 驱动的智能解析中心（例如通过自然语言对话一键生成复杂化学品的申购表单）。

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
