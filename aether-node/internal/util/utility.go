package util

import (
	"mime/multipart"

	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
)

func CloseFile(file multipart.File) {
	err := file.Close()
	if err != nil {
		logrus.Errorf("Error closing file: %s", err)
	}
}

func ProgressBar(fileSize int64, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions(int(fileSize),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetDescription("[cyan]"+description+"[reset]"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[light_green]█[reset]",
			SaucerPadding: "[reset]█",
			BarStart:      "",
			BarEnd:        "",
		}))
}
