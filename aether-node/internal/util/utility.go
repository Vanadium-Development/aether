package util

import (
	"archive/zip"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

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

func ByteProgressBar(fileSize int64, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions(int(fileSize),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetDescription("[light_cyan]"+description+"[reset]"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[light_green]█[reset]",
			SaucerPadding: "[reset]█",
			BarStart:      "",
			BarEnd:        "",
		}))
}

func SyntheticProgressBar(count int, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions(count,
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetElapsedTime(true),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetDescription("[light_blue]"+description+"[reset]"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[light_red]█[reset]",
			SaucerPadding: "[reset]█",
			BarStart:      "",
			BarEnd:        "",
		}))
}

func CompressZip(src string, dstZip string) error {
	zipFile, err := os.Create(dstZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Get Relative Path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath

		entryWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(entryWriter, file)
		file.Close()

		return err
	})

	return err
}

func DecompressZip(src string, dst string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	bar := SyntheticProgressBar(len(reader.File), "UNZIP")
	bar.RenderBlank()

	for _, zipFile := range reader.File {
		path := filepath.Join(dst, zipFile.Name)

		if strings.HasPrefix(strings.ToLower(zipFile.Name), "__macosx/") ||
			strings.HasSuffix(strings.ToLower(zipFile.Name), ".ds_store") {
			bar.Add(1)
			continue
		}

		if zipFile.FileInfo().IsDir() {
			os.MkdirAll(path, 0777)
			bar.Add(1)
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

		bar.Add(1)
		dstFile.Close()
		srcReader.Close()
	}

	fmt.Println()

	return nil
}
