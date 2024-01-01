package main

import (
	"flag"
	"fmt"
	"path/filepath"
)

func main() {

	var (
		expName    string
		yamlConfig string
	)

	flag.StringVar(&expName, "exp", "", "Phenix experiment name")
	flag.StringVar(&yamlConfig, "config", "", "Configuration file defining a linear set of UI actions")
	flag.Parse()

	if len(expName) == 0 {
		fmt.Println("An experiment name must be provided")
		return
	}

	if len(yamlConfig) == 0 {
		fmt.Println("A configuration file must be specified")
		return
	}

	if !filepath.IsAbs(yamlConfig) {
		fmt.Printf("%v:%v\n", expName, yamlConfig)
		//yamlConfig = filepath.Join(getParentDirectory(), yamlConfig)
	}

	if !pathExists(yamlConfig) {
		fmt.Printf("Path %s does not exist\n", yamlConfig)
		return
	}

	if err := processConfig(expName, yamlConfig); err != nil {
		fmt.Printf("%v", err)
	}

}
