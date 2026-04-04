# ros-address-list-tool

一个使用 Go 编写的 RouterOS address-list 管理工具。

当前已实现能力：

- 管理多个 RouterOS address-list
- 支持多来源输入（本地 JSON、远程 URL）
- 支持手工维护规则（追加、删除、优先级覆盖）
- 支持生成 RouterOS `.rsc` 脚本
- 支持 `replace_all` 与 `diff` 两种更新模式
- 支持 CLI 参数覆盖
- 支持 HTTP API
- 支持 address-list 管理
- 支持 address-list description 管理
- 支持 manual rule 管理
- 支持日志记录
- 为后续 Web 管理端预留扩展空间

---

## 当前项目结构

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
│       ├── server.go
│       ├── source.go
│       ├── store.go
│       ├── types.go
│       ├── validate.go
│       ├── validate_test.go
│       ├── source_test.go
│       ├── merge_test.go
│       └── render_test.go
├── logs/
├── output/
├── web/
├── .gitattributes
├── .gitignore
├── LICENSE
├── README.md
└── go.mod