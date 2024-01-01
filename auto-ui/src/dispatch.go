package main

import (
	//"bufio"
	"fmt"
	"path/filepath"
	"time"
	//"os"
	//gopkg.in/yaml.v3"
)

func parseSendSpecialKeys(expName string, parameters []interface{}) error {

	var (
		vmName string
		keys   string
	)

	for _, param := range parameters {

		if _, ok := param.(map[string]interface{}); !ok {
			ErrorLogger.Println("Unexpected Type for parameter")
			continue
		}

		for k, v := range param.(map[string]interface{}) {

			switch k {

			case "vm-name":
				vmName, _ = v.(string)
			case "keys":
				keys, _ = v.(string)

			}

		}
	}

	if len(vmName) == 0 {
		ErrorLogger.Println("vm name has a length of zero")
		return fmt.Errorf("vm name has a length of zero")
	}

	if len(keys) == 0 {
		ErrorLogger.Println("keys has a length of zero")
		return fmt.Errorf("keys has a alength of zero")
	}

	sendSpecialKeys(expName, vmName, keys)

	return nil

}

func parseSleep(expName string, parameters []interface{}) error {

	var (
		timeout string
	)

	for _, param := range parameters {

		if _, ok := param.(map[string]interface{}); !ok {
			ErrorLogger.Println("Unexpected Type for parameter")
			continue
		}

		for k, v := range param.(map[string]interface{}) {

			switch k {

			case "timeout":
				timeout, _ = v.(string)
			}

		}
	}

	if len(timeout) == 0 {
		ErrorLogger.Println("timeout has a length of zero")
		return fmt.Errorf("timemout has a length of zero")
	}

	duration, err := time.ParseDuration(timeout)
	if err != nil {
		ErrorLogger.Printf("%v", err)
		return fmt.Errorf("%v\n", err)
	}

	time.Sleep(duration)

	return nil

}

func parseWaitUntil(expName string, parameters []interface{}) error {

	var (
		vmName         string
		referenceImage string
		timeout        string
		threshold      string
	)

	for _, param := range parameters {

		if _, ok := param.(map[string]interface{}); !ok {
			ErrorLogger.Println("Unexpected Type for parameter")
			continue
		}

		for k, v := range param.(map[string]interface{}) {

			switch k {

			case "timeout":
				timeout, _ = v.(string)
			case "vm-name":
				vmName, _ = v.(string)
			case "reference-image":
				referenceImage, _ = v.(string)

			}

		}
	}

	if len(vmName) == 0 {
		ErrorLogger.Println("vm name has a length of zero")
		return fmt.Errorf("vm name has a length of zero")
	}

	if len(referenceImage) == 0 {
		ErrorLogger.Println("reference image has a length of zero")
		return fmt.Errorf("reference image has a alength of zero")
	}

	waitUntil(expName, vmName, filepath.Join(refImageDirectory, referenceImage), timeout, threshold)

	return nil

}

func parseUiTermCommand(expName string, parameters []interface{}) error {

	var (
		uiTerminal termCommand
	)

	for _, param := range parameters {

		if _, ok := param.(map[string]interface{}); !ok {
			ErrorLogger.Println("Unexpected Type for parameter")
			continue
		}

		for k, v := range param.(map[string]interface{}) {

			switch k {

			case "vm-name":
				uiTerminal.vmName, _ = v.(string)
			case "os":
				uiTerminal.osName, _ = v.(string)
			case "terminal-window-image":
				uiTerminal.termWindow, _ = v.(string)
			case "desktop-image":
				uiTerminal.dsktpSymbol, _ = v.(string)
			case "timeout":
				uiTerminal.timeout, _ = v.(string)
			case "threshold":
				uiTerminal.threshold, _ = v.(string)
			case "commands":
				for _, command := range v.([]interface{}) {
					if _, ok := command.(string); !ok {
						continue
					}

					uiTerminal.commands = append(uiTerminal.commands, command.(string))
				}

			}

		}
	}

	InfoLogger.Printf("Commands:%v", uiTerminal.commands)

	convertPaths(&uiTerminal)

	sendGUITermCommands(expName, &uiTerminal)

	return nil

}

func parseLoginLinux(expName string, parameters []interface{}) error {

	var (
		vmName                string
		timeout               string
		threshold             string
		linuxLoginPrompt      string
		linuxAfterLoginPrompt string
		userName              string
		password              string
		credentials           []string
	)

	for _, param := range parameters {

		if _, ok := param.(map[string]interface{}); !ok {
			ErrorLogger.Println("Unexpected Type for parameter")
			continue
		}

		for k, v := range param.(map[string]interface{}) {

			switch k {

			case "vm-name":
				vmName, _ = v.(string)
			case "linux-login-prompt":
				linuxLoginPrompt, _ = v.(string)
			case "linux-after-login-prompt":
				linuxAfterLoginPrompt, _ = v.(string)
			case "user-name":
				userName, _ = v.(string)
			case "password":
				password, _ = v.(string)
			case "timeout":
				timeout, _ = v.(string)
			case "threshold":
				threshold, _ = v.(string)

			}

		}
	}

	if len(vmName) == 0 {
		ErrorLogger.Println("vm name has a length of zero")
		return fmt.Errorf("vm name has a length of zero")
	}

	if len(linuxLoginPrompt) == 0 {
		ErrorLogger.Println("linux-login-prompt image has a length of zero")
		return fmt.Errorf("linux-login-prompt image has a alength of zero")
	}

	linuxLoginPrompt = filepath.Join(refImageDirectory, linuxLoginPrompt)

	if len(linuxAfterLoginPrompt) == 0 {
		ErrorLogger.Println("linux-after-login-prompt image has a length of zero")
		return fmt.Errorf("linux-after-login-prompt image has a alength of zero")
	}

	linuxAfterLoginPrompt = filepath.Join(refImageDirectory, linuxAfterLoginPrompt)

	if len(userName) == 0 {
		ErrorLogger.Println("user-name has a length of zero")
		return fmt.Errorf("user-name has a alength of zero")
	}

	if len(password) == 0 {
		ErrorLogger.Println("password has a length of zero")
		return fmt.Errorf("password has a alength of zero")
	}

	credentials = append(credentials, userName)
	credentials = append(credentials, password)

	loginLinux(credentials, expName, vmName, linuxLoginPrompt, linuxAfterLoginPrompt, timeout, threshold)

	return nil

}

func parseLinuxTermCommand(expName string, parameters []interface{}) error {

	var (
		vmName   string
		commands []string
	)

	for _, param := range parameters {

		if _, ok := param.(map[string]interface{}); !ok {
			ErrorLogger.Println("Unexpected Type for parameter")
			continue
		}

		for k, v := range param.(map[string]interface{}) {

			switch k {

			case "vm-name":
				vmName, _ = v.(string)
			case "commands":
				for _, command := range v.([]interface{}) {
					if _, ok := command.(string); !ok {
						continue
					}

					commands = append(commands, command.(string))
				}

			}

		}
	}

	InfoLogger.Printf("Commands:%v", commands)

	if len(vmName) == 0 {
		ErrorLogger.Println("vm name has a length of zero")
		return fmt.Errorf("vm name has a length of zero")
	}

	if len(commands) == 0 {
		ErrorLogger.Println("there are no commands to send")
		return fmt.Errorf("there are no commands to send")
	}

	sendLinuxTermCommands(expName, vmName, commands)

	return nil
}

func processConfig(expName, filePath string) error {

	playbook, err := readConfig(filePath)

	if err != nil {
		return fmt.Errorf("Reading config file %s:%v\n", filePath, err)
	}

	for _, script := range playbook.Scripts {
		for _, action := range script.Actions {

			if _, ok := action.Parameters.([]interface{}); !ok {
				ErrorLogger.Println("Unexpected Type for action.Parameters")
				continue
			}

			InfoLogger.Printf("Name:%s", action.Name)

			switch action.Name {

			case "linux-terminal-commands":
				if err := parseLinuxTermCommand(expName, action.Parameters.([]interface{})); err != nil {
					return fmt.Errorf("%v", err)
				}
			case "login-linux":
				if err := parseLoginLinux(expName, action.Parameters.([]interface{})); err != nil {
					return fmt.Errorf("%v", err)
				}
			case "send-special-keys":
				if err := parseSendSpecialKeys(expName, action.Parameters.([]interface{})); err != nil {
					return fmt.Errorf("%v", err)
				}
			case "sleep":
				if err := parseSleep(expName, action.Parameters.([]interface{})); err != nil {
					return fmt.Errorf("%v", err)
				}
			case "gui-terminal-commands":
				if err := parseUiTermCommand(expName, action.Parameters.([]interface{})); err != nil {
					return fmt.Errorf("%v", err)
				}
			case "wait-until":
				if err := parseWaitUntil(expName, action.Parameters.([]interface{})); err != nil {
					return fmt.Errorf("%v", err)
				}
			default:
				InfoLogger.Printf("%v is not a supported action", action.Name)

			}

		}
	}

	return nil
}
