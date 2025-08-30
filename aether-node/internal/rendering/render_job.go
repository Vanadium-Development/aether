package rendering

import (
	"node/internal/api"
	"node/internal/models"
	"node/internal/state"
	"node/internal/util"
	"os"

	"github.com/sirupsen/logrus"
)

// Decompress scene file into a new directory in the workspace
func prepareWorkspace(ctx *api.RouteCtx, scene *state.SceneMetadata, req *models.RenderRequest) error {
	path := ctx.Config.Data.WorkspaceDirectory + "/" + req.ID.String()
	err := os.MkdirAll(path, 0777)
	if err != nil {
		logrus.Debugf("Could not create scene directory in workspacee (%s): %s\n", path, err)
		return err
	}

	zipPath := ctx.Config.Data.ScenesDirectory + "/" + scene.Filename

	logrus.Debugf("Decompressing (%s) into (%s) ...\n", zipPath, path)

	err = util.DecompressZip(zipPath, path)
	if err != nil {
		logrus.Debugf("Could not decompress scene file (%s): %s\n", path, err)
		return err
	}
}

func InitializeRenderProcess(ctx *api.RouteCtx, scene *state.SceneMetadata, req *models.RenderRequest) {
	prepareWorkspace(ctx, scene, req)
}
