package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"phenix-apps/util"
	"phenix-apps/util/mmcli"
	"phenix/api/cluster"
	"phenix/types"

	"github.com/mitchellh/mapstructure"
)

type ArchiveSpec struct {
	Name      string   `json:"name" mapstructure:"name"`
	Directory string   `json:"directory" mapstructure:"directory"`
	Filters   []string `json:"filters" mapstructure:"filters"`
	Cleanup   bool     `json:"cleanup" mapstructure:"cleanup"`
	Type      string   `json:"type" mapstructure:"type"`
	Output    string   `json:"output" mapstructure:"output"`

	// Internal use for keeping track of the files
	// being put in the archive
	files map[string]bool
}

type RestoreSpec struct {
	Name      string   `json:"name" mapstructure:"name"`
	Directory string   `json:"directory" mapstructure:"directory"`
	Filters   []string `json:"filters" mapstructure:"filters"`
}

type ArchiveOptions struct {
	Archives    []*ArchiveSpec `json:"archives" mapstructure:"archives"`
	Retrievals  []*RestoreSpec `json:"retrievals" mapstructure:"retrievals"`
	RestorePath string         `json:"restore_path" mapstructure:"restore_path"`

	// internal use

	// default name with timestamp
	// for consistency as the timestamp
	// will change with each call
	defaultArchiveName string

	// location of the minimega files
	// directory.  Used to infer other
	// directory locations
	mmFilesDirectory string

	// Needed to help grab any files
	// on a remote node to the headnode
	expName             string
	remoteFilesObtained bool
}

var (
	logger          *log.Logger
	phenixLocation  string
	placeholders    = regexp.MustCompile(`(?i)[<]([^<>]+)[>]`)
	angleBrackets   = regexp.MustCompile(`[<>]`)
	restoreTime     = regexp.MustCompile(`([0-9][0-9\-_]+)`)
	startTimeRe     = regexp.MustCompile(`[\d-]+[T][\d-:]+`)
	globalTimestamp = timestamp()
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

	logger = log.New(out, " archive-experiment ", log.Ldate|log.Ltime|log.Lmsgprefix)

	if len(os.Args) != 2 {
		logger.Fatal("incorrect amount of args provided")
	}

	body, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		logger.Fatal("unable to read JSON from STDIN")
	}

	stage := os.Args[1]

	if stage != "configure" && stage != "cleanup" {		
		fmt.Print(string(body))
		return
	}

	exp, err := DecodeExperiment(body)
	if err != nil {
		logger.Fatalf("decoding experiment: %v", err)
	}

	switch stage {
	case "configure":
		if err := configure(exp); err != nil {
			logger.Fatalf("failed to execute configure stage: %v", err)
		}
	case "cleanup":
		if err := cleanup(exp); err != nil {
			logger.Fatalf("failed to execute cleanup stage: %v", err)
		}
	}

	body, err = json.Marshal(exp)
	if err != nil {
		logger.Fatalf("unable to convert experiment to JSON")
	}

	
	fmt.Print(string(body))
}

func configure(exp *types.Experiment) error {

	options := getArchiveOptions(exp)

	// Make sure that there is a desire to restore
	if options == nil {
		return nil
	}

	if len(options.Retrievals) == 0 {
		logger.Print("No retrievals were specified")
		return nil
	}

	phenixLocation = getParentProcLocation()
	options.expName = exp.Spec.ExperimentName()

	if len(options.RestorePath) == 0 {
		options.RestorePath = "/phenix/configurations"
	}

	// Clear the restoration path
	os.RemoveAll(options.RestorePath)

	// Add the restore specification to obtain
	// the experiment configuration files
	restoreConfigs := getExperimentRetrieval(options)

	for _, restoreSpec := range options.Retrievals {

		// Skip over specifications wihtout an archive name
		if len(restoreSpec.Name) == 0 {
			continue
		}

		// Skip over specifications that no longer exist
		if !pathExists(restoreSpec.Name) {
			logger.Printf("%v no longer exists", restoreSpec.Name)
			continue
		}

		if strings.Contains(restoreSpec.Name, "tar.gz") {
			extractFromTarGz(restoreSpec)

			if !restoreTime.MatchString(restoreSpec.Name) {
				continue
			}

			// Try to extract the experiment configurations
			restoreConfigs.Name = restoreSpec.Name
			if err := extractFromTarGz(restoreConfigs); err != nil {
				logger.Printf("unable to extract from %v error:%v", restoreSpec.Name, err)
			}
		} else {
			extractFromZip(restoreSpec)

			if !restoreTime.MatchString(restoreSpec.Name) {
				continue
			}

			// Try to extract the experiment configurations
			restoreConfigs.Name = restoreSpec.Name
			if err := extractFromZip(restoreConfigs); err != nil {
				logger.Printf("unable to extract from %v", restoreSpec.Name)
			}
		}

	}

	experimentConfigs := getFileList(options.RestorePath, []string{"experiment.yml"})

	var expConfigPath string

	if experimentConfigs != nil {
		expConfigPath = mapToSlice(experimentConfigs)[0]
	}

	logger.Printf("Experiment Configuration:%v", expConfigPath)

	if len(expConfigPath) == 0 {
		return nil
	}

	if !pathExists(expConfigPath) {
		logger.Printf("%v no longer exists", expConfigPath)
		return nil
	}

	restoredExp, err := DecodeExperimentFromFile(expConfigPath)

	if err != nil {
		return err
	}

	savedTime := getRestoreTime(filepath.Dir(restoreConfigs.Name))

	logger.Printf("Restored:%v", restoredExp.Spec.ExperimentName())
	restoreExperiment(restoredExp.Spec.ExperimentName(), expConfigPath, savedTime)

	return nil

}

func cleanup(exp *types.Experiment) error {

	options := getArchiveOptions(exp)

	// Make sure that an archive was specified
	if options == nil {
		logger.Print("no experiment-archive apps were found")
		return nil
	}

	phenixLocation = getParentProcLocation()

	options.expName = exp.Spec.ExperimentName()
	options.defaultArchiveName = defaultArchiveName(options.expName)
	options.mmFilesDirectory, _ = getMMFilesDirectory(options.expName)

	// Add the experiment configuration files to an archive
	addExpConfigFiles(exp, options)

	archivesAdded := make(map[string]bool)

	for _, archive := range options.Archives {		

		// If no archive name has been specified
		// assign the default name
		if len(archive.Name) == 0 {
			archive.Name = options.defaultArchiveName
		} else {
			// Replace any variables defined
			if strings.Contains(archive.Name, "<") {
				archive.Name = replacePlaceholders(archive.Name, options.expName)
			}
		}

		// Skip archives already processed
		if _, ok := archivesAdded[archive.Name]; ok {
			logger.Printf("skipping %v already added", archive.Name)
			continue
		}

		if strings.ToLower(archive.Directory) == "experiment_directory" {
			archive.Directory = fmt.Sprintf("%s/%s/files", options.mmFilesDirectory, exp.Spec.ExperimentName())
			logger.Printf("Archive Output:%v", archive.Directory)
		}

		// Get all the archive files
		getArchiveFiles(archive, options)
		archivesAdded[archive.Name] = true

		// Do not create an empty archive
		if len(archive.files) == 0 {
			logger.Printf("skipping %v empty files", archive.Name)
			continue
		}

		// Set the default type to "zip"
		if len(archive.Type) == 0 {
			archive.Type = "zip"
		}

		// Set the default output location
		if len(archive.Output) == 0 {
			archive.Output = fmt.Sprintf("/phenix/Archives/%s", options.defaultArchiveName)
		} else {
			// Replace any variables defined
			if strings.Contains(archive.Output, "<") {
				archive.Output = replacePlaceholders(archive.Output, options.expName)
			}
		}

		// Archive by the specified type
		switch strings.ToLower(archive.Type) {
		case "targz":
			{
				if !strings.Contains(archive.Name, ".") {
					archive.Name = fmt.Sprintf("%s.%s", archive.Name, "tar.gz")
				}

				if err := createTarGz(archive); err != nil {
					logger.Printf("unable to create tar gz %v", err)
					return fmt.Errorf("unable to create tar gz %v", err)

				}

			}
		case "zip":
			{
				if !strings.Contains(archive.Name, ".") {
					archive.Name = fmt.Sprintf("%s.%s", archive.Name, "zip")
				}

				if err := createZipArchive(archive); err != nil {
					logger.Printf("unable to create zip %v", err)
					return fmt.Errorf("unable to create zip %v", err)
				}

			}
		}

		os.Chmod(filepath.Join(archive.Output, archive.Name), 0777)
		if archive.Cleanup {
			removeArchiveFiles(archive, options.mmFilesDirectory)
		}

	}

	return nil

}

func getArchiveFiles(archive *ArchiveSpec, options *ArchiveOptions) {

	// If the specified directory is a subdirectory of
	// the minimega files directory, move all files to the
	// headnode
	if strings.Contains(archive.Directory, options.mmFilesDirectory) {
		getRemoteExpFiles(options)
	}

	logger.Printf("Directory:%v Filters:%v", archive.Directory, archive.Filters)
	archive.files = getFileList(archive.Directory, archive.Filters)

	// Get a list of files to put into the archive
	// Archives with the same name will be put into the
	// same archive
	for _, archiveSpec := range options.Archives {

		// If no archive name has been specified
		// assign the default name
		if len(archiveSpec.Name) == 0 {
			archiveSpec.Name = options.defaultArchiveName
		}

		// Skip archives that are not the same
		if archiveSpec.Name != archive.Name {
			continue
		}

		// If the specified directory is a subdirectory of
		// the minimega files directory, move all files to the
		// headnode
		if strings.Contains(archive.Directory, options.mmFilesDirectory) {
			getRemoteExpFiles(options)
		}

		archiveSpecFiles := getFileList(archiveSpec.Directory, archiveSpec.Filters)

		concatMaps(archiveSpecFiles, archive.files)

		// Make the cleanup flag consistent
		if archive.Cleanup == false {
			if archiveSpec.Cleanup == true {
				archive.Cleanup = true
			}
		} else {
			archiveSpec.Cleanup = true
		}

		// Make the archive types consistent
		archiveSpec.Type = archive.Type

		// Keep the output directories consistent
		if len(archive.Output) == 0 {
			archive.Output = archiveSpec.Output
		} else {
			archiveSpec.Output = archive.Output
		}

	}

}

func getRemoteExpFiles(options *ArchiveOptions) {

	// No need to try to obtain the experiemnt
	// files multiple times
	if options.remoteFilesObtained {
		return
	}

	headNodeFiles := getHeadNodeFiles(options.mmFilesDirectory)
	expFiles, _ := cluster.GetExperimentFileNames(options.expName)
	headNode, _ := os.Hostname()

	for _, path := range expFiles {

		// Copy all experiment files to the
		// headnode
		if _, ok := headNodeFiles[path]; !ok {
			if err := cluster.CopyFile(path, headNode, nil); err != nil {
				logger.Printf("copying %s to %s", path, headNode)
			}

			// Add the file to the headnode files
			headNodeFiles[path] = true

		}

	}

	options.remoteFilesObtained = true

}

func getHeadNodeFiles(directory string) map[string]bool {

	return getFileList(directory, []string{})

}

func defaultArchiveName(expName string) string {

	return fmt.Sprintf("%s_%s", expName, timestamp())
}

func timestamp() string {

	refTime := time.Now()
	return fmt.Sprintf("%d-%02d-%02d_%02d%02d", refTime.Year(), refTime.Month(),
		refTime.Day(), refTime.Hour(), refTime.Minute())
}

func getMMFilesDirectory(expName string) (string, error) {

	cmd := mmcli.NewCommand(mmcli.Namespace(expName))
	cmd.Command = "vm info"
	cmd.Columns = []string{"host", "name", "id", "disks"}

	status := mmcli.RunTabular(cmd)

	if len(status) == 0 {
		logger.Print("unable to get minimega files directory")
		return "", fmt.Errorf("unable to get minimega files directory")
	}

	// The location of any snapshot should point to the
	// minimega "files" directory
	snapshotPath := strings.Split(status[0]["disks"], ",")[0]
	return filepath.Dir(snapshotPath), nil

}

func getArchiveOptions(exp *types.Experiment) *ArchiveOptions {

	app := util.ExtractApp(exp.Spec.Scenario(), "archive-experiment")

	if app == nil {
		logger.Printf("unable to extract %v app", "archive-experiment")
		return nil
	}

	var options ArchiveOptions

	if err := mapstructure.Decode(app.Metadata(), &options); err != nil {
		logger.Printf("mapsructure can't decode %v", app.Metadata())
		return nil
	}

	return &options

}

func removeArchiveFiles(archive *ArchiveSpec, mmFilesDirectory string) {

	for filePath, _ := range archive.files {

		// If the file is part of the minimega path,
		// then delete the file on all nodes in the cluster
		if strings.Contains(filePath, mmFilesDirectory) {
			relativePath, _ := filepath.Rel(mmFilesDirectory, filePath)

			// minimega file management api works off of
			// relative paths from the minimega files directory
			cluster.DeleteFile(relativePath)
		} else {
			os.Remove(filePath)
		}

	}

}

func addExpConfigFiles(exp *types.Experiment, options *ArchiveOptions) {

	// Create a temporary directory to write
	// the experiment configuration files
	dir := filepath.Join(os.TempDir(), "configFiles")

	// Get and write the topology configuration file
	writeConfigurationFile("topology", dir, exp)

	// Get and write the scenario configuration file
	writeConfigurationFile("scenario", dir, exp)

	// Get and write the scenario configuration file
	writeConfigurationFile("experiment", dir, exp)

	// Create and add the archive spec to the list of archives
	archive := &ArchiveSpec{
		Name:      options.defaultArchiveName,
		Directory: dir,
		Cleanup:   true,
		Type:      "targz",
	}

	options.Archives = append(options.Archives, archive)

}

func getParentProcLocation() string {

	var out bytes.Buffer
	cmd := exec.Command("ls", "-alht", fmt.Sprintf("/proc/%d/exe", os.Getppid()))
	cmd.Stdout = &out

	logger.Printf("Command:%v", cmd.String())
	err := cmd.Run()

	if err != nil {
		logger.Printf("Error:%v", err)
		return ""
	}

	binPath := strings.Split(out.String(), "-> ")[1]
	binPath = strings.ReplaceAll(binPath, "\n", "")

	logger.Printf("Output:%v", binPath)

	return binPath

}

func createConfigurationFile(configName, outputPath string) error {

	// Make sure phenix is running
	if len(phenixLocation) == 0 {
		logger.Print("phenix was not found running")
		return fmt.Errorf("phenix was not found running")
	}

	createDirectory(filepath.Dir(outputPath))

	var (
		out        bytes.Buffer
		outputFile *os.File
		err        error
	)

	cmd := exec.Command(phenixLocation, "config", "get", configName)
	cmd.Stdout = &out

	err = cmd.Run()

	outputFile, err = os.Create(outputPath)

	if err != nil {
		logger.Printf("can not create %v", outputPath)
		return fmt.Errorf("can not create %v", outputPath)
	}

	_, err = outputFile.WriteString(out.String())
	if err != nil {
		logger.Printf("can not write %v", outputPath)
		return fmt.Errorf("can not write to %v", outputPath)
	}

	outputFile.Close()
	os.Chmod(outputPath, 0777)

	return nil

}

func writeConfigurationFile(configType, outputDir string, exp *types.Experiment) {

	var configName string

	if configType == "experiment" {
		configName = fmt.Sprintf("%s/%s", configType, exp.Spec.ExperimentName())
	} else {
		configName = fmt.Sprintf("%s/%s", configType, exp.Metadata.Annotations[configType])
	}

	filename := fmt.Sprintf("%s_%s.yml", exp.Spec.ExperimentName(), configType)
	outputPath := filepath.Join(outputDir, filename)

	createConfigurationFile(configName, outputPath)

}

func getExperimentRetrieval(options *ArchiveOptions) *RestoreSpec {

	return &RestoreSpec{
		Directory: options.RestorePath,
		Filters:   []string{"(?i)[^\n]+[_][est][^\n]+[.]yml"},
	}
}

func replacePlaceholders(input, expName string) string {

	matches := placeholders.FindAllStringSubmatch(input, -1)

	if matches == nil {
		return input
	}

	logger.Printf("Matches:%v", matches)

	for _, variable := range matches {
		switch variable[1] {
		case "timestamp":
			{
				input = strings.ReplaceAll(input, variable[1], globalTimestamp)
			}
		case "experiment_name":
			{
				input = strings.ReplaceAll(input, variable[1], expName)

			}

		}
	}

	// Replace all the angle brackets
	input = angleBrackets.ReplaceAllString(input, "")

	logger.Printf("Replaced:%v", input)

	return input

}

func getRestoreTime(filename string) string {

	matches := restoreTime.FindAllStringSubmatch(filename, -1)

	if matches == nil {
		return filename
	}

	return matches[0][1]

}

func restoreExperiment(expName, expConfigPath, savedTime string) error {

	// Make sure phenix is running
	if len(phenixLocation) == 0 {
		logger.Print("phenix was not found running")
		return fmt.Errorf("phenix was not found running")
	}

	var newExpName string

	if len(savedTime) == 0 {
		newExpName = expName
	} else {
		newExpName = fmt.Sprintf("%s_%s", expName, savedTime)
	}

	newConfigPath, _ := updateExpConfig(expConfigPath, expName, newExpName)

	logger.Printf("NewConfigPath:%v", newConfigPath)

	cmd := exec.Command(phenixLocation, "config", "create", newConfigPath)

	err := cmd.Run()

	if err != nil {
		logger.Printf("can not create %v", expConfigPath)
		return fmt.Errorf("can not create %v", expConfigPath)
	}

	return nil

}

func updateExpConfig(expConfigPath, oldExpName, expName string) (string, error) {

	if !pathExists(expConfigPath) {
		return "", fmt.Errorf("path %v does not exist", expConfigPath)
	}

	var (
		err    error
		output *os.File
		fh     *os.File
	)

	tmpPath := filepath.Join(filepath.Dir(expConfigPath), fmt.Sprintf("%s.yml", expName))
	output, err = os.Create(tmpPath)
	defer output.Close()

	if err != nil {
		logger.Printf("unable to create %v", tmpPath)
		return "", fmt.Errorf("unable to open %v", tmpPath)
	}

	bufferedOut := bufio.NewWriter(output)

	fh, err = os.Open(expConfigPath)
	defer fh.Close()

	if err != nil {
		logger.Printf("unable to open %v", expConfigPath)
		return "", fmt.Errorf("unable to open %v", expConfigPath)
	}

	scanner := bufio.NewScanner(fh)

	for scanner.Scan() {
		line := scanner.Text()

		// Null out the start time so that the
		// restored configuration does not show a
		// status of started
		if strings.Contains(line, "startTime:") {
			line = startTimeRe.ReplaceAllString(line, "") + "\n"
		} else {
			line = strings.ReplaceAll(line, oldExpName, expName) + "\n"
		}

		bufferedOut.WriteString(line)

	}

	bufferedOut.Flush()

	os.Chmod(tmpPath, 0777)

	return tmpPath, nil

}

func deleteExpConfig(expName string) error {

	configName := fmt.Sprintf("experiment/%s", expName)
	cmd := exec.Command(phenixLocation, "config", "delete", configName)

	err := cmd.Run()

	if err != nil {
		logger.Printf("can not delete %v", configName)
		return fmt.Errorf("can not delete %v", configName)
	}

	return nil
}
