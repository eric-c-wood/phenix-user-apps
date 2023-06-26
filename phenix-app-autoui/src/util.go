package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func getDirectory(dirName string) string {

	refDirectory := filepath.Join(getParentDirectory(), dirName)

	if err := createDirectory(refDirectory); err != nil {
		return ""
	}

	fmt.Println(refDirectory)

	return refDirectory
}

func getParentDirectory() string {

	wd, err := os.Getwd()

	if err != nil {
		return ""
	}

	return wd
}

// Tries to extract the attribute from the command "name" specified by pattern
func getCommandAttribute(name, pattern string) (string, error) {

	cmd := "ps"
	psPath, err := exec.LookPath(cmd)

	if err != nil {
		return "", fmt.Errorf("Cannot find %s", cmd)
	}

	cmd = "grep"
	grepPath, err := exec.LookPath(cmd)

	if err != nil {
		return "", fmt.Errorf("Cannot find %s", cmd)
	}

	psCmd := exec.Command(psPath, "-ax")
	psStdout, _ := psCmd.StdoutPipe()
	defer psStdout.Close()

	grepCmd := exec.Command(grepPath, name)
	grepCmd.Stdin = psStdout

	psCmd.Start()

	output, _ := grepCmd.Output()

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	scanner.Split(bufio.ScanLines)

	attributeRe := regexp.MustCompile(pattern)

	var matches []string

	// Try to find the specified attribute
	for scanner.Scan() {

		matches = attributeRe.FindStringSubmatch(scanner.Text())
		if matches == nil {
			continue
		}

		return matches[1], nil
	}

	return "", fmt.Errorf("%s could not be found", name)

}

func getCommandPath(command string) (string, error) {

	result, err := getCommandAttribute(command, `\d+[:]\d+ ([/][^ ]+)`)

	return result, err
}

func getCommandFlag(command, flag string) (string, error) {

	result, err := getCommandAttribute(command, fmt.Sprintf(`%s=([^ ]+)`, flag))

	return result, err
}

func readJson(jsonPath string) (map[string]interface{}, error) {

	if !pathExists(jsonPath) {
		return nil, fmt.Errorf("Path %s does not exist", jsonPath)
	}

	var content map[string]interface{}

	jsonMap, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Printf("%v", err)
	}

	if err := json.Unmarshal(jsonMap, &content); err != nil {
		fmt.Printf("Error:%s", err)
		return nil, fmt.Errorf("Error parsing json:%v", err)
	}

	//fmt.Printf("Contents:%s\n", content)

	return content, nil

}

func pathExists(path string) bool {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true

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

func writeFile(path string, lines []string) error {

	if !pathExists(filepath.Dir(path)) {
		createDirectory(filepath.Dir(path))
	}

	output, err := os.Create(path)
	defer output.Close()

	if err != nil {
		return fmt.Errorf("unable to open %v", path)
	}

	bufferedOut := bufio.NewWriter(output)

	for _, line := range lines {
		bufferedOut.WriteString(line)
	}

	bufferedOut.Flush()

	return nil

}
