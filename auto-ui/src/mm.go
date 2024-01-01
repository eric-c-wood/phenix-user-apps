package main

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"phenix/util/mm/mmcli"
)

func getFileFromHost(hostName, srcPath, dstPath string) error {

	cmd := "scp"
	cmdPath, err := exec.LookPath(cmd)

	if err != nil {
		return fmt.Errorf("Cannot find %s", cmd)
	}

	srcLocation := fmt.Sprintf("%s:%s", hostName, srcPath)

	scpCmd := exec.Command(cmdPath, srcLocation, dstPath)

	if err := scpCmd.Run(); err != nil {
		ErrorLogger.Printf("Can not get %s from %s:%v", srcPath, hostName, err)
		return fmt.Errorf("Can not get %s from %s:%v\n", srcPath, hostName, err)
	}

	return nil
}

func takeScreenshot(expName, vmName, dstPath string) error {

	if !pathExists(filepath.Dir(dstPath)) {
		createDirectory(filepath.Dir(dstPath))
	}

	cmd := mmcli.NewNamespacedCommand(expName)
	cmd.Command = fmt.Sprintf("vm screenshot %s file %s", vmName, dstPath)

	if err := mmcli.ErrorResponse(mmcli.Run(cmd)); err != nil {
		ErrorLogger.Printf("Can not take screenshot of VM %s in namespace %s: %v", vmName, expName, err)
		return fmt.Errorf("Can not take screenshot of VM %s in namespace %s: %v\n", vmName, expName, err)
	}

	return nil
}

func getHost(expName, vmName string) (string, error) {

	cmd := mmcli.NewNamespacedCommand(expName)
	cmd.Command = "vm info"
	cmd.Columns = []string{"host"}
	cmd.Filters = []string{"name=" + vmName}

	status := mmcli.RunTabular(cmd)

	if len(status) == 0 {
		return "", fmt.Errorf("Host %s not found\n", vmName)
	}

	return status[0]["host"], nil
}
