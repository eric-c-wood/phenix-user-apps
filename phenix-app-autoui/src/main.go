package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"phenix-apps/util"
	"phenix/types"

	"github.com/mitchellh/mapstructure"
)

type AppOptions struct {
	templateDirectory string `json:"template_directory" mapstructure:"template_directory"`
}

type fileList struct {
	files   map[string]bool
	filters []string
}

var (
	logger                   *log.Logger
	defaultTemplateDirectory = "/phenix/mako"
	pidList                  []string
)

func main() {

	out := os.Stderr

	if env, ok := os.LookupEnv("PHENIX_LOG_FILE"); ok {
		var err error

		out, err = os.OpenFile(env, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal("unable to open phenix log file for writing")
		}

		defer out.Close()
	}

	logger = log.New(out, " auto-ui ", log.Ldate|log.Ltime|log.Lshortfile)

	if len(os.Args) != 2 {
		logger.Fatal("incorrect amount of args provided")
	}

	body, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		logger.Fatal("unable to read JSON from STDIN")
	}

	stage := os.Args[1]

	exp, err := DecodeExperiment(body)
	if err != nil {
		logger.Fatalf("decoding experiment: %v", err)
	}

	switch stage {
	case "pre-start":
		if err := preStart(exp); err != nil {
			logger.Fatalf("failed to execute pre-start stage: %v", err)
		}
	case "post-start":
		if err := postStart(exp); err != nil {
			logger.Fatalf("failed to execute post-start stage: %v", err)
		}
	case "cleanup":
		if err := cleanup(exp); err != nil {
			logger.Fatalf("failed to execute cleanup stage: %v", err)
		}
	default:
		fmt.Print(string(body))
		return
	}

	body, err = json.Marshal(exp)
	if err != nil {
		logger.Fatalf("unable to convert experiment to JSON")
	}

	fmt.Print(string(body))
}

func preStart(exp *types.Experiment) error {

	var (
		templatePath      string
		tmplName          string
		templateDirectory string
		vmName            string
		playbook          *Playbook
	)

	appOptions := getAppOptions(exp)

	templateDirectory = defaultTemplateDirectory
	if appOptions != nil {
		if len(appOptions.templateDirectory) > 0 {
			templateDirectory = appOptions.templateDirectory
		}
	}

	//Make sure the template directory exists
	createDirectory(templateDirectory)
	createDirectory(exp.Spec.BaseDir())

	for _, node := range exp.Spec.Topology().Nodes() {

		// Skip VMs that are not set to boot
		if node.General().DoNotBoot() != nil {
			if *node.General().DoNotBoot() {
				continue
			}

		}

		if len(node.Labels()) == 0 {
			continue
		}

		if _, ok := node.Labels()["ui-scripts"]; !ok {
			continue
		}

		vmName = node.General().Hostname()

		scripts := strings.Split(node.Labels()["ui-scripts"], ",")

		// Add playbook header
		playbook = &Playbook{
			Name:        fmt.Sprintf("%s-auto-gen", vmName),
			Description: fmt.Sprintf("Auto generated scripts for %s", vmName),
			Author:      "auto-ui-user-app",
			Date:        getTimestamp(),
		}

		for _, script := range scripts {

			// Build script template file path
			tmplName = fmt.Sprintf("%s.tmpl", script)
			templatePath = filepath.Join(templateDirectory, tmplName)

			// Make sure that the template path exists
			if !pathExists(templatePath) {
				continue
			}

			// Add all scripts to new list of scripts
			playBookScripts, err := getScripts(templatePath, &node)

			if err != nil {
				logger.Printf("%v", err)
				continue
			}

			playbook.Scripts = append(playbook.Scripts, playBookScripts...)

		}

		ymlOut := filepath.Join(fmt.Sprintf("%s/startup/", exp.Spec.BaseDir()),
			fmt.Sprintf("%s_ui_scripts.yml", vmName))

		// Write out new playbook file
		writeConfig(playbook, ymlOut)

	}

	return nil

}

func postStart(exp *types.Experiment) error {

	// Find all yaml files in the base startup directory

	searchDirectory := fmt.Sprintf("%s/startup", exp.Spec.BaseDir())

	configs := getFileList(searchDirectory, []string{"yml"})

	for config, _ := range configs {

		// Skip temporary yaml files
		if strings.Contains(config, "tmp") {
			continue
		}

		if err := startAutoUi(exp.Spec.ExperimentName(), config); err != nil {
			logger.Printf("error starting auto-ui for %s", config)
			continue
		}
	}

	// Write the list of pids to a file
	autoUiPidsPath := fmt.Sprintf("%s/%s", exp.Spec.BaseDir(), "auto-ui-pids.txt")

	writeFile(autoUiPidsPath, pidList)

	return nil

}

func startAutoUi(expName, configPath string) error {

	// Try to find the auto-ui executable
	autoUiPath, err := exec.LookPath("auto-ui")

	if err != nil {
		logger.Println("unable to locate auto-ui")
		return fmt.Errorf("unable to locate auto-ui")
	}

	cmd := exec.Command(autoUiPath, "-exp", expName, "-config", configPath)

	if err := cmd.Start(); err != nil {
		logger.Printf("unable to start auto-ui:%v", err)
		return fmt.Errorf("unable to start auto-ui:%v", err)
	}

	// Add the pid to the list
	pidList = append(pidList, strconv.Itoa(cmd.Process.Pid))

	return nil

}

func cleanup(exp *types.Experiment) error {

	// Kill any processes that are still running
	autoUiPidsPath := fmt.Sprintf("%s/%s", exp.Spec.BaseDir(), "auto-ui-pids.txt")

	if !pathExists(autoUiPidsPath) {
		return fmt.Errorf("%s does not exist", autoUiPidsPath)
	}

	fh, err := os.Open(autoUiPidsPath)
	defer fh.Close()

	if err != nil {
		logger.Printf("unable to open %v", autoUiPidsPath)
		return fmt.Errorf("unable to open %v", autoUiPidsPath)
	}

	scanner := bufio.NewScanner(fh)

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 {
			continue
		}

		intPid, _ := strconv.Atoi(line)

		proc, err := os.FindProcess(intPid)

		if err != nil {
			logger.Printf("%v", err)
			continue
		}

		if err := proc.Kill(); err != nil {
			logger.Printf("unable to kill %v", intPid)

		}

	}

	return nil
}

func getAppOptions(exp *types.Experiment) *AppOptions {

	app := util.ExtractApp(exp.Spec.Scenario(), "autoui")

	if app == nil {
		logger.Printf("unable to extract %v app", "autoui")
		return nil
	}

	var options AppOptions

	if err := mapstructure.Decode(app.Metadata(), &options); err != nil {
		logger.Printf("mapsructure can't decode %v", app.Metadata())
		return nil
	}

	return &options

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

func getTimestamp() string {

	return time.Now().Format(time.DateTime)
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
		bufferedOut.WriteString(fmt.Sprintf("%s\n", line))
	}

	bufferedOut.Flush()

	return nil

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
