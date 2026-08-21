package logger

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"time"
)

func makeArchive(dir string, days int) error {
	timeArch := time.Now().AddDate(0, 0, -days).String()[:10]

	subdirs, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, subdir := range subdirs {
		if !subdir.IsDir() || subdir.Name() >= timeArch {
			continue
		}

		subdirPath := filepath.Join(dir, subdir.Name())
		zipPath := subdirPath + ".zip"

		if err := archiveDir(subdirPath, zipPath); err != nil {
			return err
		}

		if err := os.RemoveAll(subdirPath); err != nil {
			return err
		}
	}

	return nil
}

func archiveDir(srcDir, dstZip string) error {
	zipFile, err := os.Create(dstZip)
	if err != nil {
		return err
	}
	defer zipFile.Close() // nolint: errcheck

	w := zip.NewWriter(zipFile)
	defer w.Close() // nolint: errcheck

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(srcDir, entry.Name())
		if err := addToZip(w, srcPath); err != nil {
			return err
		}
	}

	return nil
}

func addToZip(zipWriter *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close() // nolint: errcheck

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = info.Name()
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}
