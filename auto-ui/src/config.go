package main

import (
	"bufio"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Action struct {
	Name       string      `yaml:"name"`
	Parameters interface{} `yaml:"parameters"`
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

func readConfig(filePath string) (*Playbook, error) {

	configFile, err := os.ReadFile(filePath)

	if err != nil {
		ErrorLogger.Printf("Reading config file %s:%v", filePath, err)
		return nil, fmt.Errorf("Reading config file %s:%v\n", filePath, err)
	}

	var playbook Playbook

	err = yaml.Unmarshal(configFile, &playbook)

	if err != nil {
		ErrorLogger.Printf("Decoding config file %s:%v", filePath, err)
		return nil, fmt.Errorf("Decoding config file %s:%v\n", filePath, err)
	}

	return &playbook, nil

}

func writeConfig(playbook *Playbook, filePath string) error {

	configFile, err := yaml.Marshal(&playbook)

	if err != nil {
		ErrorLogger.Printf("Decoding playbook structure %v", err)
		return fmt.Errorf("Decoding playbook structure %v\n", err)
	}

	outputFile, err := os.Create(filePath)
	defer outputFile.Close()

	if err != nil {
		ErrorLogger.Printf("Cannot create %s:%v", filePath, err)
		return fmt.Errorf("Cannot create %s:%v\n", filePath, err)
	}

	bufWriter := bufio.NewWriter(outputFile)

	bufWriter.WriteString(string(configFile))

	bufWriter.Flush()

	return nil

}
