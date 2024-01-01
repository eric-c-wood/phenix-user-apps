package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"phenix/store"
	"phenix/types"
	ifaces "phenix/types/interfaces"
	"phenix/types/version"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

type Action struct {
	Name       string        `yaml:"name"`
	Parameters []interface{} `yaml:"parameters"`
}

type Script struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty`
	Date        string   `yaml:"date,omitempty`
	Author      string   `yaml:"author,omitempty`
	Actions     []Action `yaml:"actions"`
}

type Playbook struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty`
	Date        string   `yaml:"date,omitempty`
	Author      string   `yaml:"author,omitempty`
	Scripts     []Script `yaml:"scripts"`
}

func DecodeExperiment(body []byte) (*types.Experiment, error) {
	var mapper map[string]interface{}

	if err := json.Unmarshal(body, &mapper); err != nil {
		return nil, fmt.Errorf("unable to parse JSON: %w", err)
	}

	iface, err := version.GetVersionedSpecForKind("Experiment", "v1")
	if err != nil {
		return nil, fmt.Errorf("getting versioned spec for experiment: %w", err)
	}

	if err := mapstructure.Decode(mapper["spec"], &iface); err != nil {
		return nil, fmt.Errorf("decoding versioned spec: %w", err)
	}

	spec, ok := iface.(ifaces.ExperimentSpec)
	if !ok {
		return nil, fmt.Errorf("invalid experiment spec")
	}

	iface, err = version.GetVersionedStatusForKind("Experiment", "v1")
	if err != nil {
		return nil, fmt.Errorf("getting versioned status for experiment: %w", err)
	}

	if err := mapstructure.Decode(mapper["status"], &iface); err != nil {
		return nil, fmt.Errorf("decoding versioned status: %w", err)
	}

	status, ok := iface.(ifaces.ExperimentStatus)
	if !ok {
		return nil, fmt.Errorf("invalid experiment status")
	}

	var metadata store.ConfigMetadata

	if err := mapstructure.Decode(mapper["metadata"], &metadata); err != nil {
		return nil, fmt.Errorf("decoding experiment metadata: %w", err)
	}

	return &types.Experiment{Spec: spec, Status: status, Metadata: metadata}, nil
}

func DecodeExperimentFromFile(filePath string) (*types.Experiment, error) {

	var mapper map[string]interface{}

	body, err := os.ReadFile(filePath)
	if err != nil {
		logger.Fatalf("unable to read YAML from %s", filePath)
	}

	if err := yaml.Unmarshal(body, &mapper); err != nil {
		return nil, fmt.Errorf("unable to parse YAML: %w", err)
	}

	iface, err := version.GetVersionedSpecForKind("Experiment", "v1")
	if err != nil {
		return nil, fmt.Errorf("getting versioned spec for experiment: %w", err)
	}

	if err := mapstructure.Decode(mapper["spec"], &iface); err != nil {
		return nil, fmt.Errorf("decoding versioned spec: %w", err)
	}

	spec, ok := iface.(ifaces.ExperimentSpec)
	if !ok {
		return nil, fmt.Errorf("invalid experiment spec")
	}

	iface, err = version.GetVersionedStatusForKind("Experiment", "v1")
	if err != nil {
		return nil, fmt.Errorf("getting versioned status for experiment: %w", err)
	}

	if err := mapstructure.Decode(mapper["status"], &iface); err != nil {
		return nil, fmt.Errorf("decoding versioned status: %w", err)
	}

	status, ok := iface.(ifaces.ExperimentStatus)
	if !ok {
		return nil, fmt.Errorf("invalid experiment status")
	}

	var metadata store.ConfigMetadata

	if err := mapstructure.Decode(mapper["metadata"], &metadata); err != nil {
		return nil, fmt.Errorf("decoding experiment metadata: %w", err)
	}

	return &types.Experiment{Spec: spec, Status: status, Metadata: metadata}, nil

}

func readConfig(filePath string) (*Playbook, error) {

	configFile, err := os.ReadFile(filePath)

	if err != nil {
		logger.Printf("Reading config file %s:%v", filePath, err)
		return nil, fmt.Errorf("Reading config file %s:%v\n", filePath, err)
	}

	var playbook Playbook

	err = yaml.Unmarshal(configFile, &playbook)

	if err != nil {
		logger.Printf("Decoding config file %s:%v", filePath, err)
		return nil, fmt.Errorf("Decoding config file %s:%v\n", filePath, err)
	}

	return &playbook, nil

}

func writeConfig(playbook *Playbook, filePath string) error {

	configFile, err := yaml.Marshal(&playbook)

	if err != nil {
		logger.Printf("Decoding playbook structure %v", err)
		return fmt.Errorf("Decoding playbook structure %v\n", err)
	}

	outputFile, err := os.Create(filePath)
	defer outputFile.Close()

	if err != nil {
		logger.Printf("Cannot create %s:%v", filePath, err)
		return fmt.Errorf("Cannot create %s:%v\n", filePath, err)
	}

	bufWriter := bufio.NewWriter(outputFile)

	bufWriter.WriteString(string(configFile))

	bufWriter.Flush()

	return nil

}

func getScripts(tmplPath string, node *ifaces.NodeSpec) ([]Script, error) {

	buf, err := applyTemplate(tmplPath, node)

	if err != nil {
		logger.Printf("%v", err)
		return nil, fmt.Errorf("%v", err)
	}

	var playbook Playbook

	err = yaml.Unmarshal(buf, &playbook)

	if err != nil {
		logger.Printf("error decoding buffer:%v", err)
		return nil, fmt.Errorf("error decoding buffer:%v", err)
	}

	return playbook.Scripts, nil
}

func applyTemplate(tmplPath string, node *ifaces.NodeSpec) ([]byte, error) {

	var tmp bytes.Buffer

	if err := GenerateFromTemplate(tmplPath, node, &tmp); err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	return tmp.Bytes(), nil
}
