package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"node/internal/banner"
	"node/internal/config"
	"node/internal/dto/id"
	"node/internal/dto/progress"
	"node/internal/dto/render"
	"node/internal/dto/scenes"
	"node/internal/persistence"
	"node/internal/rendering"
	"node/internal/state"
	"node/internal/util"
	"node/internal/version"
	"os"
	"path/filepath"
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

	var status string = "IDLE"
	if ctx.Node.State.RendererState != nil {
		status = "RENDERING"
	}

	if err != nil {
		goto fallback
	}

	err = tmpl.Execute(writer, map[string]interface{}{
		"UUID":      ctx.Node.ID.String(),
		"Name":      ctx.Node.Name,
		"Port":      strconv.Itoa(int(ctx.Node.Port)),
		"NodeColor": template.CSS(fmt.Sprintf("rgb(%d,%d,%d)", ctx.Node.Color.R, ctx.Node.Color.G, ctx.Node.Color.B)),
		"Version":   version.AetherVersion,
		"Blender":   ctx.Config.Node.Blender,
		"Status":    status,
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

// Retrieve a list of scenes stored in the scene index
func (ctx *RouteCtx) getScenesHandler(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(scenes.SceneIndexResponseFromIndex(&ctx.SceneStore))
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

// Start a rendering job on a previously uploaded scene
func (ctx *RouteCtx) postRenderHandler(writer http.ResponseWriter, req *http.Request) {
	if !ctx.Node.State.RenderLock.TryLock() {
		http.Error(writer, "Aether node is currently rendering.", http.StatusServiceUnavailable)
		logrus.Debug("Refusing incoming rendering request (Renderer is busy).")
		return
	}

	var request render.RenderRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(writer, "Could not parse JSON render request", http.StatusBadRequest)
		logrus.Debugf("Could not parse JSON render request: %s\n", err)

		ctx.Node.State.RenderLock.Unlock()
		return
	}

	if request.FrameStart == nil {
		http.Error(writer, "Expected required field \"frame_start\" as part of render request", http.StatusBadRequest)
		logrus.Debugf("Render request did not contain required field \"frame_start\"\n")

		ctx.Node.State.RenderLock.Unlock()
		return
	}

	if request.FrameEnd == nil {
		http.Error(writer, "Expected required field \"frame_end\" as part of render request", http.StatusBadRequest)
		logrus.Debugf("Render request did not contain required field \"frame_end\"\n")

		ctx.Node.State.RenderLock.Unlock()
		return
	}

	if request.ID == nil {
		http.Error(writer, "Expected required field \"id\" as part of render request", http.StatusBadRequest)
		logrus.Debugf("Render request did not contain required field \"id\"\n")

		ctx.Node.State.RenderLock.Unlock()
		return
	}

	var scene *state.SceneMetadata

	if scene = ctx.SceneStore.FindSceneById(*request.ID); scene == nil {
		http.Error(writer, "A scene with this ID does not exist", http.StatusBadRequest)
		logrus.Debug("Could not find a scene with the requested ID (%s)\n", request.ID)

		ctx.Node.State.RenderLock.Unlock()
		return
	}

	// This is where we create the RenderState for the first time
	ctx.Node.State.RendererState = &state.RendererState{Scene: *scene, Request: request, CurrentFrame: 0, FramePercent: 0.0}

	err := rendering.InitializeRenderProcess(ctx.Config, &ctx.Node.State, &request)
	if err != nil {
		http.Error(writer, "Could not invoke renderer", http.StatusInternalServerError)
		logrus.Debugf("Could not invoke renderer: %s\n", err)

		ctx.Node.State.RendererState = nil
		ctx.Node.State.RenderLock.Unlock()
		return
	}

	writer.Write([]byte("OK"))
	// Note: RenderLock is still locked if InitializeRenderProcess succeeded!
}

// Retrieve information about the current rendering job
func (ctx *RouteCtx) getStatusHandler(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	renderState := ctx.Node.State.RendererState
	_ = json.NewEncoder(writer).Encode(progress.StatusResponseFromRenderState(renderState))
	return
}

// Retrieve the last render result of a given scene
func (ctx *RouteCtx) getRenderResult(writer http.ResponseWriter, req *http.Request) {
	var request id.IDRequest
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		http.Error(writer, "Could not parse JSON request", http.StatusBadRequest)
		logrus.Debugf("Could not parse JSON request: %s\n", err)
		return
	}

	filename := request.ID.String() + ".zip"
	path := filepath.Join(ctx.Config.Data.OutputDirectory, filename)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.Error(writer, "A last render result does not exist for this scene", http.StatusNotFound)
		logrus.Debugf("Render result does not exist: %s\n", path)
		return
	}

	f, err := os.Open(path)
	if err != nil {
		http.Error(writer, "Could not open file for reading", http.StatusInternalServerError)
		logrus.Debugf("Could not open file for reading (%s): %s\n", path, err)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		http.Error(writer, "Could not stat file", http.StatusInternalServerError)
		logrus.Debugf("Could not stat file (%s): %s\n", path, err)
		return
	}

	logrus.Debugf("Returning render result for scene: %s\n", request.ID)

	writer.Header().Set("Content-Type", "application/zip")
	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	http.ServeContent(writer, req, filename, info.ModTime(), f)
}
