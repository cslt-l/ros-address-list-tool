package app

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxRequestBodyBytes int64 = 1 << 20 // 1 MiB

// NewHTTPHandler 创建整个 HTTP API 与静态页面的根处理器。
func NewHTTPHandler(store *ConfigStore, logger *log.Logger) http.Handler {
	cfg := store.GetConfig()

	s := &apiServer{
		store:    store,
		logger:   logger,
		sessions: NewSessionManager(time.Duration(cfg.Server.SessionTTLMinutes) * time.Minute),
	}

	mux := http.NewServeMux()

	// ===================== 基础与认证接口 =====================

	mux.HandleFunc("/healthz", s.wrap(s.handleHealthz))
	mux.HandleFunc("/api/v1/auth/login", s.wrap(s.handleLogin))
	mux.HandleFunc("/api/v1/auth/logout", s.wrap(s.handleLogout))
	mux.HandleFunc("/api/v1/auth/me", s.wrap(s.handleAuthMe))
	mux.HandleFunc("/api/v1/auth/change-password", s.wrap(s.handleChangePassword))

	// ===================== 业务 API 接口 =====================

	mux.HandleFunc("/api/v1/config", s.wrap(s.handleConfig))
	mux.HandleFunc("/api/v1/render", s.wrap(s.handleRender))

	mux.HandleFunc("/api/v1/lists", s.wrap(s.handleLists))
	mux.HandleFunc("/api/v1/lists/", s.wrap(s.handleListByName))

	mux.HandleFunc("/api/v1/manual-rules", s.wrap(s.handleManualRules))
	mux.HandleFunc("/api/v1/manual-rules/", s.wrap(s.handleManualRuleByID))

	mux.HandleFunc("/api/v1/sources/desired", s.wrap(s.handleDesiredSources))
	mux.HandleFunc("/api/v1/sources/desired/", s.wrap(s.handleDesiredSourceByName))
	mux.HandleFunc("/api/v1/sources/current", s.wrap(s.handleCurrentSources))
	mux.HandleFunc("/api/v1/sources/current/", s.wrap(s.handleCurrentSourceByName))
	mux.HandleFunc("/api/v1/sources/test", s.wrap(s.handleSourceTest))

	// ===================== 静态页面与入口 =====================

	if cfg.Server.EnableWeb && dirExists(cfg.Server.WebDir) {
		mux.HandleFunc("/", s.handleWeb)
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{
				"message":     "RouterOS address-list HTTP API is running",
				"web_enabled": cfg.Server.EnableWeb,
				"web_dir":     cfg.Server.WebDir,
				"routes": []string{
					"GET  /healthz",
					"POST /api/v1/auth/login",
					"POST /api/v1/auth/logout",
					"GET  /api/v1/auth/me",
					"POST /api/v1/auth/change-password",
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
	store    *ConfigStore
	logger   *log.Logger
	sessions *SessionManager
}

func extractBearerToken(authHeader string) string {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return ""
	}

	const prefix = "Bearer "
	if len(authHeader) < len(prefix) || !strings.EqualFold(authHeader[:len(prefix)], prefix) {
		return ""
	}

	return strings.TrimSpace(authHeader[len(prefix):])
}

func secureTokenEqual(got, want string) bool {
	if got == "" || want == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

func redactHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	out := make(map[string]string, len(headers))
	for k, v := range headers {
		lk := strings.ToLower(strings.TrimSpace(k))
		switch {
		case lk == "authorization",
			lk == "proxy-authorization",
			strings.Contains(lk, "token"),
			strings.Contains(lk, "secret"),
			strings.Contains(lk, "api-key"),
			strings.Contains(lk, "apikey"):
			out[k] = "***redacted***"
		default:
			out[k] = v
		}
	}
	return out
}

func redactSources(items []SourceConfig) []SourceConfig {
	if len(items) == 0 {
		return nil
	}

	out := make([]SourceConfig, len(items))
	copy(out, items)
	for i := range out {
		out[i].Headers = redactHeaders(out[i].Headers)
	}
	return out
}

func redactConfig(cfg AppConfig) AppConfig {
	cfg.Server.AuthToken = ""
	cfg.Server.LoginPassword = ""
	cfg.Server.LoginPasswordHash = ""
	cfg.DesiredSources = redactSources(cfg.DesiredSources)
	cfg.CurrentStateSources = redactSources(cfg.CurrentStateSources)
	return cfg
}

func (s *apiServer) sessionCookieName() string {
	cfg := s.store.GetConfig()
	name := strings.TrimSpace(cfg.Server.SessionCookieName)
	if name == "" {
		return defaultSessionCookieName
	}
	return name
}

func (s *apiServer) currentSession(r *http.Request) (Session, bool) {
	cookie, err := r.Cookie(s.sessionCookieName())
	if err != nil {
		return Session{}, false
	}
	return s.sessions.Touch(strings.TrimSpace(cookie.Value))
}

func (s *apiServer) setSessionCookie(w http.ResponseWriter, r *http.Request, sess Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.sessionCookieName(),
		Value:    sess.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		Expires:  sess.ExpiresAt,
	})
}

func (s *apiServer) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.sessionCookieName(),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *apiServer) allowAnonymousRoute(path string) bool {
	switch path {
	case "/healthz", "/api/v1/auth/login", "/login", "/login.html":
		return true
	default:
		return path == "/favicon.svg" || path == "/app.css"
	}
}

func (s *apiServer) allowDuringPasswordChange(path string) bool {
	switch path {
	case "/api/v1/auth/me", "/api/v1/auth/logout", "/api/v1/auth/change-password", "/login", "/login.html", "/healthz":
		return true
	default:
		return path == "/favicon.svg" || path == "/app.css"
	}
}

func (s *apiServer) authorizeAPIRequest(w http.ResponseWriter, r *http.Request) bool {
	if s.allowAnonymousRoute(r.URL.Path) {
		return true
	}

	cfg := s.store.GetConfig()

	if strings.HasPrefix(r.URL.Path, "/api/") && isAuthorizedAPIRequest(r, strings.TrimSpace(cfg.Server.AuthToken)) {
		return true
	}

	sess, hasSession := s.currentSession(r)

	if strings.HasPrefix(r.URL.Path, "/api/") {
		if !cfg.Server.LoginEnabled {
			if strings.HasPrefix(r.URL.Path, "/api/v1/auth/") {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "login disabled"})
				return false
			}
			if strings.TrimSpace(cfg.Server.AuthToken) == "" {
				return true
			}
			w.Header().Set("WWW-Authenticate", `Bearer realm="ros-address-list-tool"`)
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return false
		}

		if !hasSession {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"error":        "unauthorized",
				"login_needed": true,
			})
			return false
		}

		if sess.RequirePasswordChange && !s.allowDuringPasswordChange(r.URL.Path) {
			writeJSON(w, http.StatusForbidden, map[string]any{
				"error":                "password change required",
				"must_change_password": true,
			})
			return false
		}
		return true
	}

	if !cfg.Server.LoginEnabled {
		return true
	}
	if !hasSession {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return false
	}
	if sess.RequirePasswordChange && !s.allowDuringPasswordChange(r.URL.Path) {
		http.Redirect(w, r, "/login.html?force_change=1", http.StatusFound)
		return false
	}
	return true
}

func (s *apiServer) wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.authorizeAPIRequest(w, r) {
			return
		}
		next(w, r)
	}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (s *apiServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	cfg := s.store.GetConfig()
	if !cfg.Server.LoginEnabled {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "login disabled"})
		return
	}

	var req loginRequest
	if !decodeJSONBody(w, r, &req, false) {
		return
	}

	if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(req.Username)), []byte(resolveLoginUsername(cfg))) != 1 ||
		!verifyLoginPassword(cfg, req.Password) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "用户名或密码错误"})
		return
	}

	sess, err := s.sessions.Create(resolveLoginUsername(cfg), cfg.Server.PasswordChangeRequired)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建会话失败"})
		return
	}

	s.setSessionCookie(w, r, sess)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":                   true,
		"username":             sess.Username,
		"must_change_password": sess.RequirePasswordChange,
	})
}

func (s *apiServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	if cookie, err := r.Cookie(s.sessionCookieName()); err == nil {
		s.sessions.Delete(strings.TrimSpace(cookie.Value))
	}
	s.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *apiServer) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	cfg := s.store.GetConfig()
	if !cfg.Server.LoginEnabled {
		writeJSON(w, http.StatusOK, map[string]any{
			"login_enabled": false,
			"authenticated": false,
		})
		return
	}

	sess, ok := s.currentSession(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"authenticated": false,
			"login_enabled": true,
			"login_url":     "/login.html",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated":        true,
		"login_enabled":        true,
		"username":             sess.Username,
		"must_change_password": sess.RequirePasswordChange,
	})
}

func (s *apiServer) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	cfg := s.store.GetConfig()
	if !cfg.Server.LoginEnabled {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "login disabled"})
		return
	}

	sess, ok := s.currentSession(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req changePasswordRequest
	if !decodeJSONBody(w, r, &req, false) {
		return
	}

	if !verifyLoginPassword(cfg, req.CurrentPassword) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "当前密码不正确"})
		return
	}
	if err := validateNewPassword(req.NewPassword); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	newHash, err := hashNewLoginPassword(req.NewPassword)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "生成密码摘要失败"})
		return
	}

	if err := s.store.Update(func(next *AppConfig) error {
		next.Server.LoginEnabled = true
		next.Server.LoginUsername = resolveLoginUsername(*next)
		next.Server.LoginPassword = ""
		next.Server.LoginPasswordHash = newHash
		next.Server.PasswordChangeRequired = false
		return nil
	}); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	updated, _ := s.sessions.UpdateRequirePasswordChange(sess.ID, false)
	s.setSessionCookie(w, r, updated)

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":                   true,
		"username":             updated.Username,
		"must_change_password": false,
	})
}

func (s *apiServer) handleWeb(w http.ResponseWriter, r *http.Request) {
	cfg := s.store.GetConfig()
	fileServer := http.FileServer(http.Dir(cfg.Server.WebDir))

	switch r.URL.Path {
	case "/", "/index.html":
		if cfg.Server.LoginEnabled {
			sess, ok := s.currentSession(r)
			if !ok {
				http.Redirect(w, r, "/login.html", http.StatusFound)
				return
			}
			if sess.RequirePasswordChange {
				http.Redirect(w, r, "/login.html?force_change=1", http.StatusFound)
				return
			}
		}
		http.ServeFile(w, r, filepath.Join(cfg.Server.WebDir, "index.html"))
		return

	case "/login", "/login.html":
		http.ServeFile(w, r, filepath.Join(cfg.Server.WebDir, "login.html"))
		return

	case "/app.css", "/favicon.svg":
		fileServer.ServeHTTP(w, r)
		return

	case "/app.js":
		if cfg.Server.LoginEnabled {
			if _, ok := s.currentSession(r); !ok {
				http.Redirect(w, r, "/login.html", http.StatusFound)
				return
			}
		}
		fileServer.ServeHTTP(w, r)
		return

	default:
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		if cfg.Server.LoginEnabled {
			if _, ok := s.currentSession(r); !ok {
				http.Redirect(w, r, "/login.html", http.StatusFound)
				return
			}
		}
		fileServer.ServeHTTP(w, r)
	}
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
	writeJSON(w, http.StatusOK, redactConfig(cfg))
}

type renderRequest struct {
	Mode RenderMode `json:"mode,omitempty"`
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
	if !decodeJSONBody(w, r, &req, true) {
		return
	}

	if req.Mode != "" {
		cfg.Output.Mode = req.Mode
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
		if !decodeJSONBody(w, r, &def, false) {
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
	name, isDescription, status, errMsg := parseListRequestPath(r.URL.Path)
	if status != 0 {
		writeJSON(w, status, map[string]string{
			"error": errMsg,
		})
		return
	}

	if isDescription {
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
		if !decodeJSONBody(w, r, &def, false) {
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
		if !decodeJSONBody(w, r, &body, false) {
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
		if !decodeJSONBody(w, r, &rule, false) {
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
	id, status, errMsg := parseSinglePathValue(r.URL.Path, "/api/v1/manual-rules/", "rule id")
	if status != 0 {
		writeJSON(w, status, map[string]string{
			"error": errMsg,
		})
		return
	}

	switch r.Method {
	case http.MethodPut:
		var rule ManualRule
		if !decodeJSONBody(w, r, &rule, false) {
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
		writeJSON(w, http.StatusOK, redactSources(cfg.DesiredSources))

	case http.MethodPost:
		var src SourceConfig
		if !decodeJSONBody(w, r, &src, false) {
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
	name, status, errMsg := parseSinglePathValue(r.URL.Path, "/api/v1/sources/desired/", "source name")
	if status != 0 {
		writeJSON(w, status, map[string]string{
			"error": errMsg,
		})
		return
	}

	switch r.Method {
	case http.MethodPut:
		var src SourceConfig
		if !decodeJSONBody(w, r, &src, false) {
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
		writeJSON(w, http.StatusOK, redactSources(cfg.CurrentStateSources))

	case http.MethodPost:
		var src SourceConfig
		if !decodeJSONBody(w, r, &src, false) {
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
	name, status, errMsg := parseSinglePathValue(r.URL.Path, "/api/v1/sources/current/", "source name")
	if status != 0 {
		writeJSON(w, status, map[string]string{
			"error": errMsg,
		})
		return
	}

	switch r.Method {
	case http.MethodPut:
		var src SourceConfig
		if !decodeJSONBody(w, r, &src, false) {
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

	if !isLoopbackRequest(r) {
		writeJSON(w, http.StatusForbidden, map[string]string{
			"error": "source test 仅允许本机请求",
		})
		return
	}

	var src SourceConfig
	if !decodeJSONBody(w, r, &src, false) {
		return
	}

	switch strings.TrimSpace(src.Type) {
	case "file":
		writeJSON(w, http.StatusForbidden, map[string]string{
			"error": "HTTP source test 不允许测试 file 类型来源",
		})
		return

	case "url":
		if err := validateProbeURL(src.URL); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid source url: " + err.Error(),
			})
			return
		}

	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "source.type 只支持 file 或 url",
		})
		return
	}

	result := ProbeSource(src)
	writeJSON(w, http.StatusOK, result)
}

func loggingMiddleware(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("HTTP %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any, allowEmpty bool) bool {
	if r.Body == nil {
		if allowEmpty {
			return true
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "request body is required",
		})
		return false
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		if allowEmpty && errors.Is(err, io.EOF) {
			return true
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body: " + err.Error(),
		})
		return false
	}

	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "request body must contain exactly one JSON object",
		})
		return false
	}

	return true
}

func parseListRequestPath(rawPath string) (name string, isDescription bool, status int, errMsg string) {
	tail := strings.TrimPrefix(rawPath, "/api/v1/lists/")
	if tail == "" {
		return "", false, http.StatusNotFound, "not found"
	}

	parts := strings.Split(tail, "/")
	switch {
	case len(parts) == 1:
		decoded, status, errMsg := decodePathValue(parts[0], "list name")
		return decoded, false, status, errMsg

	case len(parts) == 2 && parts[1] == "description":
		decoded, status, errMsg := decodePathValue(parts[0], "list name")
		return decoded, true, status, errMsg

	default:
		return "", false, http.StatusNotFound, "not found"
	}
}

func parseSinglePathValue(rawPath, prefix, label string) (string, int, string) {
	tail := strings.TrimPrefix(rawPath, prefix)
	if tail == "" {
		return "", http.StatusNotFound, "not found"
	}
	if strings.Contains(tail, "/") {
		return "", http.StatusNotFound, "not found"
	}
	return decodePathValue(tail, label)
}

func decodePathValue(rawValue, label string) (string, int, string) {
	value, err := url.PathUnescape(rawValue)
	if err != nil {
		return "", http.StatusBadRequest, "bad " + label
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", http.StatusNotFound, "not found"
	}
	if strings.Contains(value, "/") {
		return "", http.StatusBadRequest, "bad " + label
	}
	return value, 0, ""
}

func validateProbeURL(raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("只允许 http 或 https")
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return fmt.Errorf("url 缺少 host")
	}
	return nil
}

func isLoopbackRequest(r *http.Request) bool {
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		host = strings.TrimSpace(r.RemoteAddr)
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
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
