package util

import (
	"archive/zip"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

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

func DecompressZip(src string, dst string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, zipFile := range reader.File {
		path := filepath.Join(dst, zipFile.Name)

		if zipFile.FileInfo().IsDir() {
			os.MkdirAll(path, 0777)
			continue
		}

		// Create all directories leading up to the file
		err = os.MkdirAll(filepath.Dir(path), 0777)
		if err != nil {
			return err
		}

		// Open Destination file
		dstFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			return err
		}

		// Open Source file
		srcReader, err := zipFile.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(dstFile, srcReader)
		if err != nil {
			return err
		}

		dstFile.Close()
		srcReader.Close()
	}

	return nil
}
