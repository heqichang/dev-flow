# DevFlow CLI 使用教程

## 目录

- [简介](#简介)
- [安装](#安装)
- [全局选项](#全局选项)
- [项目初始化 - init](#项目初始化---init)
- [脚本运行 - run](#脚本运行---run)
- [环境变量管理 - env](#环境变量管理---env)
- [项目配置管理 - config](#项目配置管理---config)
- [工作区管理 - workspace](#工作区管理---workspace)
- [Git 工作流自动化 - git](#git-工作流自动化---git)
- [任务管理 - task](#任务管理---task)
- [代码质量 - quality](#代码质量---quality)
- [依赖管理 - deps](#依赖管理---deps)
- [Shell 补全 - completion](#shell-补全---completion)
- [配置文件详解](#配置文件详解)
- [完整工作流示例](#完整工作流示例)

---

## 简介

DevFlow 是一款面向开发者的命令行工作流工具，帮助您：

- 快速初始化项目
- 管理项目配置
- 运行和管理开发脚本
- 处理环境变量和密钥
- 管理多项目工作区
- 自动化 Git 工作流（Git Flow / GitHub Flow）
- 任务管理与分支/PR 自动关联
- 代码质量检查与 CI 集成
- 多语言依赖管理

---

## 安装

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/your-org/devflow.git
cd devflow

# 编译安装
make install

# 或直接构建
make build
```

### Windows 构建

```powershell
.\build.ps1 build
```

### 验证安装

```bash
devflow version
```

---

## 全局选项

以下选项适用于所有命令：

| 选项 | 简写 | 说明 |
|------|------|------|
| `--no-color` | `-n` | 禁用彩色输出 |
| `--verbose` | `-v` | 显示详细输出 |

示例：

```bash
# 禁用彩色输出
devflow --no-color run build

# 显示详细输出
devflow -v run test
```

---

## 项目初始化 - init

创建一个新项目，支持交互式和参数式两种方式。

### 基本用法

```bash
# 交互式初始化（逐步引导）
devflow init
```

### 命令行参数

| 选项 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--name` | | | 项目名称 |
| `--template` | `-t` | | 项目模板 |
| `--language` | `-l` | | 编程语言 |
| `--author` | `-a` | | 作者 |
| `--license` | `-L` | `MIT` | 许可证 |
| `--git` | `-g` | `true` | 初始化 Git 仓库 |
| `--force` | `-f` | `false` | 强制覆盖现有目录 |

### 可用模板

| 模板名 | 说明 | 默认语言 |
|--------|------|----------|
| `frontend` | 前端项目 | javascript |
| `backend` | 后端项目 | go |
| `fullstack` | 全栈项目 | javascript |
| `cli` | CLI 工具 | go |
| `library` | 库项目 | javascript |

### 示例

```bash
# 交互式创建项目
devflow init

# 一行命令创建前端项目
devflow init --name my-app --template frontend --author "Zhang San"

# 创建后端项目并指定语言
devflow init -n my-api -t backend -l go -a "Zhang San"

# 强制覆盖已存在的目录
devflow init -n my-app -t frontend --force

# 创建项目但不初始化 Git
devflow init -n my-app -t cli --git=false
```

### 交互式流程

当不提供参数时，`devflow init` 会进入交互模式：

```
DevFlow 项目初始化

? 项目名称: my-project
可用模板:
  1. frontend  - 前端项目
  2. backend   - 后端项目
  3. fullstack - 全栈项目
  4. cli       - CLI 工具
  5. library   - 库项目
? 选择模板 (1-5): 2
? 编程语言 [go]:
? 作者: Zhang San
```

### 生成的项目结构

以前端项目为例：

```
my-project/
├── .devflow.yml        # DevFlow 配置文件
├── .gitignore
├── README.md
├── package.json
├── vite.config.js
├── public/
│   └── index.html
└── src/
    ├── index.js
    └── App.js
```

---

## 脚本运行 - run

运行在 `.devflow.yml` 中预定义的项目脚本。

### 基本用法

```bash
# 列出所有可用脚本
devflow run

# 运行指定脚本
devflow run <script-name>
```

### 命令行参数

| 选项 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--parallel` | `-p` | `false` | 并行执行所有依赖脚本 |
| `--env` | `-e` | `development` | 运行环境 |
| `--timeout` | `-t` | `0` | 超时时间（秒），0 表示不限制 |

### 支持的环境

| 参数值 | 等效简写 | 环境 |
|--------|----------|------|
| `development` | `dev` | 开发环境 |
| `staging` | | 预发布环境 |
| `production` | `prod` | 生产环境 |

### 示例

```bash
# 列出所有可用脚本
devflow run

# 运行开发服务器
devflow run dev

# 在生产环境下运行构建
devflow run build --env production

# 使用简写指定环境
devflow run build -e prod

# 设置超时时间（60秒）
devflow run build --timeout 60

# 并行执行依赖脚本
devflow run test --parallel

# 组合使用
devflow run build -e staging -t 120 -p
```

### 脚本依赖

脚本可以定义依赖关系，运行时会先执行依赖脚本：

```yaml
scripts:
  test:
    command: npm test
    dependsOn:
      - clean
    timeout: 60
  clean:
    command: rm -rf dist/
```

执行 `devflow run test` 时，会先执行 `clean`，再执行 `test`。

- 默认情况下依赖脚本**顺序执行**
- 使用 `--parallel` / `-p` 标志可以**并行执行**依赖脚本

### 超时机制

超时优先级：**命令行 `--timeout` > 脚本配置 `timeout` > 无限制**

```bash
# 使用脚本配置中的 timeout
devflow run build

# 命令行覆盖超时为 120 秒
devflow run build --timeout 120
```

---

## 环境变量管理 - env

管理项目的 `.env` 文件和环境变量。

### 子命令概览

| 子命令 | 说明 |
|--------|------|
| `list` | 列出所有环境变量 |
| `get` | 获取单个环境变量 |
| `set` | 设置环境变量 |
| `switch` | 切换环境文件 |
| `check` | 检查必需的环境变量 |

### env list - 列出环境变量

```bash
devflow env list
```

输出示例：

```
环境变量
==================================================

  API_KEY=your-api-key
  DATABASE_URL=postgres://localhost:5432/mydb
  DEBUG=true
  NODE_ENV=development

共 4 个环境变量
```

### env get - 获取单个环境变量

```bash
devflow env get <key>
```

示例：

```bash
# 获取 DATABASE_URL 的值
devflow env get DATABASE_URL

# 输出: postgres://localhost:5432/mydb
```

### env set - 设置环境变量

```bash
devflow env set <key> <value>
```

示例：

```bash
# 设置环境变量
devflow env set NODE_ENV production

# 设置包含空格的值（需要引号）
devflow env set API_KEY "my secret key"

# 首次 set 会自动创建 .env 文件
devflow env set NEW_VAR hello
```

### env switch - 切换环境文件

将 `.env.<environment>` 的内容复制为 `.env`，实现环境切换。

```bash
devflow env switch <environment>
```

前提：需要提前创建对应的环境文件：

```
项目根目录/
├── .env                # 当前生效的环境文件
├── .env.development    # 开发环境
├── .env.staging        # 预发布环境
└── .env.production     # 生产环境
```

示例：

```bash
# 切换到开发环境
devflow env switch development

# 切换到生产环境
devflow env switch production
```

### env check - 检查必需环境变量

根据 `.devflow.yml` 中 `requiredEnv` 的配置，检查 `.env` 文件中是否已设置所有必需变量。

```bash
devflow env check
```

`.devflow.yml` 中的配置：

```yaml
requiredEnv:
  - API_KEY
  - DATABASE_URL
```

输出示例（全部满足）：

```
检查必需环境变量...
✓ API_KEY 已设置
✓ DATABASE_URL 已设置
✓ 所有必需环境变量已设置
```

输出示例（有缺失）：

```
检查必需环境变量...
✓ API_KEY 已设置
✗ DATABASE_URL 未设置
✗ 缺少以下必需环境变量: DATABASE_URL
```

---

## 项目配置管理 - config

管理 `.devflow.yml` 项目配置文件。

### 子命令概览

| 子命令 | 说明 |
|--------|------|
| `show` | 显示当前配置 |
| `get` | 获取单个配置项 |
| `set` | 设置单个配置项 |
| `validate` | 验证配置文件 |

### config show - 显示配置

```bash
devflow config show
```

输出示例：

```
项目配置
==================================================

项目名称: my-awesome-project
版本: 1.0.0
语言: javascript
框架: react
作者: Zhang San
许可证: MIT

可用脚本:
  dev: npm run dev
  build: npm run build
  test: npm test
  lint: npm run lint
  clean: rm -rf dist/

环境变量:
  NODE_ENV=development
  PORT=3000
```

### config get - 获取配置项

```bash
devflow config get <key>
```

支持的 key：

| key | 说明 |
|-----|------|
| `projectName` | 项目名称 |
| `version` | 版本号 |
| `language` | 编程语言 |
| `framework` | 框架 |
| `author` | 作者 |
| `license` | 许可证 |

示例：

```bash
devflow config get projectName
# 输出: my-awesome-project

devflow config get version
# 输出: 1.0.0
```

### config set - 设置配置项

```bash
devflow config set <key> <value>
```

示例：

```bash
# 修改版本号
devflow config set version 2.0.0

# 修改框架
devflow config set framework vue

# 修改作者
devflow config set author "Li Si"
```

### config validate - 验证配置

检查 `.devflow.yml` 配置文件是否合法。

```bash
devflow config validate
```

输出示例：

```
✓ 配置文件验证通过
```

如果配置有误（如缺少必填的 `projectName`）：

```
✗ 配置验证失败: 配置验证失败: Key: 'ProjectConfig.ProjectName' Error:Field validation for 'ProjectName' failed on the 'required' tag
```

---

## 工作区管理 - workspace

管理多项目工作区，支持批量操作和项目依赖管理。

### 子命令概览

| 子命令 | 说明 |
|--------|------|
| `init` | 初始化工作区 |
| `add` | 添加项目到工作区 |
| `list` | 列出工作区项目 |
| `status` | 查看所有项目 Git 状态 |
| `sync` | 批量同步所有项目 |
| `clone` | 克隆缺失的项目 |

### workspace init - 初始化工作区

在当前目录创建工作区配置文件 `.devflow.workspace.yml`。

```bash
# 使用默认名称（当前目录名）
devflow workspace init

# 指定工作区名称
devflow workspace init my-workspace
```

输出示例：

```
初始化工作区

工作区初始化成功！
工作区目录: /projects/my-workspace
配置文件: /projects/my-workspace/.devflow.workspace.yml

下一步:
  devflow workspace add <name> <path> --url <repo-url>
```

### workspace add - 添加项目

将项目添加到工作区配置中。

```bash
devflow workspace add <name> <path> [flags]
```

| 选项 | 说明 |
|------|------|
| `--url` | Git 仓库 URL（用于克隆） |
| `--depends-on` | 依赖的项目名称（可多个） |

示例：

```bash
# 添加本地项目
devflow workspace add frontend ./frontend

# 添加带远程仓库的项目
devflow workspace add backend ./backend --url https://github.com/org/backend.git

# 添加带依赖关系的项目
devflow workspace add api ./api --url https://github.com/org/api.git --depends-on backend

# 添加多个依赖
devflow workspace add web ./web --depends-on frontend --depends-on api
```

### workspace list - 列出项目

列出工作区中的所有项目及其依赖关系。

```bash
devflow workspace list
```

输出示例：

```
工作区: my-workspace
项目列表:

1. frontend
  路径: ./frontend
  仓库: https://github.com/org/frontend.git

2. backend
  路径: ./backend
  仓库: https://github.com/org/backend.git

3. api
  路径: ./api
  仓库: https://github.com/org/api.git
  依赖: backend
```

### workspace status - 查看项目状态

并发获取工作区中所有项目的 Git 状态（分支、是否有未提交更改、领先/落后远程）。

```bash
devflow workspace status
```

输出示例：

```
工作区: my-workspace
项目状态:

✓ frontend [feature/login] 干净
! backend [develop] 有未提交更改
✓ api [main] 本地领先
✓ shared [develop] 干净
```

状态图标说明：

| 图标 | 含义 |
|------|------|
| ✓ 绿色 | 项目干净，无未提交更改 |
| ! 黄色 | 有未提交更改 |

分支颜色说明：

| 颜色 | 含义 |
|------|------|
| 青色 | 普通分支 |
| 红色 | 保护分支（main/master/develop/release/production） |

### workspace sync - 批量同步

并发拉取所有项目的最新代码（3 并发限制）。

```bash
devflow workspace sync
```

输出示例：

```
同步工作区项目

正在同步项目: frontend
正在同步项目: backend
正在同步项目: api
项目 frontend 同步完成
项目 backend 同步完成
项目 api 同步完成

所有项目同步完成！
```

如果部分项目同步失败：

```
部分项目同步失败:
项目 backend 同步失败: pull 失败: ...
```

### workspace clone - 克隆缺失项目

克隆工作区中尚未在本地的项目。

```bash
# 交互式确认每个克隆
devflow workspace clone

# 自动确认所有克隆
devflow workspace clone --yes
devflow workspace clone -y
```

输出示例：

```
克隆项目 frontend (https://github.com/org/frontend.git) 到 /projects/my-workspace/frontend? [Y/n] y
正在克隆项目: frontend
项目 frontend 克隆完成
```

### 工作区配置文件

工作区配置存储在 `.devflow.workspace.yml` 中：

```yaml
workspaceName: my-workspace
projects:
  - name: frontend
    path: ./frontend
    url: https://github.com/org/frontend.git
  - name: backend
    path: ./backend
    url: https://github.com/org/backend.git
  - name: api
    path: ./api
    url: https://github.com/org/api.git
    dependsOn:
      - backend
```

---

## Git 工作流自动化 - git

自动化 Git 工作流，支持 Git Flow 规范、Conventional Commits 和变更日志生成。

### 子命令概览

| 子命令 | 说明 |
|--------|------|
| `flow start` | 创建功能/修复/发布分支 |
| `flow finish` | 合并分支并删除 |
| `flow pr` | 创建 Pull Request |
| `commit` | 交互式提交（Conventional Commits） |
| `check` | 检查提交规范 |
| `changelog` | 生成变更日志 |

### git flow start - 创建分支

遵循 Git Flow 规范，从正确的基准分支创建新分支。

```bash
devflow git flow start <feature|hotfix|release> <name>
```

分支类型与基准分支对应关系：

| 分支类型 | 基准分支 | 用途 |
|----------|----------|------|
| `feature` | `develop` | 新功能开发 |
| `hotfix` | `main` | 紧急修复 |
| `release` | `main` | 版本发布 |

示例：

```bash
# 创建功能分支（从 develop 创建 feature/login）
devflow git flow start feature login

# 创建修复分支（从 main 创建 hotfix/fix-auth）
devflow git flow start hotfix fix-auth

# 创建发布分支（从 main 创建 release/v2.0）
devflow git flow start release v2.0
```

输出示例：

```
从 develop 创建分支 feature/login
分支 feature/login 创建成功
```

如果在保护分支上操作，会收到警告：

```
警告: 当前在保护分支 main 上操作
建议在功能分支上进行开发
```

### git flow finish - 完成分支

将当前功能分支合并回目标分支，并可选删除分支。

```bash
# 合并并删除分支
devflow git flow finish

# 合并但保留分支
devflow git flow finish --no-delete
devflow git flow finish -n
```

合并目标分支规则：

| 分支类型 | 合并目标 |
|----------|----------|
| `feature/*` | `develop` |
| `hotfix/*` | `main` |
| `release/*` | `main` |

输出示例：

```
合并 feature/login 到 develop
删除分支 feature/login
合并完成
```

### git flow pr - 创建 Pull Request

显示 PR 信息和建议的创建 URL。

```bash
devflow git flow pr [flags]
```

| 选项 | 说明 |
|------|------|
| `--title` | PR 标题（默认为当前分支名） |
| `--body` | PR 描述 |
| `--base` | 目标分支（默认根据分支类型推断） |

目标分支推断规则：

| 当前分支 | 推断目标 |
|----------|----------|
| `feature/*` | `develop` |
| 其他 | `main` |

示例：

```bash
# 使用默认参数
devflow git flow pr

# 指定标题和描述
devflow git flow pr --title "Add login feature" --body "Implements OAuth2 login"

# 指定目标分支
devflow git flow pr --base develop
```

输出示例：

```
Pull Request 信息:
  源分支: feature/login
  目标分支: develop
  标题: feature/login
  描述:
  仓库: https://github.com/org/my-project

PR 创建需要配置 GitHub/GitLab 凭据
请手动创建 Pull Request 或配置 API 访问令牌
建议的 PR URL 格式: https://github.com/org/my-project/compare/develop...feature/login
```

### git commit - 交互式提交

使用 Conventional Commits 规范交互式创建提交。

```bash
devflow git commit
```

交互式流程：

```
Conventional Commits 引导
可用的提交类型:
  1. feat     - 新功能
  2. fix      - 修复 bug
  3. docs     - 文档更新
  4. style    - 代码格式
  5. refactor - 重构
  6. perf     - 性能优化
  7. test     - 测试
  8. chore    - 构建/工具
  9. build    - 构建系统
  10. ci      - CI 配置
  11. revert  - 回滚

选择提交类型 (1-11): 1
作用域 (可选，直接回车跳过): auth
简短描述: add OAuth2 login support
详细描述 (可选，直接回车跳过):

提交信息: feat(auth): add OAuth2 login support
提交成功！
```

### git check - 检查提交规范

检查最后一次提交是否符合 Conventional Commits 规范。

```bash
devflow git check
```

输出示例（符合规范）：

```
✓ 提交符合 Conventional Commits 规范
```

输出示例（不符合规范）：

```
✗ 提交不符合规范: 不符合 Conventional Commits 规范
提交信息: update some files
```

Conventional Commits 格式：

```
<type>(<scope>): <subject>

示例:
  feat: add login page
  fix(api): handle null response
  feat(core)!: breaking API change
  docs: update README
```

### git changelog - 生成变更日志

基于提交历史自动生成变更日志，按类型分组。

```bash
devflow git changelog [flags]
```

| 选项 | 简写 | 说明 |
|------|------|------|
| `--since` | | 从哪个标签开始 |
| `--output` | `-o` | 输出到文件 |

示例：

```bash
# 输出到终端
devflow git changelog

# 输出到文件
devflow git changelog -o CHANGELOG.md

# 从指定标签开始
devflow git changelog --since v1.0.0
```

输出示例：

```
# Changelog - 2024-01-15

## 新功能

- feat(auth): add OAuth2 login support (a1b2c3d)
- feat(api): add pagination (e4f5g6h)

## 修复

- fix(core): handle null response (i7j8k9l)

## 文档

- docs: update API reference (m0n1o2p)

## 其他

- chore: update dependencies (q3r4s5t)
```

---

## 任务管理 - task

项目任务管理，支持与 Git 分支和 PR 自动关联。

### 子命令概览

| 子命令 | 说明 |
|--------|------|
| `add` | 添加任务 |
| `list` | 列出任务 |
| `start` | 开始任务（自动创建分支） |
| `done` | 完成任务（提示创建 PR） |
| `update` | 更新任务 |
| `delete` | 删除任务 |
| `show` | 查看任务详情 |

### task add - 添加任务

```bash
devflow task add <title> [flags]
```

| 选项 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--description` | `-d` | | 任务描述 |
| `--priority` | `-p` | `medium` | 优先级 (low, medium, high, urgent) |
| `--tags` | | | 标签（逗号分隔） |

示例：

```bash
# 添加简单任务
devflow task add "实现用户登录功能"

# 添加带描述和优先级的任务
devflow task add "修复支付超时" -d "支付接口超过30秒未响应" -p urgent

# 添加带标签的任务
devflow task add "优化首页加载" -p high --tags performance,frontend
```

输出示例：

```
任务 #1 已添加
  标题: 修复支付超时
  优先级: urgent
```

### task list - 列出任务

```bash
devflow task list [flags]
```

| 选项 | 简写 | 说明 |
|------|------|------|
| `--status` | `-s` | 过滤状态 (todo, in_progress, done) |

示例：

```bash
# 列出所有任务
devflow task list

# 只看待办任务
devflow task list -s todo

# 只看进行中的任务
devflow task list -s in_progress

# 只看已完成的任务
devflow task list -s done
```

输出示例：

```
任务列表

● 🟠 #1 修复支付超时 bug
    分支: task/1-fix-payment-timeout
○ 🟡 #2 优化首页加载 performance frontend
✓ 🟢 #3 更新文档
```

状态图标说明：

| 图标 | 状态 |
|------|------|
| ○ 白色 | 待办 (todo) |
| ● 黄色 | 进行中 (in_progress) |
| ✓ 绿色 | 已完成 (done) |

优先级图标说明：

| 图标 | 优先级 |
|------|--------|
| 🔴 | urgent |
| 🟠 | high |
| 🟡 | medium |
| 🟢 | low |

### task start - 开始任务

开始任务时自动创建 `task/{id}-{title}` 格式的分支。

```bash
devflow task start <id>
```

示例：

```bash
devflow task start 1
```

输出示例：

```
创建分支: task/1-fix-payment-timeout
任务 #1 已开始
分支: task/1-fix-payment-timeout
```

如果当前分支有未提交更改，会收到警告：

```
当前分支有未提交的更改，请先提交或 stash
```

### task done - 完成任务

完成任务时检查分支关联，并提示创建 PR。

```bash
devflow task done <id>
```

示例：

```bash
devflow task done 1
```

输出示例：

```
请创建 Pull Request:
  分支: task/1-fix-payment-timeout
  仓库: https://github.com/org/my-project

任务 #1 已完成
```

如果当前不在任务分支上：

```
当前不在任务分支 task/1-fix-payment-timeout 上
任务 #1 已完成
```

### task update - 更新任务

```bash
devflow task update <id> [flags]
```

| 选项 | 简写 | 说明 |
|------|------|------|
| `--title` | `-t` | 新标题 |
| `--description` | `-d` | 新描述 |
| `--priority` | `-p` | 新优先级 |
| `--tags` | | 新标签 |

示例：

```bash
# 修改标题
devflow task update 1 -t "修复支付超时问题"

# 修改优先级
devflow task update 1 -p high

# 修改描述
devflow task update 1 -d "支付接口需要增加重试机制"
```

### task delete - 删除任务

```bash
# 交互式确认删除
devflow task delete <id>

# 强制删除（不确认）
devflow task delete <id> --force
devflow task delete <id> -f
```

### task show - 查看任务详情

```bash
devflow task show <id>
```

输出示例：

```
任务 #1

标题: 修复支付超时
状态: in_progress
优先级: urgent
标签: bug, payment
描述: 支付接口超过30秒未响应
分支: task/1-fix-payment-timeout
创建时间: 2024-01-15T10:30:00+08:00
更新时间: 2024-01-15T11:00:00+08:00
开始时间: 2024-01-15T11:00:00+08:00
```

### 任务数据存储

任务数据存储在项目根目录的 `.devflow/tasks.json` 中：

```json
{
  "tasks": [
    {
      "id": 1,
      "title": "修复支付超时",
      "description": "支付接口超过30秒未响应",
      "status": "in_progress",
      "priority": "urgent",
      "tags": ["bug", "payment"],
      "branch": "task/1-fix-payment-timeout",
      "createdAt": "2024-01-15T10:30:00+08:00",
      "updatedAt": "2024-01-15T11:00:00+08:00",
      "startedAt": "2024-01-15T11:00:00+08:00"
    }
  ],
  "nextID": 2,
  "updatedAt": "2024-01-15T11:00:00+08:00"
}
```

---

## 代码质量 - quality

代码质量检查工具，自动检测项目语言并运行对应的检查工具。

### 子命令概览

| 子命令 | 别名 | 说明 |
|--------|------|------|
| `lint` | | 运行代码检查 |
| `test` | | 运行测试 |
| `check` | | 综合检查（lint + test + type check） |
| `hook` | | 安装 pre-commit hook |
| `ci` | | 生成 CI 配置 |

### 支持的语言和工具

| 语言 | Lint | Test | Type Check |
|------|------|------|------------|
| Go | go vet / golangci-lint | go test | go build |
| JavaScript | eslint | npm test | - |
| TypeScript | eslint | npm test | tsc --noEmit |
| Python | flake8 | pytest | - |
| Rust | cargo clippy | cargo test | cargo check |

### quality lint - 代码检查

```bash
devflow quality lint
```

自动检测项目语言并运行对应的 linter。如果 linter 未安装，会跳过并提示。

输出示例：

```
检测到项目语言: go
运行代码检查...
代码检查通过 (1.2s)
```

如果检查失败：

```
检测到项目语言: go
运行代码检查...
代码检查发现问题:
main.go:10:2: undeclared name: fmt
```

### quality test - 运行测试

```bash
devflow quality test [flags]
```

| 选项 | 简写 | 说明 |
|------|------|------|
| `--watch` | `-w` | watch 模式（仅 JS/TS 项目支持） |

示例：

```bash
# 运行测试
devflow quality test

# watch 模式
devflow quality test --watch
```

### quality check - 综合检查

一次性运行 lint + test + type check。

```bash
devflow quality check [flags]
```

| 选项 | 简写 | 说明 |
|------|------|------|
| `--json` | `-j` | JSON 格式输出 |

示例：

```bash
# 表格输出
devflow quality check

# JSON 输出
devflow quality check --json
```

输出示例：

```
DevFlow 综合检查
检测到项目语言: go

检查项          状态    耗时
========================================
lint            ✓ 通过  1.2s
test            ✓ 通过  3.5s
typecheck       ✓ 通过  2.1s

所有检查通过！
```

JSON 输出示例：

```json
{
  "language": "go",
  "results": [
    {"name": "lint", "passed": true, "duration": "1.2s"},
    {"name": "test", "passed": true, "duration": "3.5s"},
    {"name": "typecheck", "passed": true, "duration": "2.1s"}
  ],
  "passed": true
}
```

### quality hook - 安装 pre-commit hook

安装 Git pre-commit hook，在每次提交前自动运行 `devflow check`。

```bash
devflow quality hook
```

输出示例：

```
pre-commit hook 已安装
```

安装后，每次 `git commit` 会自动触发：

```
Running DevFlow checks...
[DevFlow check 输出...]

# 如果检查通过，提交继续
# 如果检查失败，提交被中止
DevFlow checks failed. Commit aborted.
```

幂等性：重复运行不会重复添加 hook 内容。

### quality ci - 生成 CI 配置

根据项目语言生成 CI 配置文件。

```bash
devflow quality ci <platform>
```

支持的平台：

| 平台 | 生成文件 |
|------|----------|
| `github` | `.github/workflows/devflow.yml` |
| `gitlab` | `.gitlab-ci.yml` |

示例：

```bash
# 生成 GitHub Actions 配置
devflow quality ci github

# 生成 GitLab CI 配置
devflow quality ci gitlab
```

---

## 依赖管理 - deps

多语言依赖管理工具，支持列出、检查、更新和审计依赖。

### 子命令概览

| 子命令 | 说明 |
|--------|------|
| `list` | 列出项目依赖 |
| `outdated` | 检查过时依赖 |
| `update` | 更新依赖 |
| `audit` | 安全漏洞检查 |

### 支持的包管理器

| 包管理器 | 语言 | 检测标识文件 |
|----------|------|-------------|
| npm | JavaScript | `package-lock.json` / `package.json` |
| yarn | JavaScript | `yarn.lock` |
| pnpm | JavaScript | `pnpm-lock.yaml` |
| pip | Python | `requirements.txt` |
| poetry | Python | `poetry.lock` |
| cargo | Rust | `Cargo.toml` |
| go mod | Go | `go.mod` |
| maven | Java | `pom.xml` |

### deps list - 列出依赖

```bash
devflow deps list [flags]
```

| 选项 | 简写 | 说明 |
|------|------|------|
| `--json` | `-j` | JSON 格式输出 |

示例：

```bash
# 表格输出
devflow deps list

# JSON 输出
devflow deps list --json
```

输出示例：

```
检测到包管理器: go

包管理器: go
依赖数量: 5

名称                                当前版本    最新版本    状态
================================================================
github.com/spf13/cobra              v1.8.0     -          最新
github.com/fatih/color              v1.16.0    -          最新
github.com/stretchr/testify         v1.8.4     -          最新
gopkg.in/yaml.v3                    v3.0.1     -          最新
github.com/joho/godotenv            v1.5.1     -          最新

所有依赖都是最新的
```

### deps outdated - 检查过时依赖

```bash
devflow deps outdated [flags]
```

| 选项 | 简写 | 说明 |
|------|------|------|
| `--json` | `-j` | JSON 格式输出 |

示例：

```bash
devflow deps outdated
```

输出示例：

```
检测到包管理器: npm
检查过时依赖...

包管理器: npm
依赖数量: 12

名称            当前版本    最新版本    状态
================================================================
react           18.0.0     18.2.0     ✗ 过时
react-dom       18.0.0     18.2.0     ✗ 过时
eslint          8.40.0     8.56.0     ✗ 过时

发现 3 个过时依赖
运行 'devflow deps update' 更新依赖
```

过时检测支持情况：

| 包管理器 | 支持状态 |
|----------|----------|
| npm/yarn/pnpm | ✅ 支持 |
| go mod | ⚠️ 有限支持 |
| cargo | ✅ 支持（需安装 cargo-outdated） |
| pip/poetry/maven | ❌ 暂不支持 |

### deps update - 更新依赖

```bash
devflow deps update [flags]
```

| 选项 | 简写 | 说明 |
|------|------|------|
| `--yes` | `-y` | 跳过确认直接更新 |

示例：

```bash
# 交互式确认
devflow deps update

# 跳过确认
devflow deps update -y
```

交互式流程：

```
使用包管理器: npm
更新所有依赖? [y/N] y
运行 npm update...
依赖更新完成
```

各包管理器更新命令：

| 包管理器 | 执行命令 |
|----------|----------|
| npm | `npm update` |
| yarn | `yarn upgrade` |
| pnpm | `pnpm update` |
| pip | `pip install --upgrade -r requirements.txt` |
| cargo | `cargo update` |
| go mod | `go get -u ./...` |

### deps audit - 安全漏洞检查

```bash
devflow deps audit
```

示例：

```bash
devflow deps audit
```

输出示例（npm）：

```
使用包管理器: npm
运行 npm audit...

# 如果发现漏洞，会显示详细信息
found 2 vulnerabilities (1 low, 1 high)
...

发现安全漏洞
```

输出示例（无漏洞）：

```
使用包管理器: npm
运行 npm audit...
未发现安全漏洞
```

安全审计支持情况：

| 包管理器 | 支持状态 |
|----------|----------|
| npm/pnpm | ✅ `npm audit` |
| yarn | ✅ `yarn audit` |
| go mod | ⚠️ 列出依赖，建议使用 govulncheck |
| cargo | ✅ `cargo audit`（需安装 cargo-audit） |
| pip/poetry/maven | ❌ 暂不支持 |

---

## Shell 补全 - completion

为指定的 shell 生成自动补全脚本。

### 基本用法

```bash
devflow completion <shell>
```

支持的 shell：`bash`、`zsh`、`fish`、`powershell`

### Bash

```bash
# 系统级安装
devflow completion bash > /etc/bash_completion.d/devflow

# 用户级安装
devflow completion bash > ~/.bash_completion.d/devflow
```

安装后重新加载 shell：

```bash
source ~/.bashrc
```

### Zsh

```bash
devflow completion zsh > ~/.zsh/completion/_devflow
```

确保 `~/.zsh/completion` 在你的 `fpath` 中：

```zsh
# 在 ~/.zshrc 中添加
fpath=(~/.zsh/completion $fpath)
autoload -U compinit && compinit
```

### Fish

```bash
devflow completion fish > ~/.config/fish/completions/devflow.fish
```

### PowerShell

```powershell
devflow completion powershell >> $PROFILE
```

重新加载 Profile：

```powershell
. $PROFILE
```

---

## 配置文件详解

DevFlow 使用 `.devflow.yml` 作为项目配置文件，放在项目根目录下。

### 完整配置示例

```yaml
projectName: my-awesome-project
version: 1.0.0
language: javascript
framework: react
author: Your Name
license: MIT
description: An awesome project built with DevFlow

scripts:
  dev:
    command: npm run dev
    timeout: 0

  build:
    command: npm run build
    timeout: 120

  test:
    command: npm test
    dependsOn:
      - clean
    timeout: 60

  lint:
    command: npm run lint

  clean:
    command: rm -rf dist/
    env:
      FORCE: true

env:
  NODE_ENV: development
  PORT: "3000"

requiredEnv:
  - API_KEY
  - DATABASE_URL

environments:
  development:
    env:
      DEBUG: "true"
      LOG_LEVEL: debug

  staging:
    env:
      DEBUG: "false"
      LOG_LEVEL: info

  production:
    env:
      DEBUG: "false"
      LOG_LEVEL: warn
      NODE_ENV: production
    scripts:
      build:
        command: npm run build -- --mode production
        timeout: 180
```

### 字段说明

#### 顶层字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `projectName` | string | 是 | 项目名称 |
| `version` | string | 否 | 项目版本 |
| `language` | string | 否 | 编程语言 |
| `framework` | string | 否 | 使用的框架 |
| `author` | string | 否 | 作者 |
| `license` | string | 否 | 许可证，默认 MIT |
| `description` | string | 否 | 项目描述 |
| `scripts` | map | 否 | 脚本定义 |
| `env` | map | 否 | 全局环境变量 |
| `requiredEnv` | []string | 否 | 必需的环境变量列表 |
| `environments` | map | 否 | 环境特定配置 |

#### Script 字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `command` | string | 是 | 要执行的命令 |
| `dependsOn` | []string | 否 | 依赖的脚本列表 |
| `env` | map | 否 | 脚本级环境变量 |
| `timeout` | int | 否 | 超时时间（秒），0 表示不限制 |

#### environments 字段

`environments` 允许为不同环境定义覆盖配置，支持 `development`、`staging`、`production` 三个环境。

当使用 `devflow run <script> --env <environment>` 时，DevFlow 会：

1. 加载基础配置
2. 用对应环境的配置覆盖基础配置
3. 执行脚本

环境配置可以覆盖以下内容：

- `env` - 合并环境变量（追加/覆盖）
- `scripts` - 覆盖指定脚本定义
- 其他顶层字段 - 非空即覆盖

### 全局配置

DevFlow 还支持全局配置文件，位于：

```
~/.config/devflow/config.yml
```

可配置项：

```yaml
defaultEnv: development
```

---

## 完整工作流示例

### 1. 创建新项目

```bash
# 交互式创建
devflow init

# 或一行命令创建
devflow init -n my-saas -t fullstack -a "Zhang San" -L MIT
```

### 2. 进入项目目录

```bash
cd my-saas
```

### 3. 查看项目配置

```bash
devflow config show
```

### 4. 配置环境变量

```bash
# 创建开发环境文件
devflow env set NODE_ENV development
devflow env set PORT 3000
devflow env set API_KEY dev-api-key-123
devflow env set DATABASE_URL postgres://localhost:5432/my_saas_dev

# 创建生产环境文件（手动创建 .env.production）
echo "NODE_ENV=production" > .env.production
echo "PORT=8080" >> .env.production
echo "API_KEY=prod-api-key-xxx" >> .env.production
echo "DATABASE_URL=postgres://prod-db:5432/my_saas" >> .env.production
```

### 5. 检查必需环境变量

```bash
devflow env check
```

### 6. 运行开发脚本

```bash
# 查看可用脚本
devflow run

# 启动开发服务器
devflow run dev

# 运行测试（会先执行依赖脚本 clean）
devflow run test

# 并行执行依赖脚本
devflow run test -p
```

### 7. 使用任务管理

```bash
# 添加任务
devflow task add "实现用户登录" -p high --tags feature

# 开始任务（自动创建分支）
devflow task start 1

# 查看任务列表
devflow task list

# 完成任务
devflow task done 1
```

### 8. 使用 Git 工作流

```bash
# 创建功能分支
devflow git flow start feature user-auth

# 交互式提交
devflow git commit

# 检查提交规范
devflow git check

# 完成分支
devflow git flow finish

# 或创建 PR
devflow git flow pr --title "Add user auth"
```

### 9. 代码质量检查

```bash
# 运行 lint
devflow quality lint

# 运行测试
devflow quality test

# 综合检查
devflow quality check

# 安装 pre-commit hook
devflow quality hook

# 生成 CI 配置
devflow quality ci github
```

### 10. 依赖管理

```bash
# 列出依赖
devflow deps list

# 检查过时依赖
devflow deps outdated

# 更新依赖
devflow deps update

# 安全审计
devflow deps audit
```

### 11. 生成变更日志

```bash
# 生成变更日志
devflow git changelog -o CHANGELOG.md
```

### 12. 多项目工作区

```bash
# 初始化工作区
devflow workspace init my-workspace

# 添加项目
devflow workspace add frontend ./frontend --url https://github.com/org/frontend.git
devflow workspace add backend ./backend --url https://github.com/org/backend.git
devflow workspace add shared ./shared --depends-on backend

# 查看项目状态
devflow workspace status

# 批量同步
devflow workspace sync

# 克隆缺失项目
devflow workspace clone -y
```

### 13. 构建生产版本

```bash
# 使用生产环境配置构建
devflow run build -e production -t 180
```

### 14. 切换环境

```bash
# 切换到生产环境
devflow env switch production

# 切换回开发环境
devflow env switch development
```

### 15. 修改项目配置

```bash
# 更新版本号
devflow config set version 1.1.0

# 验证配置
devflow config validate

# 查看更新后的配置
devflow config show
```

### 16. 设置 Shell 补全（可选）

```bash
# Bash
devflow completion bash > ~/.bash_completion.d/devflow

# Zsh
devflow completion zsh > ~/.zsh/completion/_devflow

# Fish
devflow completion fish > ~/.config/fish/completions/devflow.fish

# PowerShell
devflow completion powershell >> $PROFILE
```

---

## 命令速查表

| 命令 | 说明 |
|------|------|
| `devflow version` | 查看版本 |
| `devflow init` | 初始化项目 |
| `devflow run` | 列出可用脚本 |
| `devflow run <name>` | 运行脚本 |
| `devflow run <name> -e prod` | 指定环境运行 |
| `devflow run <name> -t 60` | 设置超时 |
| `devflow run <name> -p` | 并行执行依赖 |
| `devflow env list` | 列出环境变量 |
| `devflow env get <key>` | 获取环境变量 |
| `devflow env set <key> <value>` | 设置环境变量 |
| `devflow env switch <env>` | 切换环境 |
| `devflow env check` | 检查必需变量 |
| `devflow config show` | 显示配置 |
| `devflow config get <key>` | 获取配置项 |
| `devflow config set <key> <value>` | 设置配置项 |
| `devflow config validate` | 验证配置 |
| `devflow workspace init [name]` | 初始化工作区 |
| `devflow workspace add <name> <path>` | 添加项目 |
| `devflow workspace list` | 列出项目 |
| `devflow workspace status` | 查看项目状态 |
| `devflow workspace sync` | 批量同步 |
| `devflow workspace clone` | 克隆缺失项目 |
| `devflow git flow start <type> <name>` | 创建分支 |
| `devflow git flow finish` | 完成分支 |
| `devflow git flow pr` | 创建 PR |
| `devflow git commit` | 交互式提交 |
| `devflow git check` | 检查提交规范 |
| `devflow git changelog` | 生成变更日志 |
| `devflow task add <title>` | 添加任务 |
| `devflow task list` | 列出任务 |
| `devflow task start <id>` | 开始任务 |
| `devflow task done <id>` | 完成任务 |
| `devflow task update <id>` | 更新任务 |
| `devflow task delete <id>` | 删除任务 |
| `devflow task show <id>` | 查看任务详情 |
| `devflow quality lint` | 代码检查 |
| `devflow quality test` | 运行测试 |
| `devflow quality check` | 综合检查 |
| `devflow quality hook` | 安装 pre-commit hook |
| `devflow quality ci <platform>` | 生成 CI 配置 |
| `devflow deps list` | 列出依赖 |
| `devflow deps outdated` | 检查过时依赖 |
| `devflow deps update` | 更新依赖 |
| `devflow deps audit` | 安全漏洞检查 |
| `devflow completion <shell>` | 生成补全脚本 |
