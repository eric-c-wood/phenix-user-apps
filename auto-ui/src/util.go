package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	InfoLogger    *log.Logger
	WarningLogger *log.Logger
	ErrorLogger   *log.Logger
)

func init() {

	logDirectory := getDirectory("Logs")

	fmt.Println(logDirectory)

	if len(logDirectory) == 0 {
		return
	}

	logFilePath := filepath.Join(logDirectory, "Auto-Ui.log")
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(logFile, "INFO:", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(logFile, "WARN:", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(logFile, "ERROR:", log.Ldate|log.Ltime|log.Lshortfile)
}

func getDirectory(dirName string) string {

	refDirectory := filepath.Join(getParentDirectory(), dirName)

	if err := createDirectory(refDirectory); err != nil {
		return ""
	}

	if InfoLogger != nil {
		InfoLogger.Printf("getDirectory - Reference:%v", refDirectory)
	}

	return refDirectory
}

func getParentDirectory() string {

	wd, err := os.Executable()

	if err != nil {
		return ""
	}

	return filepath.Dir(filepath.Dir(wd))
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

func getTimestamp() string {

	return time.Now().Format(time.DateTime)
}
