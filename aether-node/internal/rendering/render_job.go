package rendering

import (
	"bufio"
	"fmt"
	"io/fs"
	"math"
	"node/internal/config"
	"node/internal/state"
	"node/internal/util"
	"node/pkg/dtos"
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
func prepareWorkspace(cfg *config.NodeConfig, scene *state.SceneMetadata, req *dtos.AetherRenderDto) error {
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

func parseBlenderOutput(line string) (ok bool, frame int, elapsed float64, remaining float64, progress float64) {
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

	return true, currFrame, timeElapsed, timeRemaining, framePercent
}

func collectResults(cfg *config.NodeConfig, aetherDir string, req *dtos.AetherRenderDto) error {
	dst := filepath.Join(cfg.Data.OutputDirectory, req.ID.String()+".zip")

	// Remove previous results file
	_ = os.RemoveAll(dst)

	err := util.CompressZip(aetherDir, dst)
	if err != nil {
		return err
	}

	logrus.Infof("Collected render results to: %s", dst)

	// Delete workspace directory
	path := filepath.Join(cfg.Data.WorkspaceDirectory, req.ID.String())
	err = os.RemoveAll(path)
	if err != nil {
		logrus.Errorf("Could not remove workspace directory (%s): %s\n", path, err)
		return err
	}

	logrus.Debugf("Removed workspace directory: %s", path)
	return nil
}

func invokeBlender(file string, aetherDir string, state *state.State, cfg *config.NodeConfig, req *dtos.AetherRenderDto) {
	// Mark the node as not busy once this function exits
	defer state.RenderLock.Unlock()

	cmd := exec.Command(
		cfg.Node.Blender,
		"-b", file,
		"-s", strconv.Itoa(int(*req.FrameStart)),
		"-e", strconv.Itoa(int(*req.FrameEnd)),
		"-o", filepath.Join(aetherDir, "aether-frame_####"),
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

	var bar *progressbar.ProgressBar = nil
	lastFrame := -1

	for scanner.Scan() {
		line := scanner.Text()

		ok, frame, elapsed, remaining, progress := parseBlenderOutput(line)
		if !ok {
			continue
		}

		if lastFrame != frame {
			// Make sure all frames end on 100%
			if bar != nil {
				bar.Set(100)
				fmt.Println()
			}
			lastFrame = frame
			bar = util.SyntheticProgressBar(100, "FRAME "+strconv.Itoa(frame))
			bar.RenderBlank()
		}

		bar.Set(int(progress))

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

	// Make sure the last frame also ends on 100%
	if bar != nil {
		bar.Set(100)
		fmt.Println()
	}

	state.RendererState = nil
	logrus.Debugf("Blender task finished successfully. Preparing result set")

	err = collectResults(cfg, aetherDir, req)
	if err != nil {
		logrus.Errorf("Could not collect results: %s\n", err)
		return
	}

	logrus.Infof("Task completed successfully.")
}

func findBlendFile(cfg *config.NodeConfig, req *dtos.AetherRenderDto) string {
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

func InitializeRenderProcess(cfg *config.NodeConfig, state *state.State, req *dtos.AetherRenderDto) error {
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
