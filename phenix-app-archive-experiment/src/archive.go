package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"path/filepath"
)

type fileList struct {
	files   map[string]bool
	filters []string
}

func (this *fileList) addFile(path string, info os.FileInfo, err error) error {

	// Only interested in files
	if info.IsDir() {
		return nil
	}

	if _, ok := this.files[path]; !ok {

		// If there are no filters defined,
		// add the file
		if len(this.filters) == 0 {
			this.files[path] = true
		}

		for _, filterItem := range this.filters {
			if ok, _ := regexp.MatchString(filterItem, path); ok {
				this.files[path] = true
				break
			}
		}
	}

	return nil

}

func getFileList(parentDirectory string, filterSpec []string) map[string]bool {

	files := new(fileList)
	files.files = make(map[string]bool)
	files.filters = filterSpec

	// Make sure the parent directory exists
	if !pathExists(parentDirectory) {
		return files.files
	}

	// Get a list of all the files to compress
	filepath.Walk(parentDirectory, files.addFile)

	return files.files

}

func createTarGz(archiveSpec *ArchiveSpec) error {

	filesToArchive := archiveSpec.files

	createDirectory(archiveSpec.Output)
	output, err := os.Create(filepath.Join(archiveSpec.Output, archiveSpec.Name))

	if err != nil {
		return fmt.Errorf("unable to create %v", output)
	}

	compressedArchive := gzip.NewWriter(output)
	defer compressedArchive.Close()

	archive := tar.NewWriter(compressedArchive)
	defer archive.Close()

	for filePath, _ := range filesToArchive {

		info, _ := os.Stat(filePath)

		// create a new dir/file header
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		// Use the base name if adding a configuration file
		// to avoid putting the name of the temporary directory
		// in the archive
		if strings.Contains(filePath, "configFiles") {
			header.Name = filepath.Base(filePath)
		} else {
			header.Name = filePath
		}

		// write the header
		if err := archive.WriteHeader(header); err != nil {
			return fmt.Errorf("error writing header %v\n", err)

		}

		fd, err := os.Open(filePath)
		if err != nil {
			fd.Close()
			return fmt.Errorf("error opening %v\n", err)

		}

		if _, err := io.Copy(archive, fd); err != nil {
			fd.Close()
			return fmt.Errorf("error copying %v\n", err)

		}

		fd.Close()

	}

	return nil

}

func createZipArchive(archiveSpec *ArchiveSpec) error {

	filesToArchive := archiveSpec.files

	createDirectory(archiveSpec.Output)

	output, err := os.Create(filepath.Join(archiveSpec.Output, archiveSpec.Name))

	if err != nil {
		return fmt.Errorf("unable to create %v", output)
	}

	archive := zip.NewWriter(output)
	defer archive.Close()

	for filePath, _ := range filesToArchive {

		info, _ := os.Stat(filePath)

		// create a new dir/file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Use the base name if adding a configuration file
		// to avoid putting the name of the temporary directory
		// in the archive
		if strings.Contains(filePath, "configFiles") {
			header.Name = filepath.Base(filePath)
		} else {
			header.Name = filePath
		}

		// write the header
		zipWriter, err := archive.CreateHeader(header)

		if err != nil {
			return fmt.Errorf("error writing header %v\n", err)

		}

		fd, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("error opening %v\n", err)
			fd.Close()
		}

		if _, err := io.Copy(zipWriter, fd); err != nil {
			return fmt.Errorf("error copying %v\n", err)
		}

		fd.Close()

	}

	return nil

}

func extractFromZip(restore *RestoreSpec) error {

	archive, err := zip.OpenReader(restore.Name)
	if err != nil {
		return fmt.Errorf("opening archive %v", restore.Name)
	}
	defer archive.Close()

	// Find the files to extract
	for _, fh := range archive.File {

		baseName := filepath.Base(fh.Name)

		match := false

		for _, filter := range restore.Filters {

			if ok, _ := regexp.MatchString(filter, baseName); ok {
				match = true
			}
		}

		if len(restore.Filters) == 0 {
			match = true
		}

		if !match {
			continue
		}

		// Restore to the original path if a
		// directory is not specified
		if len(restore.Directory) == 0 {
			restore.Directory = filepath.Dir(fh.Name)
		}

		// Make sure the output directory exists
		createDirectory(restore.Directory)

		outputFilePath := filepath.Join(restore.Directory, baseName)

		outputFile, err := os.Create(outputFilePath)
		if err != nil {
			outputFile.Close()
			return fmt.Errorf("can not create %v", outputFilePath)
		}

		ofh, err := fh.Open()
		if err != nil {
			return fmt.Errorf("can not open %v", fh.Name)
		}
		_, err = io.Copy(outputFile, ofh)
		if err != nil {
			return fmt.Errorf("can not copy %v to %v", fh.Name, outputFilePath)
		}

		outputFile.Close()
		ofh.Close()
	}

	return nil

}

func extractFromTarGz(restore *RestoreSpec) error {

	archive, err := os.Open(restore.Name)
	if err != nil {
		return fmt.Errorf("opening archive %v", restore.Name)
	}
	defer archive.Close()

	gzReader, err := gzip.NewReader(archive)

	if err != nil {
		return fmt.Errorf("gzip can not open %v", archive)

	}

	tarReader := tar.NewReader(gzReader)

	defer gzReader.Close()

	// Find the files to extract
	for {

		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("can not read tar header")

		}

		// Skip directories
		if header.Typeflag == tar.TypeDir {
			continue
		}

		baseName := filepath.Base(header.Name)

		match := false

		for _, filter := range restore.Filters {

			if ok, _ := regexp.MatchString(filter, baseName); ok {
				match = true
			}
		}

		if len(restore.Filters) == 0 {
			match = true
		}

		if !match {
			continue
		}

		// Restore to the original path if a
		// directory is not specified
		if len(restore.Directory) == 0 {
			restore.Directory = filepath.Dir(header.Name)
		}

		// Make sure the output directory exists
		createDirectory(restore.Directory)

		outputFilePath := filepath.Join(restore.Directory, baseName)

		outputFile, err := os.Create(outputFilePath)
		if err != nil {
			outputFile.Close()
			return fmt.Errorf("can not create %v", outputFilePath)
		}

		_, err = io.Copy(outputFile, tarReader)
		if err != nil {
			return fmt.Errorf("can not copy %v to %v", baseName, outputFilePath)
		}

		outputFile.Close()

	}

	return nil

}

func sliceToMap(items []string) map[string]bool {

	output := make(map[string]bool)

	for _, value := range items {

		if _, ok := output[value]; !ok {
			output[value] = true
		}

	}

	return output
}

func mapToSlice(items map[string]bool) []string {

	var dst []string

	for item, _ := range items {
		dst = append(dst, item)
	}

	return dst

}

func concatMaps(src map[string]bool, dst map[string]bool) {

	for item, _ := range src {

		if _, ok := dst[item]; !ok {
			dst[item] = true
		}

	}

}

func createDirectory(directoryName string) error {

	// Make sure the output directory exists
	if err := os.MkdirAll(directoryName, 0777); err != nil {
		return fmt.Errorf("creating directory %v", directoryName)
	}

	// Work around for umask issues when
	// calling os.MkdirAll
	os.Chmod(directoryName, 0777)

	return nil
}

func pathExists(path string) bool {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true

}
