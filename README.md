# ros-address-list-tool

一个使用 Go 编写的 RouterOS address-list 管理工具。

当前项目目标：

- 管理多个 RouterOS address-list
- 支持多来源输入（本地 JSON、远程 URL）
- 支持手工维护规则（追加、删除、优先级覆盖）
- 支持生成 RouterOS `.rsc` 脚本
- 支持 `replace_all` 与 `diff` 两种更新模式
- 支持 HTTP API
- 为后续 Web 管理端预留扩展空间

---

## 当前阶段

当前仓库处于**项目骨架搭建阶段**，已经包含：

- Go 项目结构
- 命令行启动入口
- 最小配置文件
- Git / GitHub 连接准备
- 后续核心业务目录预留

---

## 目录结构

```text
ros-address-list-tool/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   └── app/
│       └── doc.go
├── configs/
│   └── config.json
├── data/
├── output/
├── logs/
├── web/
├── scripts/
├── .gitignore
├── README.md
├── LICENSE
└── go.mod