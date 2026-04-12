# 模块化开发工作流规则

本文档定义了基于 Skills 和 TDD 驱动的开发流程规则。

## 核心原则

### 模块化思维
- 模块通过 TDD 隐式定义：测试边界 = 模块边界
- Deep Modules：小接口隐藏大实现（参见 tdd skill 的 deep-modules 文档）
- 不需要显式记录模块状态，测试覆盖率通过工具自动获取

### 信任机制
- 通过测试的模块默认可信，当作黑盒处理
- 只看模块的公开接口，不看内部实现
- 只有输入→输出不符合预期时，才进入模块内部排查
- 进入模块后，对子模块也是同样处理（递归黑盒）

---

## 开发流程

### 规划阶段

使用 Skills 创建 Issue：

| 场景 | Skill 链 | Issue 输出 |
|------|---------|-----------|
| 新需求 | `/write-a-prd` → `/prd-to-issues` | requirement + task(s) |
| Bug 修复 | `/qa` → `/triage-issue` | bug + task(s) |
| 架构重构 | `/improve-codebase-architecture` → `/request-refactor-plan` | requirement + task(s) |

可选辅助 Skills：
- `/design-an-interface`：设计接口（多种方案对比）
- `/grill-me`：面试式需求确认

### 执行阶段

使用 `/tdd` skill 执行实现：
1. 读取 Issue 内容
2. 识别涉及的模块（可信模块当作黑盒）
3. 垂直切片（tracer bullet）：一个测试 → 一个实现 → 重复
4. 重构
5. 提交代码

---

## Issue 结构

### 类型标签

| 标签 | 用途 |
|------|------|
| `type: requirement` | 需求/方案/重构计划 Issue |
| `type: bug` | Bug 报告 Issue |
| `type: task` | 执行 Issue（sub-issue） |

### 层级关系

```
requirement Issue (type: requirement)
├── task Issue #1 (type: task)
├── task Issue #2 (type: task)
└── task Issue #3 (type: task)

bug Issue (type: bug)
├── task Issue #4 (type: task)
└── task Issue #5 (type: task)
```

### Issue 模板位置
- 需求/方案：`.claude/templates/issue-requirement.md`
- 执行：`.claude/templates/issue-task.md`

---

## 分支命名规范

```
task/编号-功能简述
```

示例：
- `task/3-add-retry-strategy`
- `task/15-fix-signature-bug`

---

## 文档维护

### 项目文档结构

| 文件 | 内容 | 维护时机 |
|------|------|----------|
| `README.md` | 项目概述、架构、命令、配置、数据结构 | 项目信息变更时 |
| `CLAUDE.md` | AI 指引、开发流程、Skills 引用 | 流程变更时 |

---

## 检查清单

### 规划阶段完成
- [ ] 方案已获得用户确认
- [ ] 已检查 open issue 防止重复
- [ ] 父级 Issue 已创建（requirement 或 bug）
- [ ] 执行 Issue(s) 已创建并关联为 sub-issues

### 执行阶段完成
- [ ] 分支已创建
- [ ] 测试用例已编写（垂直切片）
- [ ] 代码已实现
- [ ] 测试通过
- [ ] 重构完成
- [ ] PR 已创建，commit message 标注 Issue 编号
