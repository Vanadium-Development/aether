package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"node/internal/banner"
	"node/internal/config"
	"node/internal/models"
	"node/internal/persistence"
	"node/internal/rendering"
	"node/internal/state"
	"node/internal/util"
	"node/internal/version"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
)

type RouteCtx struct {
	Node       *state.AetherNode
	Config     *config.NodeConfig
	SceneStore persistence.SceneIndex
}

// Print basic information page if showing the page fails for whatever reason
func (ctx *RouteCtx) handleFallbackInfoPage(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	_, _ = fmt.Fprintf(writer, "%s\n\n--------------------\nAether node is up and running!\nPort: %d\nName: %s\nUUID: %s\nNode Version: %s\n--------------------", banner.AetherBanner, ctx.Node.Port, ctx.Node.Name, ctx.Node.ID, version.AetherVersion)
}

// Return information about current node as a human-readable page
func (ctx *RouteCtx) getRootHandler(writer http.ResponseWriter, req *http.Request) {
	tmpl, err := template.ParseFiles("static/index.html")

	if err != nil {
		goto fallback
	}

	err = tmpl.Execute(writer, map[string]interface{}{
		"UUID":      ctx.Node.ID.String(),
		"Name":      ctx.Node.Name,
		"Port":      strconv.Itoa(int(ctx.Node.Port)),
		"NodeColor": template.CSS(fmt.Sprintf("rgb(%d,%d,%d)", ctx.Node.Color.R, ctx.Node.Color.G, ctx.Node.Color.B)),
		"Version":   version.AetherVersion,
	})

	if err != nil {
		goto fallback
	}

	return

fallback:
	ctx.handleFallbackInfoPage(writer, req)
	logrus.Errorf("Could not parse template: %s\n", err)
	return
}

// Return information about current node as JSON
func (ctx *RouteCtx) getInfoHandler(writer http.ResponseWriter, req *http.Request) {
	RespondJson(writer, ctx.Node.NodeInfoMap())
}

// Upload compressed scene file and store it in the scene index
func (ctx *RouteCtx) postUploadHandler(writer http.ResponseWriter, req *http.Request) {
	if !ctx.Node.State.UploadLock.TryLock() {
		logrus.Debug("Refusing incoming upload request (Handler is busy).")
		http.Error(writer, "Aether node is busy", http.StatusServiceUnavailable)
		return
	}

	defer ctx.Node.State.UploadLock.Unlock()

	logrus.Debugf("Parsing request (%s) ...\n", humanize.Bytes(uint64(req.ContentLength)))

	err := req.ParseMultipartForm(1_000_000)
	if err != nil {
		http.Error(writer, "Could not upload", http.StatusInternalServerError)
		logrus.Debugf("Could not parse multipart form: %s. Cancelling\n", err)
		return
	}

	jsonMeta := req.FormValue("metadata")
	if jsonMeta == "" {
		http.Error(writer, "Metadata is required", http.StatusBadRequest)
		logrus.Debugf("Request did not contain Metadata. Cancelling\n")
		return
	}
	logrus.Debugf("Received Metadata: %s\n", jsonMeta)

	var metadata = acquireMetadata(jsonMeta, writer)
	if metadata == nil {
		return
	}

	// Check if there already is a file with the same checksum
	if existingScene := ctx.SceneStore.FindSceneByChecksum(metadata.Checksum); ctx.SceneStore.FindSceneByChecksum(metadata.Checksum) != nil {
		logrus.Infof("Scene with checksum (%x) already exists. Skipping upload.", metadata.Checksum)
		RespondJson(writer, map[string]interface{}{
			"id": existingScene.ID,
		})
		return
	}

	file, header, err := req.FormFile("file")
	if err != nil {
		http.Error(writer, "Could not retrieve file", http.StatusInternalServerError)
		logrus.Debugf("Could not retrieve form file: %s\n", err)
		return
	}
	defer util.CloseFile(file)

	metadata.CreatedAt = time.Now().UnixNano()

	ok := processFile(ctx, header.Size, header.Filename, file, metadata, writer)
	if !ok {
		return
	}

	// Store Scene Metadata in Scene Index
	ctx.SceneStore.AddScene(*metadata, ctx.Config)

	RespondJson(writer, map[string]interface{}{
		"id": metadata.ID,
	})
}

func (ctx *RouteCtx) postRenderHandler(writer http.ResponseWriter, req *http.Request) {
	if !ctx.Node.State.RenderLock.TryLock() {
		http.Error(writer, "Aether node is currently rendering.", http.StatusServiceUnavailable)
		logrus.Debug("Refusing incoming rendering request (Handler is busy).")
		return
	}
	defer ctx.Node.State.RenderLock.Unlock()

	var request models.RenderRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(writer, "Could not parse JSON render request", http.StatusBadRequest)
		logrus.Debugf("Could not parse JSON render request: %s\n", err)
		return
	}

	rendering.ExecuteRenderScript()
}
