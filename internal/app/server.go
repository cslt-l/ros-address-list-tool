package app

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// NewHTTPHandler 创建整个 HTTP API 的根处理器。
// 当前阶段接口分为三组：
//
// 第一组：基础服务接口
// - GET  /healthz
// - GET  /api/v1/config
// - POST /api/v1/render
//
// 第二组：address-list 管理接口
// - GET    /api/v1/lists
// - POST   /api/v1/lists
// - GET    /api/v1/lists/{name}
// - PUT    /api/v1/lists/{name}
// - DELETE /api/v1/lists/{name}
// - GET    /api/v1/lists/{name}/description
// - PUT    /api/v1/lists/{name}/description
//
// 注意：
// manual rule 管理接口先不在这一步做，留到下一步。
func NewHTTPHandler(store *ConfigStore, logger *log.Logger) http.Handler {
	s := &apiServer{
		store:  store,
		logger: logger,
	}

	mux := http.NewServeMux()

	// 基础接口
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/api/v1/config", s.handleConfig)
	mux.HandleFunc("/api/v1/render", s.handleRender)

	// address-list 接口
	mux.HandleFunc("/api/v1/lists", s.handleLists)
	mux.HandleFunc("/api/v1/lists/", s.handleListByName)

	// 根路径说明
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"message": "RouterOS address-list HTTP API is running",
			"routes": []string{
				"GET  /healthz",
				"GET  /api/v1/config",
				"POST /api/v1/render",
				"GET  /api/v1/lists",
				"POST /api/v1/lists",
				"GET  /api/v1/lists/{name}",
				"PUT  /api/v1/lists/{name}",
				"DELETE /api/v1/lists/{name}",
				"GET  /api/v1/lists/{name}/description",
				"PUT  /api/v1/lists/{name}/description",
			},
		})
	})

	return loggingMiddleware(logger, mux)
}

// apiServer 封装 HTTP 服务所需依赖。
type apiServer struct {
	store  *ConfigStore
	logger *log.Logger
}

// handleHealthz 健康检查接口。
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

// handleConfig 返回当前完整配置。
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
type renderRequest struct {
	Mode       RenderMode `json:"mode,omitempty"`
	OutputPath string     `json:"output_path,omitempty"`
}

// handleRender 触发一次渲染。
func (s *apiServer) handleRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	cfg := s.store.GetConfig()

	var req renderRequest

	if r.Body != nil {
		defer r.Body.Close()

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			if err.Error() != "EOF" {
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error": "invalid request body: " + err.Error(),
				})
				return
			}
		}
	}

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

// handleLists 处理：
// - GET  /api/v1/lists   查看所有 list
// - POST /api/v1/lists   新增或更新一个 list
func (s *apiServer) handleLists(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := s.store.GetConfig()
		writeJSON(w, http.StatusOK, cfg.Lists)

	case http.MethodPost:
		var def ListDefinition
		if err := json.NewDecoder(r.Body).Decode(&def); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid request body: " + err.Error(),
			})
			return
		}

		if err := s.store.UpsertList(def); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "list upserted",
		})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
	}
}

// handleListByName 处理带 list 名称的接口：
// - GET    /api/v1/lists/{name}
// - PUT    /api/v1/lists/{name}
// - DELETE /api/v1/lists/{name}
// - GET    /api/v1/lists/{name}/description
// - PUT    /api/v1/lists/{name}/description
func (s *apiServer) handleListByName(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/lists/")
	if path == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "not found",
		})
		return
	}

	parts := strings.Split(path, "/")
	name, err := url.PathUnescape(parts[0])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "bad list name",
		})
		return
	}

	// /api/v1/lists/{name}/description
	if len(parts) == 2 && parts[1] == "description" {
		s.handleListDescription(w, r, name)
		return
	}

	switch r.Method {
	case http.MethodGet:
		cfg := s.store.GetConfig()
		for _, item := range cfg.Lists {
			if item.Name == name {
				writeJSON(w, http.StatusOK, item)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "list not found",
		})

	case http.MethodPut:
		var def ListDefinition
		if err := json.NewDecoder(r.Body).Decode(&def); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid request body: " + err.Error(),
			})
			return
		}

		// 路径参数优先，避免请求体里的 name 和 URL 不一致。
		def.Name = name

		if err := s.store.UpsertList(def); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "list updated",
		})

	case http.MethodDelete:
		if err := s.store.DeleteList(name); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "list deleted",
		})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
	}
}

// handleListDescription 专门处理 description 的单字段接口：
// - GET /api/v1/lists/{name}/description
// - PUT /api/v1/lists/{name}/description
//
// 之所以单独拆出来，是因为“单字段更新”在未来 Web 管理端里会非常常见。
func (s *apiServer) handleListDescription(w http.ResponseWriter, r *http.Request, name string) {
	switch r.Method {
	case http.MethodGet:
		cfg := s.store.GetConfig()
		for _, item := range cfg.Lists {
			if item.Name == name {
				writeJSON(w, http.StatusOK, map[string]string{
					"name":        item.Name,
					"description": item.Description,
				})
				return
			}
		}

		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "list not found",
		})

	case http.MethodPut:
		var body struct {
			Description string `json:"description"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid request body: " + err.Error(),
			})
			return
		}

		if err := s.store.SetListDescription(name, body.Description); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "description updated",
		})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
	}
}

// loggingMiddleware 为每个 HTTP 请求记录访问日志。
func loggingMiddleware(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("HTTP %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// writeJSON 统一输出 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
