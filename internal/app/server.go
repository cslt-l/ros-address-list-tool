package app

import (
	"encoding/json"
	"log"
	"net/http"
)

// NewHTTPHandler 创建整个 HTTP API 的根处理器。
// 当前阶段先只暴露最核心的接口：
// 1. GET  /healthz           健康检查
// 2. GET  /api/v1/config     读取当前配置
// 3. POST /api/v1/render     触发一次渲染
//
// 为什么现在只做这三个？
// 因为第 10 步的核心目标是：
// “让 serve 模式真正启动 HTTP 服务，并且打通最核心的调用链”。
// address-list 管理、description 管理、manual rule 管理，
// 会在后续第 11 步、第 12 步分别展开。
func NewHTTPHandler(store *ConfigStore, logger *log.Logger) http.Handler {
	s := &apiServer{
		store:  store,
		logger: logger,
	}

	mux := http.NewServeMux()

	// 健康检查接口。
	mux.HandleFunc("/healthz", s.handleHealthz)

	// 返回当前配置。
	mux.HandleFunc("/api/v1/config", s.handleConfig)

	// 执行渲染。
	mux.HandleFunc("/api/v1/render", s.handleRender)

	// 根路径给一个简单说明，方便你浏览器直接打开时知道服务是否活着。
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"message": "RouterOS address-list HTTP API is running",
			"routes": []string{
				"GET /healthz",
				"GET /api/v1/config",
				"POST /api/v1/render",
			},
		})
	})

	// 增加一层日志中间件，统一记录请求。
	return loggingMiddleware(logger, mux)
}

// apiServer 封装 HTTP 服务运行所需依赖。
// 当前只需要：
// 1. store：读取当前配置
// 2. logger：记录访问日志与渲染日志
type apiServer struct {
	store  *ConfigStore
	logger *log.Logger
}

// handleHealthz 用于最基础的健康检查。
// 这是服务化程序的标准接口之一。
// 后续如果你要接反向代理、监控、容器健康检查，这个接口会很有用。
func (s *apiServer) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// handleConfig 返回当前配置。
// 当前阶段先只支持 GET。
func (s *apiServer) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	cfg := s.store.GetConfig()
	writeJSON(w, http.StatusOK, cfg)
}

// renderRequest 表示 /api/v1/render 的请求体。
// 当前阶段允许调用方“临时覆盖本次渲染参数”，但不会修改磁盘配置。
// 这和 CLI 里的 -mode / -output 是一样的理念：
// 只影响本次执行，不回写 config.json。
type renderRequest struct {
	// Mode 表示本次请求临时指定渲染模式。
	Mode RenderMode `json:"mode,omitempty"`

	// OutputPath 表示本次请求临时指定输出文件路径。
	OutputPath string `json:"output_path,omitempty"`
}

// handleRender 触发一次渲染。
// 当前阶段只支持 POST。
//
// 处理流程：
// 1. 从 store 中取出当前配置副本
// 2. 解析请求体中的临时覆盖参数
// 3. 调用 Execute 执行完整渲染流程
// 4. 返回 JSON 结果
func (s *apiServer) handleRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	cfg := s.store.GetConfig()

	var req renderRequest

	// 允许空请求体：
	// - 空体时直接按当前配置执行
	// - 非空时再尝试解析 JSON
	if r.Body != nil {
		defer r.Body.Close()

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		// 这里的处理方式稍微宽松一点：
		// 如果请求体为空，不视为错误；
		// 如果请求体有内容但不是合法 JSON，则返回 400。
		if err := decoder.Decode(&req); err != nil {
			// 针对完全空 body 的情况，json.Decoder 也会报 EOF。
			// 这里把 EOF 当成“没有传参数”，继续走默认配置即可。
			if err.Error() != "EOF" {
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error": "invalid request body: " + err.Error(),
				})
				return
			}
		}
	}

	// 对本次执行应用临时覆盖，但不回写 store。
	if req.Mode != "" {
		cfg.Output.Mode = req.Mode
	}
	if req.OutputPath != "" {
		cfg.Output.Path = req.OutputPath
	}

	result, err := Execute(cfg, s.logger)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"mode":        result.Mode,
		"list_count":  result.ListCount,
		"entry_count": result.EntryCount,
		"output_path": result.OutputPath,
		"script":      result.Script,
	})
}

// loggingMiddleware 为每一个 HTTP 请求打日志。
// 这层日志和业务日志不同：
// - 业务日志记录“渲染流程内部发生了什么”
// - 这层日志记录“谁访问了哪个接口”
func loggingMiddleware(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("HTTP %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// writeJSON 统一输出 JSON 响应。
// 这样可以避免每个 handler 都重复写：
// - 设置 Content-Type
// - 设置状态码
// - JSON 编码
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
