# ros-address-list-tool

<div align="center">


![Go](https://img.shields.io/badge/Go-Project-00ADD8?logo=go&logoColor=white)
![RouterOS](https://img.shields.io/badge/RouterOS-Address--List-2D6CDF)
![License](https://img.shields.io/badge/License-MIT-green)
![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-blueviolet)

一个使用 **Go** 编写的 **RouterOS Address List 管理工具**。  
用于把多来源地址数据、手工规则、差异计算与 RouterOS `.rsc` 脚本生成整合到同一套可维护流程中。

</div>

---

## 目录

- [项目简介](#项目简介)
- [功能概览](#功能概览)
- [适用场景](#适用场景)
- [项目结构](#项目结构)
- [工作原理](#工作原理)
- [运行环境](#运行环境)
- [快速开始](#快速开始)
- [配置文件详解](#配置文件详解)
- [Source 详解](#source-详解)
- [Manual Rule 详解](#manual-rule-详解)
- [渲染模式详解](#渲染模式详解)
- [Web 管理界面](#web-管理界面)
- [HTTP API 使用说明](#http-api-使用说明)
- [推荐配置示例](#推荐配置示例)
- [RouterOS 导入方法](#routeros-导入方法)
- [发布与运维建议](#发布与运维建议)
- [常见问题排查](#常见问题排查)
- [开发与测试](#开发与测试)
- [路线建议](#路线建议)
- [许可证](#许可证)

---

## 项目简介

在 RouterOS 的多出口分流、运营商分流、国内外分流、PCC 负载均衡等场景中，经常需要维护大量 `address-list` 条目。

直接在 RouterOS 设备上手工维护会有这些问题：

- 列表量大，人工维护容易出错
- 多来源数据难以统一管理
- 缺少版本化和可回滚能力
- 手工修改无法沉淀为长期配置
- 批量替换、差异更新、Web 管理都不方便

本项目的目标就是把这些动作收敛为一个统一流程：

1. 定义要管理的 address-list
2. 从本地文件或远程 URL 加载地址源
3. 对地址做校验、规范化、去重
4. 叠加手工规则进行最终覆盖
5. 生成 RouterOS 可导入的 `.rsc` 脚本
6. 通过 CLI / HTTP API / Web 页面进行管理

你可以把它理解为：

> **“用配置文件和数据源，自动生成 RouterOS address-list 管理脚本。”**

---

## 功能概览

当前项目围绕 address-list 管理，提供了如下能力：

### 核心能力

- 管理多个 RouterOS address-list
- 支持 IPv4 / IPv6
- 支持多来源输入
    - 本地文件
    - 远程 URL
- 支持多种来源格式
    - `json`
    - `plain_cidr`
- 支持手工维护规则
    - `add`
    - `remove`
    - 优先级覆盖
- 支持地址校验、规范化、去重
- 支持生成 RouterOS `.rsc` 脚本

### 渲染能力

- `replace_all` 全量替换模式
- `diff` 差异更新模式

### 管理能力

- HTTP API
- Web 静态管理界面
- list 管理
- source 管理
- manual rule 管理
- 描述信息管理
- 配置持久化
- 日志输出

### 安全与稳定性建议

如果你使用的是已经完成本地修复的版本，建议具备如下行为：

- 默认监听本机地址
- 配置读取时脱敏敏感字段
- 渲染输出路径受限制
- URL source 仅允许合法公网 `http/https`
- 前端支持 API Token 输入

---

## 适用场景

本项目特别适合下面这些场景：

### 1. 多运营商出口分流

例如：

- 电信走电信出口
- 移动走移动出口
- 联通走联通出口

### 2. 国内 / 国外流量分流

例如：

- 国内地址直连
- 国际地址走代理或特定 WAN

### 3. PCC 负载均衡辅助列表

例如维护：

- 适合做 PCC 的目标地址池
- 特定业务地址候选集合

### 4. 自动化生成 RouterOS Address List

例如：

- 从远程运营商 IP 库自动抓取
- 本地修订后自动生成 `.rsc`
- 再导入 RouterOS

### 5. 通过 Web/API 管理策略

例如：

- 在浏览器里管理 sources / lists / rules
- API 驱动渲染与输出
- 为后续自动化平台集成做准备

---

## 项目结构

```text
ros-address-list-tool/
├── cmd/
│   └── server/
│       └── main.go
├── configs/
│   └── config.json
├── data/
│   ├── desired-1.json
│   ├── desired-2.json
│   └── current.json
├── internal/
│   └── app/
│       ├── config.go
│       ├── engine.go
│       ├── logger.go
│       ├── merge.go
│       ├── render.go
│       ├── security.go
│       ├── server.go
│       ├── source.go
│       ├── store.go
│       ├── types.go
│       ├── validate.go
│       ├── *_test.go
├── logs/
├── output/
├── web/
│   └── dist/
├── .gitignore
├── LICENSE
├── README.md
└── go.mod
```

### 目录说明

| 路径            | 说明                                                 |
| --------------- | ---------------------------------------------------- |
| `cmd/server/`   | 程序入口，负责参数解析、配置加载、启动服务或执行渲染 |
| `configs/`      | 配置文件目录                                         |
| `data/`         | 本地测试 / 示例数据                                  |
| `internal/app/` | 核心业务逻辑目录                                     |
| `logs/`         | 日志输出目录                                         |
| `output/`       | 生成的 `.rsc` 脚本输出目录                           |
| `web/dist/`     | 前端静态页面目录                                     |

---

## 工作原理

程序完整流程如下：

```text
配置文件
   ↓
lists + desired_sources + current_state_sources + manual_rules
   ↓
校验 / 规范化 / 去重 / 合并
   ↓
生成 desired snapshot / current snapshot
   ↓
按 replace_all 或 diff 渲染
   ↓
输出 RouterOS .rsc
```

---

## 运行环境

### 基础要求

- Go 版本与项目 `go.mod` 保持一致
- Windows / Linux / macOS 任一平台
- 若使用 URL source，需要网络可访问对应地址源
- 若要实际导入结果，需要一台 RouterOS 设备

### 推荐环境

- Windows + PowerShell
- Go
- Git
- 现代浏览器
- 一台测试用 RouterOS

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

### 3. 以服务模式启动

```bash
go run ./cmd/server -config ./configs/config.json -serve -listen 127.0.0.1:8090
```

启动后访问：

```text
http://127.0.0.1:8090/
```

### 4. 仅执行一次渲染

```bash
go run ./cmd/server -config ./configs/config.json
```

程序会读取配置，加载 source，合并规则，并把最终脚本写到 `output.path` 指定的位置。

---

## 配置文件详解

一个完整配置大致如下：

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
    "web_dir": "./web/dist",
    "auth_token": ""
  }
}
```

### 顶层字段说明

#### `auto_create_lists`

是否允许在 source 或 manual rule 中出现未定义 list 时自动创建。

- `true`：允许自动创建
- `false`：必须先在 `lists` 中定义

**建议生产环境使用 `false`。**

---

#### `log_file`

日志文件路径，例如：

```json
"log_file": "./logs/app.log"
```

---

#### `lists`

预定义 address-list。

单个 list 结构示例：

```json
{
  "name": "toWanTelecom4",
  "family": "ipv4",
  "enabled": true,
  "description": "中国电信 IPv4"
}
```

字段说明：

| 字段          | 说明             |
| ------------- | ---------------- |
| `name`        | list 名称        |
| `family`      | `ipv4` 或 `ipv6` |
| `enabled`     | 是否启用         |
| `description` | 描述信息         |

> 注意：**一个 list 只能对应一个地址族**。  
> 因此 IPv4 和 IPv6 不要共用同一个 list 名。

例如：

- `toWanTelecom4`
- `toWanTelecom6`

而不要把二者都写成 `toWanTelecom`。

---

#### `desired_sources`

目标状态来源。  
这些来源会被加载、合并，形成最终想要下发到 RouterOS 的数据。

---

#### `current_state_sources`

当前状态来源。  
主要用于 `diff` 模式，告诉程序当前已有的 address-list 内容是什么。

---

#### `manual_rules`

手工规则。  
用于对 source 加载结果做最后覆盖：

- 强制补充
- 强制移除
- 高优先级覆盖

---

#### `output`

输出配置。

| 字段              | 说明                             |
| ----------------- | -------------------------------- |
| `path`            | 默认输出路径                     |
| `mode`            | `replace_all` 或 `diff`          |
| `managed_comment` | 程序生成条目时统一写入的 comment |

---

#### `server`

服务配置。

| 字段         | 说明              |
| ------------ | ----------------- |
| `listen`     | 服务监听地址      |
| `enable_web` | 是否启用静态 Web  |
| `web_dir`    | 前端目录          |
| `auth_token` | API Token，可为空 |

---

## Source 详解

本项目支持两类 source：

- `file`
- `url`

支持两种数据格式：

- `json`
- `plain_cidr`

---

### `file` 类型 source

适用于：

- 本地调试
- 离线环境
- 手工准备好的数据文件

示例：

```json
{
  "name": "desired-local-json",
  "type": "file",
  "path": "./data/desired-1.json",
  "format": "json",
  "enabled": true,
  "priority": 100,
  "timeout_seconds": 15
}
```

---

### `url` 类型 source

适用于：

- 运营商 IP 库
- 远程更新源
- 统一集中维护的地址数据

示例：

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

---

### `json` 格式

适合一个 source 中同时携带多个 list。

示例：

```json
{
  "lists": [
    {
      "name": "toWanTelecom4",
      "family": "ipv4",
      "description": "中国电信 IPv4",
      "entries": [
        "1.0.1.0/24",
        "1.0.2.0/23"
      ]
    },
    {
      "name": "toWanMobile4",
      "family": "ipv4",
      "description": "中国移动 IPv4",
      "entries": [
        "36.128.0.0/10"
      ]
    }
  ]
}
```

---

### `plain_cidr` 格式

适合每行一个 IP / CIDR 的纯文本源。

例如：

```text
1.0.1.0/24
1.0.2.0/23
36.128.0.0/10
240e::/20
```

这类 source 必须额外指定：

- `target_list_name`
- `target_list_family`

因为纯文本本身不带 list 元信息。

---

### `line_comment_prefixes`

用于 `plain_cidr` 格式，告诉程序哪些前缀是注释行。

例如：

```json
"line_comment_prefixes": ["#", "//", ";"]
```

那么下面这些行会被跳过：

```text
# comment
// comment
; comment
```

---

## Manual Rule 详解

Manual Rule 用于对 source 结果做最终人工覆盖。

### 基本结构

```json
{
  "id": "manual-add-telecom-extra",
  "list_name": "toWanTelecom4",
  "action": "add",
  "priority": 200,
  "enabled": true,
  "description": "补充遗漏地址",
  "entries": [
    "101.1.2.0/24"
  ]
}
```

### 字段说明

| 字段          | 说明                 |
| ------------- | -------------------- |
| `id`          | 规则唯一 ID          |
| `list_name`   | 作用到哪个 list      |
| `action`      | `add` 或 `remove`    |
| `priority`    | 优先级，数值越大越强 |
| `enabled`     | 是否启用             |
| `description` | 描述                 |
| `entries`     | 规则中涉及的地址     |

### 典型用法

#### 1. 手工补充地址

```json
{
  "id": "manual-add-telecom-extra",
  "list_name": "toWanTelecom4",
  "action": "add",
  "priority": 200,
  "enabled": true,
  "description": "补充电信遗漏地址",
  "entries": [
    "101.1.2.0/24"
  ]
}
```

#### 2. 强制移除地址

```json
{
  "id": "manual-remove-wrong-entry",
  "list_name": "toWanTelecom4",
  "action": "remove",
  "priority": 300,
  "enabled": true,
  "description": "剔除误命中地址",
  "entries": [
    "1.0.1.0/24"
  ]
}
```

### 优先级规则

- 优先级越高，覆盖能力越强
- 一般建议手工规则优先级高于普通 source

---

## 渲染模式详解

支持两种模式：

- `replace_all`
- `diff`

---

### `replace_all`

最简单、最稳定的方式。

程序会对每个受管 list：

1. 删除程序管理的旧项
2. 再把目标快照里的全部条目重新添加

示例输出：

```rsc
/ip firewall address-list
remove [find where list="toWanTelecom4" comment="managed-by-go"]
add list=toWanTelecom4 address=1.0.1.0/24 comment="managed-by-go"
add list=toWanTelecom4 address=1.0.2.0/23 comment="managed-by-go"
```

适合：

- 第一次部署
- 追求流程简单稳定
- 不依赖 current source

---

### `diff`

差异更新模式。

程序会比较：

- 目标快照
- 当前快照

然后只输出新增和删除项。

优点：

- 输出更精简
- 变更更小

缺点：

- 需要可靠的 `current_state_sources`
- 如果 current source 不准确，diff 结果也会不准确

适合：

- 已经在线运行一段时间
- 想减少不必要的全量替换

---

## Web 管理界面

如果配置里启用了：

```json
"server": {
  "listen": "127.0.0.1:8090",
  "enable_web": true,
  "web_dir": "./web/dist"
}
```

启动服务后访问：

```text
http://127.0.0.1:8090/
```

### 页面通常可管理的内容

- 当前配置查看
- Lists 管理
- Desired Sources 管理
- Current Sources 管理
- Manual Rules 管理
- 触发 source 测试
- 触发 render

### Token 使用方式

如果你的前端版本已经接入 API Token：

- 页面右上角可以填写 Token
- 保存后会自动带到后续 API 请求中
- 适合远程访问场景

---

## HTTP API 使用说明

> 接口前缀：`/api/v1`

---

### 健康检查

#### `GET /healthz`

用途：

- 确认服务是否存活

示例：

```bash
curl http://127.0.0.1:8090/healthz
```

---

### 配置读取

#### `GET /api/v1/config`

用途：

- 返回当前配置（建议为脱敏版本）

示例：

```bash
curl http://127.0.0.1:8090/api/v1/config
```

---

### 渲染脚本

#### `POST /api/v1/render`

用途：

- 执行一次渲染
- 返回渲染结果
- 可选指定输出路径

示例：

```bash
curl -X POST http://127.0.0.1:8090/api/v1/render \
  -H "Content-Type: application/json" \
  -d '{}'
```

如果支持输出路径覆盖：

```bash
curl -X POST http://127.0.0.1:8090/api/v1/render \
  -H "Content-Type: application/json" \
  -d '{"output_path":"./output/custom.rsc"}'
```

---

### Lists 管理

#### `GET /api/v1/lists`

```bash
curl http://127.0.0.1:8090/api/v1/lists
```

#### `POST /api/v1/lists`

```bash
curl -X POST http://127.0.0.1:8090/api/v1/lists \
  -H "Content-Type: application/json" \
  -d '{
    "name":"toWanTelecom4",
    "family":"ipv4",
    "enabled":true,
    "description":"中国电信 IPv4"
  }'
```

#### `PUT /api/v1/lists/{name}`

```bash
curl -X PUT http://127.0.0.1:8090/api/v1/lists/toWanTelecom4 \
  -H "Content-Type: application/json" \
  -d '{
    "name":"toWanTelecom4",
    "family":"ipv4",
    "enabled":true,
    "description":"更新后的描述"
  }'
```

#### `DELETE /api/v1/lists/{name}`

```bash
curl -X DELETE http://127.0.0.1:8090/api/v1/lists/toWanTelecom4
```

---

### Manual Rules 管理

#### `GET /api/v1/manual-rules`

```bash
curl http://127.0.0.1:8090/api/v1/manual-rules
```

#### `POST /api/v1/manual-rules`

```bash
curl -X POST http://127.0.0.1:8090/api/v1/manual-rules \
  -H "Content-Type: application/json" \
  -d '{
    "id":"manual-add-telecom-extra",
    "list_name":"toWanTelecom4",
    "action":"add",
    "priority":200,
    "enabled":true,
    "description":"补充遗漏地址",
    "entries":["101.1.2.0/24"]
  }'
```

---

### Desired Sources 管理

#### `GET /api/v1/sources/desired`

```bash
curl http://127.0.0.1:8090/api/v1/sources/desired
```

#### `POST /api/v1/sources/desired`

```bash
curl -X POST http://127.0.0.1:8090/api/v1/sources/desired \
  -H "Content-Type: application/json" \
  -d '{
    "name":"china-telecom-ipv4",
    "type":"url",
    "url":"https://china-operator-ip.yfgao.com/chinanet.txt",
    "format":"plain_cidr",
    "target_list_name":"toWanTelecom4",
    "target_list_family":"ipv4",
    "line_comment_prefixes":["#","//",";"],
    "enabled":true,
    "priority":100,
    "timeout_seconds":30
  }'
```

---

### Current Sources 管理

#### `GET /api/v1/sources/current`

```bash
curl http://127.0.0.1:8090/api/v1/sources/current
```

---

### Source 测试

#### `POST /api/v1/sources/test`

```bash
curl -X POST http://127.0.0.1:8090/api/v1/sources/test \
  -H "Content-Type: application/json" \
  -d '{
    "name":"china-telecom-ipv4",
    "type":"url",
    "url":"https://china-operator-ip.yfgao.com/chinanet.txt",
    "format":"plain_cidr",
    "target_list_name":"toWanTelecom4",
    "target_list_family":"ipv4",
    "line_comment_prefixes":["#","//",";"],
    "enabled":true,
    "priority":100,
    "timeout_seconds":30
  }'
```

---

## 推荐配置示例

### 中国电信 / 移动 / 联通 IPv4 + IPv6

```json
{
  "auto_create_lists": false,
  "log_file": "./logs/app.log",
  "lists": [
    {
      "name": "toWanTelecom4",
      "family": "ipv4",
      "enabled": true,
      "description": "中国电信 IPv4"
    },
    {
      "name": "toWanTelecom6",
      "family": "ipv6",
      "enabled": true,
      "description": "中国电信 IPv6"
    },
    {
      "name": "toWanMobile4",
      "family": "ipv4",
      "enabled": true,
      "description": "中国移动 IPv4"
    },
    {
      "name": "toWanMobile6",
      "family": "ipv6",
      "enabled": true,
      "description": "中国移动 IPv6"
    },
    {
      "name": "toWanUnicom4",
      "family": "ipv4",
      "enabled": true,
      "description": "中国联通 IPv4"
    },
    {
      "name": "toWanUnicom6",
      "family": "ipv6",
      "enabled": true,
      "description": "中国联通 IPv6"
    },
    {
      "name": "toWanGlobal4",
      "family": "ipv4",
      "enabled": true,
      "description": "默认国际/未知 IPv4"
    },
    {
      "name": "toWanGlobal6",
      "family": "ipv6",
      "enabled": true,
      "description": "默认国际/未知 IPv6"
    },
    {
      "name": "toPccBalance4",
      "family": "ipv4",
      "enabled": true,
      "description": "PCC 负载均衡 IPv4"
    }
  ],
  "desired_sources": [
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
    },
    {
      "name": "china-telecom-ipv6",
      "type": "url",
      "url": "https://china-operator-ip.yfgao.com/chinanet6.txt",
      "format": "plain_cidr",
      "target_list_name": "toWanTelecom6",
      "target_list_family": "ipv6",
      "line_comment_prefixes": ["#", "//", ";"],
      "enabled": true,
      "priority": 100,
      "timeout_seconds": 30
    },
    {
      "name": "china-mobile-ipv4",
      "type": "url",
      "url": "https://china-operator-ip.yfgao.com/cmcc.txt",
      "format": "plain_cidr",
      "target_list_name": "toWanMobile4",
      "target_list_family": "ipv4",
      "line_comment_prefixes": ["#", "//", ";"],
      "enabled": true,
      "priority": 100,
      "timeout_seconds": 30
    },
    {
      "name": "china-mobile-ipv6",
      "type": "url",
      "url": "https://china-operator-ip.yfgao.com/cmcc6.txt",
      "format": "plain_cidr",
      "target_list_name": "toWanMobile6",
      "target_list_family": "ipv6",
      "line_comment_prefixes": ["#", "//", ";"],
      "enabled": true,
      "priority": 100,
      "timeout_seconds": 30
    },
    {
      "name": "china-unicom-ipv4",
      "type": "url",
      "url": "https://china-operator-ip.yfgao.com/unicom.txt",
      "format": "plain_cidr",
      "target_list_name": "toWanUnicom4",
      "target_list_family": "ipv4",
      "line_comment_prefixes": ["#", "//", ";"],
      "enabled": true,
      "priority": 100,
      "timeout_seconds": 30
    },
    {
      "name": "china-unicom-ipv6",
      "type": "url",
      "url": "https://china-operator-ip.yfgao.com/unicom6.txt",
      "format": "plain_cidr",
      "target_list_name": "toWanUnicom6",
      "target_list_family": "ipv6",
      "line_comment_prefixes": ["#", "//", ";"],
      "enabled": true,
      "priority": 100,
      "timeout_seconds": 30
    }
  ],
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
    "web_dir": "./web/dist"
  }
}
```

---

## RouterOS 导入方法

### 方法 1：通过 WinBox / WebFig 上传导入

把生成的 `.rsc` 上传到 RouterOS，然后执行：

```rsc
/import file-name=routeros-address-list.rsc
```

### 方法 2：先人工检查，再导入

强烈建议先检查：

- list 名称是否正确
- IPv4 / IPv6 是否混淆
- comment 是否正常
- 是否有乱码或异常字符

---

## 发布与运维建议

### 推荐做法

- 生产环境使用 `auto_create_lists=false`
- 初期优先使用 `replace_all`
- 服务默认只监听本机
- 如果开放到局域网，配置 API Token
- 输出目录单独隔离
- 先在测试 RouterOS 导入，再上生产

### 命名建议

建议统一采用带地址族后缀的风格：

- `toWanTelecom4`
- `toWanTelecom6`
- `toWanMobile4`
- `toWanMobile6`
- `toWanUnicom4`
- `toWanUnicom6`
- `toWanGlobal4`
- `toWanGlobal6`
- `toPccBalance4`

---

## 常见问题排查

### 1. `go test ./...` 失败

建议先定位是哪个包失败：

```bash
go test ./internal/app -v
```

常见原因：

- 某个测试文件 import 未使用
- 某个结构体字段还没同步
- 某个修补逻辑尚未落到本地文件

---

### 2. 服务启动失败

常见原因：

- 端口被占用
- 配置校验失败
- 输出目录无权限
- 日志目录无权限

---

### 3. URL Source 拉取失败

常见原因：

- URL 不是 `http/https`
- URL 指向私网 / 回环地址
- 远端超时
- 远端内容不是合法 `json` 或 `plain_cidr`
- `target_list_name` / `target_list_family` 未配置正确

---

### 4. 渲染失败

常见原因：

- source 中有非法地址
- 地址族和 list family 不匹配
- `managed_comment` 含非法字符
- `diff` 模式下 current source 不正确

---

### 5. RouterOS 导入失败

常见原因：

- `.rsc` 语法错误
- 生成结果被手工修改破坏
- IPv4 / IPv6 section 写错

---

## 开发与测试

### PowerShell 常用命令

#### 代码格式化

```powershell
gofmt -w .\cmd\server\main.go
Get-ChildItem .\internal\app\*.go | ForEach-Object { gofmt -w $_.FullName }
```

#### 运行测试

```powershell
go test ./...
```

#### 启动服务

```powershell
go run ./cmd/server -config .\configs\config.json -serve -listen 127.0.0.1:8090
```

#### Git 检查

```powershell
git status
git diff
```

### 推荐发布前检查清单

- `go test ./...` 全部通过
- 服务可启动
- 前端可正常打开
- 配置可正常读取
- Source 测试通过
- Render 成功输出 `.rsc`
- RouterOS 测试导入通过
- `git diff` 干净且无敏感信息

---

## 路线建议

如果你计划继续扩展这个项目，推荐方向如下：

- 增加当前 RouterOS 状态采集器
- 增加定时拉取 / 定时渲染
- 增加执行历史与审计日志
- 增加配置版本回滚
- 增加更细粒度的用户鉴权
- 增加前端表单校验与提示
- 增加一键导出 / 导入配置
- 增加直接推送到 RouterOS 的能力

---

## 截图占位

你可以在仓库里补充以下截图后，把这里替换掉：

### Web 首页

```text
docs/screenshots/home.png
```

### Sources 管理页

```text
docs/screenshots/sources.png
```

### Render 结果页

```text
docs/screenshots/render.png
```

如果你后续准备做 GitHub 展示，建议在仓库新增：

```text
docs/
└── screenshots/
    ├── home.png
    ├── sources.png
    └── render.png
```

---

## 许可证

MIT License

---

## 致谢

项目中引用的IP信息均来自 https://github.com/gaoyifan/china-operator-ip ，在此提出特别感谢。

这个项目的核心价值在于：  
把 RouterOS address-list 的维护，从“设备上手工改”变成“配置驱动、可复用、可测试、可版本化”的工程流程。

如果你正在做：

- 多 WAN 分流
- 运营商分流
- 国内外分流
- RouterOS 地址库自动化管理

这个工具会非常适合作为你的基础设施之一。

