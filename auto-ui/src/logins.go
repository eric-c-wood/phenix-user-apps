package main

import (
	"fmt"
	"path/filepath"
	"time"
)

func loginKali(expName, vmName, loginRef string, credentials []string) error {

	InfoLogger.Println("Checking for login screen")

	// Wait for login screen to appear
	if _, _, err := waitUntil(expName, vmName, loginRef); err != nil {
		return fmt.Errorf("loginKali:%v\n", err)
	}

	// Send credentials
	username := credentials[0]
	password := credentials[1]

	InfoLogger.Printf("User:%s Pass:%s", username, password)

	sendCharacters(expName, vmName, fmt.Sprintf("%s", username))
	sendSpecialKeys(expName, vmName, "Tab")
	time.Sleep(1 * time.Second)
	sendCharacters(expName, vmName, fmt.Sprintf("%s", password))
	time.Sleep(1 * time.Second)

	// Click the login button
	if err := clickImage(1, expName, vmName, loginRef); err != nil {
		ErrorLogger.Printf("loginKali:%v", err)
		return fmt.Errorf("loginKali:%v\n", err)
	}

	return nil
}

func loginLinux(credentials []string, options ...string) error {

	var (
		srcImg    string
		timeout   = "30m"
		threshold = "0.9"
	)

	if len(options) < 4 {
		ErrorLogger.Printf("Expected 4 parameters, received %d", len(options))
		return fmt.Errorf("Expected 4 parameters, received %d", len(options))
	}

	expName := options[0]
	vmName := options[1]
	linuxLoginPrompt := options[2]
	linuxAfterLoginPrompt := options[3]

	if len(options) > 4 {
		if len(options[4]) > 0 {
			InfoLogger.Printf("Login Timeout:%v", options[4])
			timeout = options[4]
		}

	}

	if len(options) > 5 {
		if len(options[5]) > 0 {
			InfoLogger.Printf("Login Threshold:%v", options[5])
			threshold = options[5]
		}

	}

	InfoLogger.Println("Checking for Linux login screen")

	// Wait for login screen to appear
	if _, _, err := waitUntil(expName, vmName, linuxLoginPrompt, timeout, threshold); err != nil {
		return fmt.Errorf("loginLinux:%v\n", err)
	}

	// Send credentials
	username := credentials[0]
	password := credentials[1]

	InfoLogger.Printf("User:%s Pass:%s", username, password)

	sendCharacters(expName, vmName, fmt.Sprintf("%s", username))

	// Take a reference screenshot
	srcImg = filepath.Join(screenshotDirectory, fmt.Sprintf("%s_%sSrc_UntilChange.png", expName, vmName))

	InfoLogger.Println(srcImg)

	if err := takeScreenshot(expName, vmName, srcImg); err != nil {
		return fmt.Errorf("Screenshot failed %v", err)
	}

	sendSpecialKeys(expName, vmName, "Return")

	// Wait for the command to complete
	if err := waitUntilChange(expName, vmName, srcImg); err != nil {
		ErrorLogger.Printf("loginLinux:%v", err)
		return fmt.Errorf("loginLinux:%v\n", err)
	}

	sendCharacters(expName, vmName, fmt.Sprintf("%s", password))

	// Take a reference screenshot
	srcImg = filepath.Join(screenshotDirectory, fmt.Sprintf("%s_%sSrc_UntilChange.png", expName, vmName))

	InfoLogger.Println(srcImg)

	if err := takeScreenshot(expName, vmName, srcImg); err != nil {
		return fmt.Errorf("Screenshot failed %v", err)
	}

	sendSpecialKeys(expName, vmName, "Return")

	// Block until the command completes
	if err := waitUntilChange(expName, vmName, srcImg); err != nil {
		ErrorLogger.Printf("loginLinux:%v", err)
		return fmt.Errorf("loginLinux:%v\n", err)
	}

	// Wait for the command prompt
	if _, _, err := waitUntil(expName, vmName, linuxAfterLoginPrompt, timeout, threshold); err != nil {
		ErrorLogger.Printf("loginLinux:%v", err)
		return fmt.Errorf("loginLinux:%v\n", err)
	}

	return nil
}
