package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"node/internal/banner"
	"node/internal/state"
	"node/internal/version"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Print basic information page if showing the page fails for whatever reason
func (ctx *RouteCtx) handleFallbackInfoPage(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	_, _ = fmt.Fprintf(writer, "%s\n\n--------------------\nAether node is up and running!\nPort: %d\nName: %s\nUUID: %s\nNode Version: %s\n--------------------", banner.AetherBanner, ctx.Port, ctx.Node.Name, ctx.Node.ID, version.AetherVersion)
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
		"Port":      strconv.Itoa(int(ctx.Port)),
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

func acquireMetadata(jsonMeta string, writer http.ResponseWriter) *state.SceneMetadata {
	var metadata state.SceneMetadata
	err := json.Unmarshal([]byte(jsonMeta), &metadata)
	if err != nil {
		http.Error(writer, "Could not parse JSON Metadata", http.StatusBadRequest)
		logrus.Debugf("Could not parse incoming JSON Metadata.\n")
		return nil
	}
	if len(metadata.Checksum) == 0 {
		http.Error(writer, "Metadata does not contain a valid SHA256 checksum", http.StatusBadRequest)
		logrus.Debugf("Incoming JSON Metadata did not contain a SHA256 checksum.\n")
		return nil
	}
	return &metadata
}

func processFile(ctx *RouteCtx, orgFilename string, file multipart.File, metadata *state.SceneMetadata, writer http.ResponseWriter) {
	// Aether only supports *.zip files
	if !strings.HasSuffix(orgFilename, ".zip") {
		http.Error(writer, "The file must be a \"*.zip\" file", http.StatusBadRequest)
		logrus.Debugf("Rejecting file \"%s\": Not a *.zip file.", orgFilename)
		return
	}

	// Make sure both the scene and temp directories exists
	ctx.Config.EnsureFolders()

	// Generate a UUID file name to uniquely identify the file
	id, _ := uuid.NewRandom()
	assignedFilename := id.String() + ".zip"

	logrus.Debugf("File \"%s\" is now \"%s\"", orgFilename, assignedFilename)

	// Store the received file in a temp directory
	tmpFilePath := ctx.Config.Data.TempDirectory + "/" + assignedFilename
	tmpFile, err := os.Create(tmpFilePath)
	if err != nil {
		http.Error(writer, "Could not create temp file for \""+assignedFilename+"\"", http.StatusInternalServerError)
		logrus.Errorf("Could not create temp file: %s\n", err)
		return
	}
	defer tmpFile.Close()
	written, err := io.Copy(tmpFile, file)
	if err != nil {
		http.Error(writer, "Could not write to temp file", http.StatusInternalServerError)
		logrus.Debugf("Could not write incoming file: %s.\n", err)
		return
	}

	// Compare Checksums
	hash := sha256.New()
	tmpFile.Seek(0, io.SeekStart)
	if _, err := io.Copy(hash, tmpFile); err != nil {
		http.Error(writer, "Could not calculate SHA256 checksum", http.StatusUnprocessableEntity)
		logrus.Debugf("Could not calculate SHA256 checksum: %s.\n", err)
		return
	}

	checksum := hash.Sum(nil)
	if bytes.Compare(checksum, metadata.Checksum) != 0 {
		http.Error(writer, "SHA256 Checksum does not match", http.StatusBadRequest)
		logrus.Debugf("SHA256 Checksum of file \"%s\" (%x) does not match the expected value (%x). Deleting it now.\n", tmpFilePath, checksum, metadata.Checksum)
		os.Remove(tmpFilePath)
		return
	}

	// The checksum is correct; Move file to scenes directory
	scenePath := ctx.Config.Data.ScenesDirectory + "/" + assignedFilename
	os.Rename(tmpFilePath, scenePath)

	logrus.Debugf("Successfully stored %d bytes of \"%s\"\n", written, assignedFilename)
	ctx.Node.State.Scene = &state.Scene{Filename: assignedFilename, FilePath: scenePath, Metadata: *metadata}
}

func (ctx *RouteCtx) postCommitHandler(writer http.ResponseWriter, req *http.Request) {
	logrus.Debugf("Parsing request (%d bytes)\n", req.ContentLength)

	err := req.ParseMultipartForm(1_000_000)
	if err != nil {
		http.Error(writer, "Could not commit", http.StatusInternalServerError)
		return
	}

	jsonMeta := req.FormValue("metadata")
	if jsonMeta == "" {
		http.Error(writer, "Metadata is required", http.StatusBadRequest)
		return
	}
	logrus.Debugf("Received Metadata: %s\n", jsonMeta)

	var metadata = acquireMetadata(jsonMeta, writer)
	if metadata == nil {
		return
	}

	file, header, err := req.FormFile("file")
	if err != nil {
		http.Error(writer, "Could not retrieve file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	processFile(ctx, header.Filename, file, metadata, writer)

	writer.Write([]byte("OK."))
}
