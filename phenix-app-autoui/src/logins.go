package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func loginKali(expName, vmName, loginRef, credentials string) error {

	var (
		centerX int
		centerY int
		err     error
	)

	loginRef = filepath.Join(refImageDirectory, loginRef)

	fmt.Println("Checking for login screen")

	// Wait for login screen to appear
	centerX, centerY, err = waitUntil(expName, vmName, loginRef, "1m30s")

	if err != nil {
		return fmt.Errorf("loginKali:%v\n", err)
	}

	// Send credentials
	tmp := strings.Split(credentials, " ")
	username := tmp[0]
	password := tmp[1]

	fmt.Printf("User:%s Pass:%s\n", username, password)

	sendCharacters(expName, vmName, fmt.Sprintf("%s", username))
	sendSpecialKeys(expName, vmName, "Tab")
	time.Sleep(1 * time.Second)
	sendCharacters(expName, vmName, fmt.Sprintf("%s", password))
	time.Sleep(1 * time.Second)

	// Click the login button
	if err := clickImage(centerX, centerY, 1, expName, vmName); err != nil {
		return fmt.Errorf("loginKali:%v\n", err)
	}

	return nil
}
