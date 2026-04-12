# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

全裕电池数据同步服务 - Go 语言实现，定时从全裕 API 拉取数据并接收推送回调，写入 MongoDB。

详见 [README.md](README.md)

---

## 开发流程

使用 Skills 驱动的开发流程。详见各 Skill 说明。

**核心原则：**
- **TDD**：测试先行，垂直切片（tracer bullet）
- **模块信任**：通过测试的模块默认可信，当作黑盒处理，遇问题再深入
- **Deep Modules**：小接口隐藏大实现

### 工作流

```
需求类: /write-a-prd → /prd-to-issues → /tdd
问题类: /qa → /triage-issue → /tdd
架构类: /improve-codebase-architecture → /design-an-interface → /request-refactor-plan → /tdd
```

### Issue 类型

| 类型 | 说明 |
|------|------|
| `type: requirement` | 需求/方案/重构计划 |
| `type: bug` | Bug 报告 |
| `type: task` | 执行任务（sub-issue） |

---

## Issue 模板

- [需求/方案 Issue](.claude/templates/issue-requirement.md)
- [执行 Issue](.claude/templates/issue-task.md)
