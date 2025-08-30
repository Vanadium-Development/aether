package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"node/internal/state"
	"node/internal/util"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Helper function: Parse and validate JSON metadata into `SceneMetadata`
func acquireMetadata(jsonMeta string, writer http.ResponseWriter) *state.SceneMetadata {
	var metadata state.SceneMetadata
	err := json.Unmarshal([]byte(jsonMeta), &metadata)
	if err != nil {
		http.Error(writer, "Could not parse JSON Metadata", http.StatusBadRequest)
		logrus.Debugf("Could not parse incoming JSON Metadata. Cancelling.\n")
		return nil
	}
	if len(metadata.Checksum) == 0 {
		http.Error(writer, "Metadata does not contain a valid SHA256 checksum", http.StatusBadRequest)
		logrus.Debugf("Incoming JSON Metadata did not contain a SHA256 checksum. Cancelling.\n")
		return nil
	}
	return &metadata
}

func processFile(ctx *RouteCtx, fileSize int64, filename string, file multipart.File, metadata *state.SceneMetadata, writer http.ResponseWriter) bool {
	// Aether only supports *.zip files
	if !strings.HasSuffix(filename, ".zip") {
		http.Error(writer, "The file must be a \"*.zip\" file", http.StatusBadRequest)
		logrus.Debugf("Rejecting file \"%s\": Not a *.zip file.", filename)
		return false
	}

	// Generate a UUID file name to uniquely identify the file
	id, _ := uuid.NewRandom()
	randomFilename := id.String() + ".zip"

	logrus.Debugf("File \"%s\" is now \"%s\"", filename, randomFilename)

	// Store the received file in a temp directory
	bar := util.ByteProgressBar(fileSize, "TRANS ")

	tmpFilePath := ctx.Config.Data.TempDirectory + "/" + randomFilename
	tmpFile, err := os.Create(tmpFilePath)
	if err != nil {
		http.Error(writer, "Could not create temp file for \""+randomFilename+"\"", http.StatusInternalServerError)
		logrus.Errorf("Could not create temp file: %s\n", err)
		return false
	}
	defer util.CloseFile(tmpFile)
	written, err := io.Copy(io.MultiWriter(tmpFile, bar), file)
	if err != nil {
		http.Error(writer, "Could not write to temp file", http.StatusInternalServerError)
		logrus.Debugf("Could not write incoming file: %s.\n", err)
		return false
	}

	// Compare Checksums
	hash := sha256.New()
	if _, err = tmpFile.Seek(0, io.SeekStart); err != nil {
		http.Error(writer, "Could not seek temp file", http.StatusInternalServerError)
		logrus.Errorf("Could not rewind temp file: %s\n", err)
		return false
	}
	if _, err := io.Copy(hash, tmpFile); err != nil {
		http.Error(writer, "Could not calculate SHA256 checksum", http.StatusUnprocessableEntity)
		logrus.Debugf("Could not calculate SHA256 checksum: %s.\n", err)
		return false
	}

	checksum := hash.Sum(nil)
	if bytes.Compare(checksum, metadata.Checksum) != 0 {
		http.Error(writer, "SHA256 Checksum does not match", http.StatusBadRequest)
		logrus.Debugf("SHA256 Checksum of file \"%s\" (%x) does not match the expected value (%x). Deleting it now.\n", tmpFilePath, checksum, metadata.Checksum)
		if err := os.Remove(tmpFilePath); err != nil {
			logrus.Errorf("Could not remove temp file \"%s\": %s\n", tmpFilePath, err)
			return false
		}
		return false
	}

	logrus.Debugf("SHA256 Checksums match!\n")

	// The checksum is correct; Move file to scenes directory
	scenePath := ctx.Config.Data.ScenesDirectory + "/" + randomFilename
	err = os.Rename(tmpFilePath, scenePath)
	if err != nil {
		logrus.Errorf("Could not rename temp file \"%s\": %s\n", tmpFilePath, err)
		return false
	}

	fmt.Println()
	logrus.Debugf("Successfully stored %s of \"%s\"\n", humanize.Bytes(uint64(written)), randomFilename)

	metadata.Filename = randomFilename
	metadata.OriginalName = filename
	metadata.ID = id

	return true
}
