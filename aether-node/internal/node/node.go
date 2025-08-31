package node

import (
	"fmt"
	"node/internal/api"
	"node/internal/banner"
	"node/internal/config"
	"node/internal/state"
	"node/internal/util"
	"node/internal/version"
	"runtime"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func establishPlatform() state.Platform {
	system := strings.ToLower(runtime.GOOS)

	if strings.Contains(system, "windows") {
		return state.Windows
	}

	if strings.Contains(system, "linux") || strings.Contains(system, "bsd") ||
		strings.Contains(system, "solaris") || strings.Contains(system, "darwin") {
		return state.Unix
	}

	logrus.Fatal("Could not determine platform: %s\n", system)
	return state.Unix // Doesn't really matter
}

func InitializeNode() {
	fmt.Println(banner.AetherBanner)
	fmt.Printf("\t- Version %s -\n\n", version.AetherVersion)

	var nodeId, _ = uuid.NewRandom()
	var cfg = config.ParseNodeConfig()

	port := cfg.Node.Port

	var n = &state.AetherNode{
		ID:       nodeId,
		Name:     cfg.Node.Name,
		Port:     port,
		Color:    util.RandomNodeColor(),
		Platform: establishPlatform(),
		State: state.State{
			UploadLock:    sync.Mutex{},
			RenderLock:    sync.Mutex{},
			RendererState: nil,
		}}

	if n.Platform == state.Windows {
		logrus.Infof("Aether node is running on Windows")
	} else {
		logrus.Infof("Aether node is running on Unix")
	}

	logrus.Infof("Using blender path: %s\n", cfg.Node.Blender)

	// Make sure all required directories exist
	cfg.EnsureFolders()

	api.InitializeApi(port, n, cfg)
}
