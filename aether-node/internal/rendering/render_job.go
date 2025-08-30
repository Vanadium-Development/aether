package rendering

import (
	"node/internal/config"
	"node/internal/models"
	"node/internal/state"
	"node/internal/util"
	"os"

	"github.com/sirupsen/logrus"
)

// Decompress scene file into a new directory in the workspace
func prepareWorkspace(cfg *config.NodeConfig, scene *state.SceneMetadata, req *models.RenderRequest) error {
	path := cfg.Data.WorkspaceDirectory + "/" + req.ID.String()

	if info, err := os.Stat(path); err == nil {
		if !info.IsDir() {
			goto createDir
		}

		logrus.Debugf("Cleaning up previous workspace: %s\n", path)
		err = os.RemoveAll(path)
		if err != nil {
			logrus.Debugf("Failed to remove directory: %s\n", err)
			goto createDir
		}
	}

createDir:
	err := os.MkdirAll(path, 0777)
	if err != nil {
		logrus.Debugf("Could not create scene directory in workspacee (%s): %s\n", path, err)
		return err
	}

	zipPath := cfg.Data.ScenesDirectory + "/" + scene.Filename

	logrus.Debugf("Decompressing (%s) into (%s) ...\n", zipPath, path)

	err = util.DecompressZip(zipPath, path)
	if err != nil {
		logrus.Debugf("Could not decompress scene file (%s): %s\n", path, err)
		return err
	}

	logrus.Debugf("Decompressing complete.")

	return nil
}

func InitializeRenderProcess(cfg *config.NodeConfig, scene *state.SceneMetadata, req *models.RenderRequest) {
	prepareWorkspace(cfg, scene, req)
}
