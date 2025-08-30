package api

import (
	"encoding/json"
	"net/http"
	"node/internal/config"
	"node/internal/persistence"
	"node/internal/state"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func route(pattern string, method string, contentType string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		var actualType string

		if method == "*" {
			goto checkContentType
		}

		if r.Method != method {
			http.Error(w, "Method not allowed: Expected "+method+", attempted "+r.Method, http.StatusMethodNotAllowed)
			return
		}

	checkContentType:
		if contentType == "*" {
			goto invokeHandler
		}

		if actualType = r.Header.Get("Content-Type"); actualType == "" {
			actualType = "(empty)"
		}

		if !strings.HasPrefix(actualType, "multipart/form-data") {
			http.Error(w, "Expected "+contentType+", received "+actualType, http.StatusBadRequest)
			return
		}

	invokeHandler:
		handler(w, r)
	})
}

func registerApiRoutes(state *RouteCtx) {
	route("/", http.MethodGet, "*", state.getRootHandler)
	route("/info", http.MethodGet, "*", state.getInfoHandler)
	route("/upload", http.MethodPost, "multipart/form-data", state.postUploadHandler)
	route("/render", http.MethodPost, "application/json", state.postRenderHandler)
}

func RespondJson(w http.ResponseWriter, value map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func InitializeApi(port uint16, node *state.AetherNode, cfg config.NodeConfig) {
	s := &RouteCtx{
		Node:       node,
		Config:     &cfg,
		SceneStore: persistence.LoadStoredScenes(&cfg),
	}

	logrus.Infof("Aether node is listening on http://localhost:%d\n", port)

	registerApiRoutes(s)

	err := http.ListenAndServe(":"+strconv.Itoa(int(port)), nil)
	if err != nil {
		logrus.Errorf("Error initializing API: %s\n", err)
		return
	}
}
