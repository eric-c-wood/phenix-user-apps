package main

import (
	"fmt"
	"strings"
	"time"
)

func openKaliTerminal(expName string, termCommand *termCommand) error {

	fmt.Println("Trying to open terminal window")

	if err := sendKBShortcut(expName, termCommand.vmName, "Control_L+Alt_L+T"); err != nil {
		return fmt.Errorf("openKaliTerminal:%v\n", err)
	}

	// Wait for the terminal to open
	_, _, err := waitUntil(expName, termCommand.vmName, termCommand.termWindow, "3m")

	if err != nil {
		return fmt.Errorf("openKaliTerminal:%v\n", err)
	}

	time.Sleep(2 * time.Second)

	//Maximize the terminal window.  For some reason, maximizing the terminal
	//window causes taking screenshots to fail
	/*
		if err := sendKBShortcut(expName, termCommand.vmName, "Alt_L+F11"); err != nil {
			return fmt.Errorf("openKaliTerminal:%v\n", err)
		}

	*/

	return nil
}

func openWinXPTerminal(expName string, termCommand *termCommand) error {

	// Wait for the desktop to open

	if err := waitForDesktop(expName, termCommand); err != nil {
		return fmt.Errorf("openWinXPTerminal:%v\n", err)
	}

	fmt.Println("Trying to open terminal window")

	// Use the "Windows Key" to click on the start button
	if err := sendSpecialKeys(expName, termCommand.vmName, "Super_L"); err != nil {
		return fmt.Errorf("openWinXPTerminal:%v\n", err)
	}

	// Send the "R" key to open the "Run" menu
	if err := sendCharacters(expName, termCommand.vmName, "C"); err != nil {
		return fmt.Errorf("openWinXPTerminal:%v\n", err)
	}

	// Send "enter" or "Return" key to finish opening the terminal window
	if err := sendSpecialKeys(expName, termCommand.vmName, "Return"); err != nil {
		return fmt.Errorf("openWinXPterminal:%v\n", err)
	}

	// Wait for the terminal to open
	if _, _, err := waitUntil(expName, termCommand.vmName, termCommand.termWindow, "1m30s"); err != nil {
		return fmt.Errorf("openWinXPTerminal:%v\n", err)
	}

	time.Sleep(2 * time.Second)

	// Maximize the terminal window using "ALT_L + ENTER"
	if err := sendKBShortcut(expName, termCommand.vmName, "Alt_L+Return"); err != nil {
		return fmt.Errorf("openWinXPTerminal:%v\n", err)
	}

	return nil
}

func closeWinXPTerminal(expName string, termCommand *termCommand) error {

	// Close the terminal window by typing "exit"
	if err := sendCharacters(expName, termCommand.vmName, "exit"); err != nil {
		return fmt.Errorf("sendCommand:%v\n", err)
	}

	if err := sendSpecialKeys(expName, termCommand.vmName, "Return"); err != nil {
		return fmt.Errorf("sendCommand:%v\n", err)
	}

	return nil
}

func closeKaliTerminal(expName string, termCommand *termCommand) error {

	// Close the terminal window by typing "exit"
	if err := sendCharacters(expName, termCommand.vmName, "exit"); err != nil {
		return fmt.Errorf("closeKaliTerminal:%v\n", err)
	}

	if err := sendSpecialKeys(expName, termCommand.vmName, "Return"); err != nil {
		return fmt.Errorf("closeKaliTerminal:%v\n", err)
	}

	return nil
}

func isTerminalOpen(expName string, termCommand *termCommand) bool {

	_, _, err := waitUntil(expName, termCommand.vmName, termCommand.termWindow, "1m")

	if err != nil {
		return false
	}

	return false
}

func waitForDesktop(expName string, termCommand *termCommand) error {

	_, _, err := waitUntil(expName, termCommand.vmName, termCommand.dsktpSymbol, "1m30s")

	if err != nil {
		return fmt.Errorf("waitForDesktop:%v", err)
	}

	return nil

}

func closeTerminal(expName string, termCommand *termCommand) error {

	osName := strings.ToLower(termCommand.osName)

	// Call the appropriate close terminal function
	switch osName {
	case "kali":
		if err := closeKaliTerminal(expName, termCommand); err != nil {
			return fmt.Errorf("closeTerminal:%v\n", err)
		}
	case "winxp":
		if err := closeWinXPTerminal(expName, termCommand); err != nil {
			return fmt.Errorf("closeTerminal:%v\n", err)
		}

	default:
		return fmt.Errorf("closeTerminal: operating system %s not supported", osName)

	}

	return nil

}
