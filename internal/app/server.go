package app

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// NewHTTPHandler 创建整个 HTTP API 与静态页面的根处理器。
func NewHTTPHandler(store *ConfigStore, logger *log.Logger) http.Handler {
	s := &apiServer{
		store:  store,
		logger: logger,
	}

	mux := http.NewServeMux()

	// ===================== API 接口 =====================

	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/api/v1/config", s.handleConfig)
	mux.HandleFunc("/api/v1/render", s.handleRender)

	mux.HandleFunc("/api/v1/lists", s.handleLists)
	mux.HandleFunc("/api/v1/lists/", s.handleListByName)

	mux.HandleFunc("/api/v1/manual-rules", s.handleManualRules)
	mux.HandleFunc("/api/v1/manual-rules/", s.handleManualRuleByID)

	// 新增：sources 管理接口
	mux.HandleFunc("/api/v1/sources/desired", s.handleDesiredSources)
	mux.HandleFunc("/api/v1/sources/desired/", s.handleDesiredSourceByName)
	mux.HandleFunc("/api/v1/sources/current", s.handleCurrentSources)
	mux.HandleFunc("/api/v1/sources/current/", s.handleCurrentSourceByName)

	mux.HandleFunc("/api/v1/sources/test", s.handleSourceTest)

	// ===================== 静态文件托管 =====================

	cfg := store.GetConfig()

	if cfg.Server.EnableWeb && dirExists(cfg.Server.WebDir) {
		fileServer := http.FileServer(http.Dir(cfg.Server.WebDir))
		mux.Handle("/", fileServer)
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{
				"message":     "RouterOS address-list HTTP API is running",
				"web_enabled": cfg.Server.EnableWeb,
				"web_dir":     cfg.Server.WebDir,
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

					"GET  /api/v1/sources/desired",
					"POST /api/v1/sources/desired",
					"PUT  /api/v1/sources/desired/{name}",
					"DELETE /api/v1/sources/desired/{name}",

					"GET  /api/v1/sources/current",
					"POST /api/v1/sources/current",
					"PUT  /api/v1/sources/current/{name}",
					"DELETE /api/v1/sources/current/{name}",
					"POST /api/v1/sources/test",
				},
			})
		})
	}

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

// ===================== Sources 管理接口 =====================

func (s *apiServer) handleDesiredSources(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := s.store.GetConfig()
		writeJSON(w, http.StatusOK, cfg.DesiredSources)

	case http.MethodPost:
		var src SourceConfig
		if err := json.NewDecoder(r.Body).Decode(&src); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid request body: " + err.Error(),
			})
			return
		}

		if err := s.store.UpsertDesiredSource(src); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "desired source upserted",
		})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
	}
}

func (s *apiServer) handleDesiredSourceByName(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/v1/sources/desired/")
	if name == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "not found",
		})
		return
	}

	name, err := url.PathUnescape(name)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "bad source name",
		})
		return
	}

	switch r.Method {
	case http.MethodPut:
		var src SourceConfig
		if err := json.NewDecoder(r.Body).Decode(&src); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid request body: " + err.Error(),
			})
			return
		}

		src.Name = name

		if err := s.store.UpsertDesiredSource(src); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "desired source updated",
		})

	case http.MethodDelete:
		if err := s.store.DeleteDesiredSource(name); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "desired source deleted",
		})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
	}
}

func (s *apiServer) handleCurrentSources(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := s.store.GetConfig()
		writeJSON(w, http.StatusOK, cfg.CurrentStateSources)

	case http.MethodPost:
		var src SourceConfig
		if err := json.NewDecoder(r.Body).Decode(&src); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid request body: " + err.Error(),
			})
			return
		}

		if err := s.store.UpsertCurrentSource(src); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "current source upserted",
		})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
	}
}

func (s *apiServer) handleCurrentSourceByName(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/v1/sources/current/")
	if name == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "not found",
		})
		return
	}

	name, err := url.PathUnescape(name)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "bad source name",
		})
		return
	}

	switch r.Method {
	case http.MethodPut:
		var src SourceConfig
		if err := json.NewDecoder(r.Body).Decode(&src); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid request body: " + err.Error(),
			})
			return
		}

		src.Name = name

		if err := s.store.UpsertCurrentSource(src); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "current source updated",
		})

	case http.MethodDelete:
		if err := s.store.DeleteCurrentSource(name); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "current source deleted",
		})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
	}
}

func (s *apiServer) handleSourceTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	var src SourceConfig
	if err := json.NewDecoder(r.Body).Decode(&src); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	// 这里不落盘，只做即时测试。
	result := ProbeSource(src)
	writeJSON(w, http.StatusOK, result)
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

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
