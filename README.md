# ros-address-list-tool

一个使用 Go 编写的 RouterOS Address List 管理工具。  
它把“多来源地址数据 → 规则合并 → RouterOS `.rsc` 渲染 → HTTP / Web / 自动化调用”收敛为一套可维护流程，适合多运营商分流、国内外分流、PCC 负载均衡、云端集中生成地址列表等场景。

---

## 目录

- [项目简介](#项目简介)
- [核心特性](#核心特性)
- [当前目录结构](#当前目录结构)
- [工作方式](#工作方式)
- [运行要求](#运行要求)
- [快速开始](#快速开始)
- [配置文件说明](#配置文件说明)
- [认证与登录](#认证与登录)
- [HTTP API](#http-api)
- [RouterOS 自动更新用法](#routeros-自动更新用法)
- [Web 管理界面](#web-管理界面)
- [构建与发布](#构建与发布)
- [部署建议](#部署建议)
- [常见问题](#常见问题)
- [开发与测试](#开发与测试)
- [许可证](#许可证)

---

## 项目简介

在 RouterOS 的多出口和策略路由场景中，`address-list` 往往规模很大，而且来源复杂：

- 运营商 IPv4 / IPv6 地址段
- 本地文件维护的 CIDR 列表
- 远程 URL 提供的地址数据
- 手工例外规则
- 当前在线状态导出的已有 address-list

如果直接在 RouterOS 上手工维护，通常会遇到这些问题：

- 条目数量大，手工维护容易出错
- IPv4 / IPv6、不同运营商、不同出口之间容易混乱
- 没有统一的生成与发布流程
- 自动化能力弱，不方便接入云端统一管理
- 很难同时兼顾 Web 管理与 RouterOS 定时自动拉取

这个项目的目标就是把整个过程统一起来：

1. 定义受管 list
2. 加载目标来源 `desired_sources`
3. 可选加载当前状态来源 `current_state_sources`
4. 叠加 `manual_rules`
5. 校验、规范化、去重、合并
6. 生成 RouterOS `.rsc`
7. 通过 CLI、HTTP API、Web 页面或 RouterOS 自动脚本使用结果

---

## 核心特性

### 1. 地址列表生成能力

- 支持同时管理多个 RouterOS address-list
- 支持 IPv4 / IPv6
- 支持 `replace_all` 和 `diff` 两种渲染模式
- 支持本地文件和远程 URL 作为 source
- 支持 `plain_cidr` 和 `json` 等来源格式
- 支持手工规则增删覆盖
- 支持地址校验、规范化与去重
- 支持将结果写入本地 `.rsc` 文件

### 2. HTTP 服务能力

- `GET /healthz` 健康检查
- `GET /api/v1/config` 查看当前配置
- `POST /api/v1/render` 触发一次渲染并返回 JSON 结果
- `GET /api/v1/render/latest.rsc` 直接生成并返回最新 `.rsc`
- `GET /api/v1/render/download` 兼容下载型调用
- 适合脚本、自动化系统、RouterOS `/tool fetch` 直接调用

### 3. Web 管理能力

- 独立登录页
- 登录成功后进入后台设置页
- 支持会话 Cookie
- 支持首次登录强制修改密码
- 后台可查看配置、执行渲染、进行管理操作

### 4. 自动化友好

- RouterOS 可以直接请求下载型渲染接口
- 每次请求时由服务端实时生成新内容
- 适合放在云服务器统一部署
- 浏览器登录和机器调用分离：浏览器走登录态，自动化脚本走 Token

---

## 当前目录结构

```text
ros-address-list-tool/
├── cmd/
│   └── server/
│       └── main.go
├── configs/
│   └── config.json
├── internal/
│   └── app/
│       ├── auth.go
│       ├── config.go
│       ├── engine.go
│       ├── logger.go
│       ├── merge.go
│       ├── render.go
│       ├── security.go
│       ├── server.go
│       ├── source.go
│       ├── source_probe.go
│       ├── store.go
│       ├── types.go
│       ├── validate.go
│       └── *_test.go
├── output/
├── scripts/
│   ├── build_linux_amd64.ps1
│   └── build_windows_amd64.ps1
├── web/
│   └── src/
│       ├── app.css
│       ├── app.js
│       ├── favicon.svg
│       ├── index.html
│       └── login.html
├── .gitignore
├── LICENSE
├── README.md
└── go.mod
```

### 目录说明

- `cmd/server/`：程序入口，支持一次渲染和 HTTP 服务模式
- `configs/`：默认配置文件
- `internal/app/`：核心业务层，包含配置、渲染、认证、HTTP、存储等逻辑
- `output/`：生成的 `.rsc` 输出目录
- `scripts/`：打包脚本
- `web/src/`：前端静态资源目录，当前服务默认从这里提供页面
- `go.mod`：Go 模块声明

---

## 工作方式

项目内部流程大致如下：

```text
配置文件
   ↓
lists + desired_sources + current_state_sources + manual_rules
   ↓
source 加载 / 校验 / 规范化 / 去重
   ↓
生成 desired snapshot / current snapshot
   ↓
按 replace_all 或 diff 渲染
   ↓
输出 RouterOS .rsc
   ↓
CLI / HTTP API / Web / RouterOS 自动拉取使用结果
```

### `replace_all`

全量替换模式。  
适合云端统一生成、RouterOS 定时导入的场景。通常更稳定，也更容易排查。

### `diff`

差异更新模式。  
要求 `current_state_sources` 足够准确，否则结果会偏差。更适合已有稳定状态管理体系的场景。

---

## 运行要求

- Go 版本与 `go.mod` 保持一致
- Windows / Linux / macOS 都可以开发
- URL source 场景需要能访问对应地址源
- 如果要用 Web 管理界面，需要启用 HTTP 服务模式
- 如果要给 RouterOS 自动拉取，建议部署在带 HTTPS 的云服务器或反向代理后面

---

## 快速开始

### 1. 克隆项目

```bash
git clone git@github.com:cslt-l/ros-address-list-tool.git
cd ros-address-list-tool
```

或者：

```bash
git clone https://github.com/cslt-l/ros-address-list-tool.git
cd ros-address-list-tool
```

### 2. 运行测试

```bash
go test ./...
```

### 3. 只执行一次渲染

```bash
go run ./cmd/server -config ./configs/config.json
```

执行后会：

- 读取配置
- 拉取 / 解析来源
- 合并规则
- 渲染 `.rsc`
- 写入 `output.path`

### 4. 以 HTTP 服务模式启动

```bash
go run ./cmd/server -config ./configs/config.json -serve -listen 127.0.0.1:8090
```

启动后可访问：

- 登录页：`http://127.0.0.1:8090/login.html`
- 后台：`http://127.0.0.1:8090/`
- 健康检查：`http://127.0.0.1:8090/healthz`

---

## 配置文件说明

一个典型配置如下：

```json
{
  "auto_create_lists": false,
  "log_file": "./logs/app.log",
  "lists": [],
  "desired_sources": [],
  "current_state_sources": [],
  "manual_rules": [],
  "output": {
    "path": "./output/routeros-address-list.rsc",
    "mode": "replace_all",
    "managed_comment": "managed-by-go"
  },
  "server": {
    "listen": "127.0.0.1:8090",
    "enable_web": true,
    "web_dir": "./web/src",
    "auth_token": "please-change-me",
    "login_enabled": true,
    "login_username": "admin",
    "login_password": "password",
    "password_change_required": true,
    "session_cookie_name": "ros_list_session",
    "session_ttl_minutes": 720
  }
}
```

### 顶层字段

#### `auto_create_lists`

是否允许在 source 或手工规则里出现未预定义 list 时自动创建。

- `true`：允许自动创建
- `false`：必须先在 `lists` 中定义

生产环境建议使用 `false`。

#### `log_file`

日志文件输出路径。

#### `lists`

受管 address-list 定义。示例：

```json
{
  "name": "toWanTelecom4",
  "family": "ipv4",
  "enabled": true,
  "description": "中国电信 IPv4"
}
```

建议 IPv4 / IPv6 使用不同名称，例如：

- `toWanTelecom4`
- `toWanTelecom6`

#### `desired_sources`

目标来源。  
会被加载并合并为“最终希望下发到 RouterOS 的状态”。

单个 URL source 示例：

```json
{
  "name": "china-telecom-ipv4",
  "type": "url",
  "url": "https://china-operator-ip.yfgao.com/chinanet.txt",
  "format": "plain_cidr",
  "target_list_name": "toWanTelecom4",
  "target_list_family": "ipv4",
  "line_comment_prefixes": ["#", "//", ";"],
  "enabled": true,
  "priority": 100,
  "timeout_seconds": 30
}
```

#### `current_state_sources`

当前状态来源。  
主要用于 `diff` 模式。

#### `manual_rules`

手工规则。  
用于最终增删和覆盖。

#### `output`

输出配置。

- `path`：写出的 `.rsc` 文件路径
- `mode`：`replace_all` 或 `diff`
- `managed_comment`：受管条目的 comment 标记

#### `server`

服务配置。

- `listen`：监听地址
- `enable_web`：是否启用 Web 静态页面
- `web_dir`：当前静态资源目录，建议保持 `./web/src`
- `auth_token`：机器调用 Token，适合脚本、RouterOS、CI
- `login_enabled`：是否启用网页登录
- `login_username`：后台登录用户名
- `login_password`：明文初始密码，首次改密后建议清空
- `password_change_required`：是否强制首次登录修改密码
- `session_cookie_name`：会话 Cookie 名称
- `session_ttl_minutes`：会话有效期

---

## 认证与登录

### 浏览器后台登录

浏览器访问：

```text
http://127.0.0.1:8090/login.html
```

如果启用了登录功能，用户通过登录页建立会话后，才能进入后台设置页面。

推荐做法：

- 初始账号设置为 `admin`
- 初始密码仅用于第一次登录
- `password_change_required` 设为 `true`
- 首次登录后立即改密

### 机器调用 Token

自动化调用推荐使用 `auth_token`，不要依赖浏览器登录态。

支持两种常见方式：

#### 1. Bearer Token

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" http://127.0.0.1:8090/api/v1/config
```

#### 2. 下载接口 URL 参数

```text
http://127.0.0.1:8090/api/v1/render/latest.rsc?access_token=YOUR_TOKEN&mode=replace_all
```

对 RouterOS 自动脚本来说，第二种通常更方便。

---

## HTTP API

下面列出当前结构下最常用的接口。

### 基础接口

#### `GET /healthz`

健康检查。

#### `GET /api/v1/config`

返回当前配置。

#### `POST /api/v1/render`

执行一次渲染并返回 JSON 结果。  
适合 Web 后台或自动化系统拿摘要信息。

### 认证接口

#### `POST /api/v1/auth/login`

登录并建立会话。

#### `POST /api/v1/auth/logout`

退出登录。

#### `GET /api/v1/auth/me`

查看当前会话状态。

### 下载型渲染接口

#### `GET /api/v1/render/latest.rsc`

访问时实时生成最新 `.rsc` 并直接返回文本内容。

示例：

```text
http://127.0.0.1:8090/api/v1/render/latest.rsc?access_token=YOUR_TOKEN&mode=replace_all
```

#### `GET /api/v1/render/download`

兼容下载型入口。  
如果你的现有 RouterOS 脚本已经用这条路径，可以继续保留使用。

---

## RouterOS 自动更新用法

这个项目现在特别适合下面这个流程：

1. 服务部署在云服务器
2. RouterOS 用 `/tool fetch` 请求下载型接口
3. 服务端当场生成最新 `.rsc`
4. RouterOS 下载后执行 `/import`

### 下载地址示例

```text
https://your-domain.example.com/api/v1/render/latest.rsc?access_token=YOUR_TOKEN&mode=replace_all
```

### RouterOS 最简示例

```routeros
:local apiBase "https://your-domain.example.com"
:local apiToken "YOUR_TOKEN"
:local fileName "ros-address-list-auto.rsc"
:local url ($apiBase . "/api/v1/render/latest.rsc?access_token=" . $apiToken . "&mode=replace_all")

/tool fetch url=$url dst-path=$fileName keep-result=yes check-certificate=yes
:delay 2s
/import file-name=$fileName
```

### 推荐做法

- 云服务器前面加 HTTPS
- 给自动化单独使用一个长随机 `auth_token`
- RouterOS 先 `dry-run` 或先在测试设备验证
- 生产环境优先使用 `replace_all`

---

## Web 管理界面

当前前端静态文件位于：

```text
web/src/
```

常见访问入口：

- 登录页：`/login.html`
- 后台总览：`/`
- 后台会在登录成功后自动进入设置页面
- 你也可以在前端进一步扩展配置、source、manual rule 管理

### 发布时要注意

如果你使用打包脚本生成发布包，运行目录中的配置通常会写成：

```json
"web_dir": "./web/src"
```

因此发布包必须包含 `web/src`，不能只放 `web/dist`。

---

## 构建与发布

项目当前提供两份 PowerShell 构建脚本：

- `scripts/build_windows_amd64.ps1`
- `scripts/build_linux_amd64.ps1`

### Windows amd64

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\build_windows_amd64.ps1 -Version v1.0.0
```

### Linux amd64 交叉编译

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\build_linux_amd64.ps1 -Version v1.0.0
```

### 发布包结构

构建后的目录大致如下：

```text
dist/
└── ros-address-list-tool_v1.0.0_linux_amd64/
    ├── ros-address-list-tool
    ├── config.json
    ├── README.md
    ├── LICENSE
    ├── RUN.txt
    ├── logs/
    ├── output/
    └── web/
        └── src/
            ├── app.css
            ├── app.js
            ├── favicon.svg
            ├── index.html
            └── login.html
```

### 构建脚本当前处理内容

- 先跑 `go test ./...`
- 再构建目标平台二进制
- 复制 `config.json`、`README.md`、`LICENSE`
- 复制 `web/src`
- 创建 `logs`、`output`
- 生成 `RUN.txt`
- 打包为 ZIP

---

## 部署建议

### 1. 本地开发

直接使用：

```bash
go run ./cmd/server -config ./configs/config.json -serve -listen 127.0.0.1:8090
```

### 2. 云服务器部署

推荐：

- 程序监听本机或内网地址
- 前面接 Nginx / Caddy / Traefik 做 HTTPS
- 只开放反向代理端口
- 自动化调用走 Token
- 浏览器后台走登录页 + Cookie 会话

### 3. 不建议的方式

- 直接把无 TLS 的服务裸露在公网
- 在 RouterOS 自动脚本里使用浏览器登录态
- 使用弱口令长期不修改

---

## 常见问题

### Q1：为什么浏览器能登录，但 RouterOS 脚本不该走登录页？

因为两种认证场景不同：

- 浏览器后台：适合人工管理，用登录页和会话 Cookie
- RouterOS 自动化：适合机器调用，用 `auth_token`

### Q2：为什么发布包必须包含 `web/src`？

因为当前服务配置使用的是：

```json
"web_dir": "./web/src"
```

如果你只打包 `web/dist`，服务启动后找不到页面。

### Q3：生产环境推荐 `replace_all` 还是 `diff`？

如果是云端集中生成 + RouterOS 定时拉取，优先推荐 `replace_all`。  
它更容易理解，也更容易验证。

### Q4：首次登录要求修改密码后，配置里还要不要保留明文密码？

不建议长期保留。  
首次初始化后应尽量改为哈希存储逻辑或清理明文初始密码。

---

## 开发与测试

### 运行测试

```bash
go test ./...
```

### 本地渲染验证

```bash
go run ./cmd/server -config ./configs/config.json
```

### 本地服务验证

```bash
go run ./cmd/server -config ./configs/config.json -serve
```

### 常用调试入口

- `GET /healthz`
- `GET /api/v1/config`
- `POST /api/v1/render`
- `GET /api/v1/render/latest.rsc`
- `GET /api/v1/auth/me`

---

## 许可证

MIT License
