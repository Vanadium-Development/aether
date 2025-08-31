package rendering

import (
	"bufio"
	"fmt"
	"io/fs"
	"math"
	"node/internal/config"
	"node/internal/dto/render"
	"node/internal/state"
	"node/internal/util"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
)

// Decompress scene file into a new directory in the workspace
func prepareWorkspace(cfg *config.NodeConfig, scene *state.SceneMetadata, req *render.RenderRequest) error {
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

func parseBlenderOutput(line string, bar *progressbar.ProgressBar) (ok bool, frame int, elapsed float64, remaining float64, progress float64) {
	progressRe := regexp.MustCompile(`Fra:(\d+).*Time:(\d+:\d+.\d+).*Remaining:(\d+:\d+.\d+)`)

	matches := progressRe.FindStringSubmatch(line)

	if len(matches) != 4 {
		return false, -1, math.NaN(), math.NaN(), math.NaN()
	}

	currFrame, err := strconv.Atoi(matches[1])

	if err != nil {
		return false, -1, math.NaN(), math.NaN(), math.NaN()
	}

	timeElapsed := parseTime(matches[2])
	timeRemaining := parseTime(matches[3])

	if timeElapsed == math.NaN() || timeRemaining == math.NaN() {
		return false, -1, math.NaN(), math.NaN(), math.NaN()
	}

	framePercent := (timeElapsed / (timeElapsed + timeRemaining)) * 100
	bar.Set(int(framePercent))

	return true, currFrame, timeElapsed, timeRemaining, framePercent
}

// TODO Write to history log!
func invokeBlender(file string, aetherDir string, state *state.State, cfg *config.NodeConfig, req *render.RenderRequest) {
	// Mark the node as not busy once this function exits
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

	scanner := bufio.NewScanner(stdout)

	progressBar := util.SyntheticProgressBar(100, "BLEND")

	for scanner.Scan() {
		line := scanner.Text()

		ok, frame, elapsed, remaining, progress := parseBlenderOutput(line, progressBar)
		if !ok {
			continue
		}

		state.RendererState.FramePercent = progress
		state.RendererState.CurrentFrame = frame
		state.RendererState.TimeElapsed = elapsed
		state.RendererState.TimeRemaining = remaining
	}

	err = cmd.Wait()
	if err != nil {
		logrus.Errorf("Could not wait for blender process: %s\n", err)
		return
	}

	state.RendererState = nil
	logrus.Debugf("\nBlender task finished successfully!")
}

func findBlendFile(cfg *config.NodeConfig, req *render.RenderRequest) string {
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

func InitializeRenderProcess(cfg *config.NodeConfig, state *state.State, req *render.RenderRequest) error {
	err := prepareWorkspace(cfg, &state.RendererState.Scene, req)
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
