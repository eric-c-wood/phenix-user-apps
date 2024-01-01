package main

import (
	"fmt"
	"strings"
	//"time"
)

func openKaliTerminal(expName string, termCommand *termCommand) error {

	InfoLogger.Println("Trying to open terminal window")

	if err := sendKBShortcut(expName, termCommand.vmName, "Control_L+Alt_L+T"); err != nil {
		ErrorLogger.Println("openKaliTerminal:%v", err)
		return fmt.Errorf("openKaliTerminal:%v\n", err)
	}

	// Wait for the terminal to open
	_, _, err := waitUntil(expName, termCommand.vmName, termCommand.termWindow, "10m")

	if err != nil {
		ErrorLogger.Printf("openKaliTerminal:%v", err)
		return fmt.Errorf("openKaliTerminal:%v\n", err)
	}

	//Maximize the terminal window.  For some reason, maximizing the terminal
	//window causes taking screenshots to fail
	/*
		if err := sendKBShortcut(expName, termCommand.vmName, "Alt_L+F11"); err != nil {
			return fmt.Errorf("openKaliTerminal:%v\n", err)
		}

	*/

	return nil
}

func openWinTerminal(expName string, termCommand *termCommand) error {

	// Wait for the desktop to open

	if err := waitForDesktop(expName, termCommand); err != nil {
		ErrorLogger.Printf("openWinTerminal:%v", err)
		return fmt.Errorf("openWinTerminal:%v\n", err)
	}

	InfoLogger.Println("Trying to open terminal window")

	// Open the "Run" menu
	if err := sendKBShortcut(expName, termCommand.vmName, "Super_L+R"); err != nil {
		ErrorLogger.Printf("openWinTerminal:%v", err)
		return fmt.Errorf("openWinTerminal:%v\n", err)
	}

	// Send the "R" key to open the "Run" menu
	if err := sendCharacters(expName, termCommand.vmName, "cmd"); err != nil {
		ErrorLogger.Printf("openWinTerminal:%v", err)
		return fmt.Errorf("openWinTerminal:%v\n", err)
	}

	// Send "enter" or "Return" key to finish opening the terminal window
	if err := sendSpecialKeys(expName, termCommand.vmName, "Return"); err != nil {
		ErrorLogger.Printf("openWinTerminal:%v", err)
		return fmt.Errorf("openWinterminal:%v\n", err)
	}

	// Wait for the terminal to open
	if _, _, err := waitUntil(expName, termCommand.vmName, termCommand.termWindow, "10m"); err != nil {
		ErrorLogger.Printf("openWinTerminal:%v", err)
		return fmt.Errorf("openWinTerminal:%v\n", err)
	}

	// Maximize the terminal window using "ALT_L + ENTER"
	if err := sendKBShortcut(expName, termCommand.vmName, "Alt_L+Return"); err != nil {
		ErrorLogger.Printf("openWinTerminal:%v", err)
		return fmt.Errorf("openWinTerminal:%v\n", err)
	}

	return nil
}

func closeUITerminal(expName string, termCommand *termCommand) error {

	// Close the terminal window by typing "exit"
	if err := sendCharacters(expName, termCommand.vmName, "exit"); err != nil {
		ErrorLogger.Printf("closeTerminal:%v", err)
		return fmt.Errorf("closeerminal:%v\n", err)
	}

	if err := sendSpecialKeys(expName, termCommand.vmName, "Return"); err != nil {
		ErrorLogger.Printf("closeTerminal:%v", err)
		return fmt.Errorf("closeTerminal:%v\n", err)
	}

	return nil
}

func isTerminalOpen(expName string, termCommand *termCommand) bool {

	_, _, err := waitUntil(expName, termCommand.vmName, termCommand.termWindow, "10s")

	if err != nil {
		return false
	}

	return false
}

func waitForDesktop(expName string, termCommand *termCommand) error {

	_, _, err := waitUntil(expName, termCommand.vmName, termCommand.dsktpSymbol, "30m")

	if err != nil {
		ErrorLogger.Printf("waitForDesktop:%v", err)
		return fmt.Errorf("waitForDesktop:%v\n", err)
	}

	return nil

}

func closeTerminal(expName string, termCommand *termCommand) error {

	osName := strings.ToLower(termCommand.osName)

	// Call the appropriate close terminal function
	switch osName {
	case "kali":
		if err := closeUITerminal(expName, termCommand); err != nil {
			ErrorLogger.Printf("closeTerminal:%v", err)
			return fmt.Errorf("closeTerminal:%v\n", err)
		}
	case "winxp", "win7":
		if err := closeUITerminal(expName, termCommand); err != nil {
			ErrorLogger.Printf("closeTerminal:%v", err)
			return fmt.Errorf("closeTerminal:%v\n", err)
		}

	default:
		ErrorLogger.Printf("closeTerminal: operating system %s not supported", osName)
		return fmt.Errorf("closeTerminal: operating system %s not supported", osName)

	}

	return nil

}
