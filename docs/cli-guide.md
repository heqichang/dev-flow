# DevFlow CLI 使用教程

## 目录

- [简介](#简介)
- [安装](#安装)
- [全局选项](#全局选项)
- [项目初始化 - init](#项目初始化---init)
- [脚本运行 - run](#脚本运行---run)
- [环境变量管理 - env](#环境变量管理---env)
- [项目配置管理 - config](#项目配置管理---config)
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

### 7. 构建生产版本

```bash
# 使用生产环境配置构建
devflow run build -e production -t 180
```

### 8. 切换环境

```bash
# 切换到生产环境
devflow env switch production

# 切换回开发环境
devflow env switch development
```

### 9. 修改项目配置

```bash
# 更新版本号
devflow config set version 1.1.0

# 验证配置
devflow config validate

# 查看更新后的配置
devflow config show
```

### 10. 设置 Shell 补全（可选）

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
| `devflow completion <shell>` | 生成补全脚本 |
