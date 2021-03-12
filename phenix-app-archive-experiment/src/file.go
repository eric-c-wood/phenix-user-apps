package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"phenix-apps/util"
	"phenix-apps/util/mmcli"
)

type CopyStatus func(float64)

func GetExperimentFileNames(exp string) ([]string, error) {
	// Using a map here to weed out duplicates.
	matches := make(map[string]struct{})

	dir := fmt.Sprintf("/%s/files", exp)

	// First get file listings from mesh, then from headnode.
	commands := []string{
		"mesh send all file list " + dir,
		"file list " + dir,
	}

	cmd := mmcli.NewCommand()

	for _, command := range commands {
		cmd.Command = command

		for _, row := range mmcli.RunTabular(cmd) {
			// Only looking for files.
			if row["dir"] != "" {
				continue
			}

			name := filepath.Base(row["name"])
			matches[name] = struct{}{}
		}
	}

	var files []string

	for f := range matches {
		files = append(files, f)
	}

	return files, nil
}

func CopyFile(path, dest string, status CopyStatus) error {
	cmd := mmcli.NewCommand()

	if util.IsHeadnode(dest) {
		cmd.Command = fmt.Sprintf(`file get %s`, path)
	} else {
		cmd.Command = fmt.Sprintf(`mesh send %s file get %s`, dest, path)
	}

	if err := mmcli.ErrorResponse(mmcli.Run(cmd)); err != nil {
		return fmt.Errorf("copying file to destination: %w", err)
	}

	if util.IsHeadnode(dest) {
		cmd.Command = fmt.Sprintf(`file status`)
	} else {
		cmd.Command = fmt.Sprintf(`mesh send %s file status`, dest)
	}

	for {
		var found bool

		for _, row := range mmcli.RunTabular(cmd) {
			if row["filename"] == path {
				comp := strings.Split(row["completed"], "/")

				parts, _ := strconv.ParseFloat(comp[0], 64)
				total, _ := strconv.ParseFloat(comp[1], 64)

				if status != nil {
					status(parts / total)
				}

				found = true
				break
			}
		}

		// If the file is done transferring, then it will not have been present in
		// the results from `file status`.
		if !found {
			break
		}
	}

	return nil
}

func DeleteFile(path string) error {
	// First delete file from mesh, then from headnode.
	commands := []string{"mesh send all file delete", "file delete"}

	cmd := mmcli.NewCommand()

	for _, command := range commands {
		cmd.Command = fmt.Sprintf("%s %s", command, path)

		if err := mmcli.ErrorResponse(mmcli.Run(cmd)); err != nil {
			return fmt.Errorf("deleting file from cluster nodes: %w", err)
		}
	}

	return nil
}
