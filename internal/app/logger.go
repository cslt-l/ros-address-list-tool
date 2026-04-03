package app

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

// NewLogger 创建一个同时输出到“标准输出 + 日志文件”的日志器。
//
// 为什么单独抽一个日志初始化函数？
// 因为日志属于“运行时基础设施”，不属于业务逻辑。
// 后续无论是：
// - CLI 单次执行
// - HTTP API 服务模式
// - Web 管理端后台
// 都应该复用同一套日志初始化逻辑。
//
// 参数：
// - logFile: 日志文件路径；如果为空，则只输出到终端
//
// 返回：
// - *log.Logger: 可直接使用的日志器
// - func() error: 关闭函数；如果打开了文件，需要在程序结束时调用
// - error: 初始化失败时返回错误
func NewLogger(logFile string) (*log.Logger, func() error, error) {
	// writers 用于汇总多个输出目标。
	// 当前我们默认至少输出到终端。
	var writers []io.Writer
	writers = append(writers, os.Stdout)

	// closeFn 默认为空实现。
	// 如果后面成功打开了文件，会被替换成真正的关闭函数。
	closeFn := func() error {
		return nil
	}

	// 如果配置了日志文件路径，则尝试创建目录并打开文件。
	if logFile != "" {
		dir := filepath.Dir(logFile)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, nil, err
			}
		}

		// 以“创建 + 追加 + 只写”模式打开日志文件。
		// 这样多次运行不会覆盖历史日志，而是持续追加。
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, nil, err
		}

		writers = append(writers, f)

		// 返回真正的关闭函数。
		closeFn = f.Close
	}

	// 使用 log.New 创建标准库日志器。
	// 当前格式选择：
	// - 前缀 [ros-list]
	// - 标准日期时间
	// - 微秒级时间
	// - 文件名和行号
	//
	// 这样在开发阶段排错会非常方便。
	logger := log.New(
		io.MultiWriter(writers...),
		"[ros-list] ",
		log.LstdFlags|log.Lmicroseconds|log.Lshortfile,
	)

	return logger, closeFn, nil
}
