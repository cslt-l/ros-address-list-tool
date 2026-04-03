# ros-address-list-tool

一个使用 Go 编写的 RouterOS address-list 管理工具。

项目目标：

- 管理多个 RouterOS address-list
- 支持多来源输入（本地 JSON、远程 URL）
- 支持手工维护规则（追加、删除、优先级覆盖）
- 支持生成 RouterOS `.rsc` 脚本
- 支持 `replace_all` 与 `diff` 两种更新模式
- 支持 HTTP API
- 支持 address-list 描述信息管理
- 为后续 Web 管理端预留扩展空间

---

## 当前开发阶段

当前仓库已完成：

- Git / GitHub 连接
- Go 项目骨架
- 正式核心配置模型
- 正式领域数据结构

下一步将实现：

- 配置校验
- 地址合法性检查
- 去重与规范化

---

## 目录结构

```text
ros-address-list-tool/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   └── app/
│       ├── config.go
│       ├── doc.go
│       └── types.go
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