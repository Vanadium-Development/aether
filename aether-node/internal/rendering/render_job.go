package rendering

import (
	"bufio"
	"fmt"
	"io/fs"
	"math"
	"node/internal/config"
	"node/internal/models"
	"node/internal/state"
	"node/internal/util"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

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

func parseTime(s string) float64 {
	parts := strings.Split(s, ":")
	factors := []float64{1.0, 60.0, 3600.0}
	partCount := len(parts)
	if partCount < 1 || partCount > 3 {
		return math.NaN()
	}
	var seconds float64 = 0.0
	for i := 0; i < partCount; i++ {
		f, err := strconv.ParseFloat(parts[partCount-1-i], 64)
		if err != nil {
			return math.NaN()
		}
		seconds = seconds + f*factors[i]
	}

	return seconds
}

// TODO Write to history log!
func invokeBlender(file string, aetherDir string, state *state.State, cfg *config.NodeConfig, req *models.RenderRequest) {
	defer state.RenderLock.Unlock()

	cmd := exec.Command(
		cfg.Node.Blender,
		"-b", file,
		"-s", strconv.Itoa(int(*req.FrameStart)),
		"-e", strconv.Itoa(int(*req.FrameEnd)),
		"-o", filepath.Join(aetherDir, "frame_####"),
		"-a",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logrus.Errorf("Could not open stdout pipe to blender process: %s\n", err)
		return
	}

	cmd.Stderr = cmd.Stdout

	logrus.Infof("Starting render process on %s frames %d to %d ...\n", req.ID, *req.FrameStart, *req.FrameEnd)
	logrus.Debugf("Invoking: %s\n", cmd.String())

	if err := cmd.Start(); err != nil {
		logrus.Errorf("Could not invoke blender process: %s\n", err)
		return
	}

	progressRe := regexp.MustCompile(`Time:(\d+:\d+.\d+).*Remaining:(\d+:\d+.\d+)`)
	scanner := bufio.NewScanner(stdout)

	timeElapsed := math.NaN()
	timeRemaining := math.NaN()

	progressBar := util.SyntheticProgressBar(100, "BLEND")

	for scanner.Scan() {
		line := scanner.Text()
		matches := progressRe.FindStringSubmatch(line)

		if len(matches) != 3 {
			continue
		}

		timeElapsed = parseTime(matches[1])
		timeRemaining = parseTime(matches[2])

		if timeElapsed == math.NaN() || timeRemaining == math.NaN() {
			continue
		}

		progress := (timeElapsed / (timeElapsed + timeRemaining)) * 100
		progressBar.Set(int(progress))
	}

	err = cmd.Wait()
	if err != nil {
		logrus.Errorf("Could not wait for blender process: %s\n", err)
		return
	}

	logrus.Debugf("Blender task finished successfully!")
}

func findBlendFile(cfg *config.NodeConfig, req *models.RenderRequest) string {
	var blendFile = ""
	path := filepath.Join(cfg.Data.WorkspaceDirectory, req.ID.String())
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(d.Name()) == ".blend" && blendFile == "" {
			blendFile = path
		}

		return nil
	})

	if err != nil {
		logrus.Errorf("Could not find *.blend file in scene (%s): %s\n", req.ID.String(), err)
		return ""
	}

	logrus.Debugf("Found blendFile in scene (%s): %s", req.ID, blendFile)
	return blendFile
}

func InitializeRenderProcess(cfg *config.NodeConfig, state *state.State, req *models.RenderRequest) error {
	err := prepareWorkspace(cfg, state.Scene, req)
	if err != nil {
		return err
	}

	var blendFile = findBlendFile(cfg, req)
	if blendFile == "" {
		return fmt.Errorf("could not locate *.blend file")
	}

	var aetherDir = filepath.Join(filepath.Dir(blendFile), ".aether")

	err = os.MkdirAll(aetherDir, 0777)
	if err != nil {
		return err
	}
	logrus.Debugf("Created output directory: %s\n", aetherDir)

	go invokeBlender(blendFile, aetherDir, state, cfg, req)

	return nil
}
