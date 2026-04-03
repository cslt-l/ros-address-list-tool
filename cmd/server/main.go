package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// bootstrapConfig 是“项目骨架阶段”的最小配置结构体。
// 这里故意不把完整业务配置一次性全部写进来，
// 因为当前我们还处在“搭项目骨架”的阶段，目标是：
// 1. 先验证配置文件能够被程序读取
// 2. 先验证命令行入口能够正常工作
// 3. 先验证项目目录结构已经具备后续扩展条件
//
// 到下一步真正开始实现业务内核时，
// 这个最小结构体会被 internal/app/types.go 中的正式配置模型替代。
type bootstrapConfig struct {
	// LogFile 表示日志文件位置。
	// 当前阶段我们只读取并打印，不真正初始化日志系统。
	LogFile string `json:"log_file"`

	// Output 表示输出文件相关的最小配置。
	Output struct {
		// Path 表示未来生成 RouterOS .rsc 脚本的默认输出位置。
		Path string `json:"path"`
	} `json:"output"`

	// Server 表示 HTTP 服务相关的最小配置。
	Server struct {
		// Listen 表示未来 HTTP API 的监听地址。
		Listen string `json:"listen"`
	} `json:"server"`
}

func main() {
	// configPath 是命令行参数，用于指定配置文件路径。
	// 现在之所以先做成参数，而不是写死路径，
	// 是因为后续我们会支持：
	// 1. 不同环境使用不同配置
	// 2. CLI 和 HTTP 服务共用一个入口
	// 3. 测试时用单独的配置文件
	configPath := flag.String("config", "./configs/config.json", "配置文件路径")
	flag.Parse()

	// 打印启动提示，帮助我们确认程序已经真正进入 main 逻辑。
	fmt.Println("========================================")
	fmt.Println("RouterOS Address List Tool - 骨架启动阶段")
	fmt.Println("========================================")

	// 输出当前工作目录，方便排查“相对路径为什么找不到文件”这类问题。
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取当前工作目录失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("当前工作目录: %s\n", wd)

	// 将配置路径转成绝对路径，方便调试。
	absConfigPath, err := filepath.Abs(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析配置文件绝对路径失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("准备读取配置文件: %s\n", absConfigPath)

	// 读取配置文件原始内容。
	data, err := os.ReadFile(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取配置文件失败: %v\n", err)
		os.Exit(1)
	}

	// 将配置文件解析到最小配置结构体中。
	// 当前阶段的目的不是实现完整业务逻辑，
	// 而是验证“项目骨架 + 配置文件 + 命令行入口”已经打通。
	var cfg bootstrapConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "解析配置文件失败: %v\n", err)
		os.Exit(1)
	}

	// 打印解析结果，作为当前阶段的成功标志。
	fmt.Println()
	fmt.Println("配置文件读取成功。当前骨架信息如下：")
	fmt.Printf("- 日志文件路径: %s\n", cfg.LogFile)
	fmt.Printf("- 默认输出路径: %s\n", cfg.Output.Path)
	fmt.Printf("- 未来 HTTP 监听地址: %s\n", cfg.Server.Listen)

	fmt.Println()
	fmt.Println("项目骨架启动成功。")
	fmt.Println("下一步我们将开始实现正式的数据结构、校验逻辑和渲染引擎。")
}
