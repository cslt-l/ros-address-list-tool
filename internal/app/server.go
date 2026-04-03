package app

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// NewHTTPHandler 创建整个 HTTP API 的根处理器。
// 当前阶段接口分为四组：
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
// 第三组：manual rule 管理接口
// - GET    /api/v1/manual-rules
// - POST   /api/v1/manual-rules
// - PUT    /api/v1/manual-rules/{id}
// - DELETE /api/v1/manual-rules/{id}
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

	// manual rule 接口
	mux.HandleFunc("/api/v1/manual-rules", s.handleManualRules)
	mux.HandleFunc("/api/v1/manual-rules/", s.handleManualRuleByID)

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
				"GET  /api/v1/manual-rules",
				"POST /api/v1/manual-rules",
				"PUT  /api/v1/manual-rules/{id}",
				"DELETE /api/v1/manual-rules/{id}",
			},
		})
	})

	return loggingMiddleware(logger, mux)
}

type apiServer struct {
	store  *ConfigStore
	logger *log.Logger
}

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

type renderRequest struct {
	Mode       RenderMode `json:"mode,omitempty"`
	OutputPath string     `json:"output_path,omitempty"`
}

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

// handleManualRules 处理：
// - GET  /api/v1/manual-rules
// - POST /api/v1/manual-rules
func (s *apiServer) handleManualRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := s.store.GetConfig()
		writeJSON(w, http.StatusOK, cfg.ManualRules)

	case http.MethodPost:
		var rule ManualRule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid request body: " + err.Error(),
			})
			return
		}

		if err := s.store.UpsertManualRule(rule); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "manual rule upserted",
		})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
	}
}

// handleManualRuleByID 处理：
// - PUT    /api/v1/manual-rules/{id}
// - DELETE /api/v1/manual-rules/{id}
func (s *apiServer) handleManualRuleByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/manual-rules/")
	if id == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "not found",
		})
		return
	}

	id, err := url.PathUnescape(id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "bad rule id",
		})
		return
	}

	switch r.Method {
	case http.MethodPut:
		var rule ManualRule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid request body: " + err.Error(),
			})
			return
		}

		// 路径参数优先
		rule.ID = id

		if err := s.store.UpsertManualRule(rule); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "manual rule updated",
		})

	case http.MethodDelete:
		if err := s.store.DeleteManualRule(id); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "manual rule deleted",
		})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
	}
}

func loggingMiddleware(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("HTTP %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
