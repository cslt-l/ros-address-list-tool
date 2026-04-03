// Package app 提供 RouterOS address-list 工具的核心业务层。
//
// 当前这个文件的作用只有两个：
// 1. 先把 internal/app 目录变成一个真实可追踪的 Go 包目录
// 2. 为后续逐步加入 types、validate、source、merge、render、engine、server 等文件预留位置
//
// 之所以在项目骨架阶段就创建这个包，是因为我们从一开始就要明确：
// cmd/server/main.go 只负责“程序启动和参数入口”，
// 真正的业务逻辑全部应该沉淀在 internal/app 里。
// 这样未来无论是 CLI、HTTP API，还是 Web 管理端，
// 都可以复用同一套核心逻辑，而不需要把业务代码写散。
package app
