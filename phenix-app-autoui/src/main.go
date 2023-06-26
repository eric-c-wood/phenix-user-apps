package main

import (
//"fmt"
//"path/filepath"
)

func main() {

	/*
		if err := sendCommand("AutoUi2", "Kali2023a", "kali" filepath.Join(refImageDirectory, "termBtn3.png"),
			filepath.Join(refImageDirectory, "termWindow2.png"), "ls -alht"); err != nil {
			fmt.Printf("Error:%v\n", err)
		}

		return
	*/

	//loginKali("AutoUi2", "WinXP", "termWinXPWindow.png", "kali2023 kali2023")
	/*
		terminalCommand := &termCommand{
			vmName:      "WinXP",
			osName:      "WinXP",
			termWindow:  "termWinXPWindow.png",
			dsktpSymbol: "startXPBtn.png",
			commands:    []string{"dir /w", "help"},
		}
	*/

	terminalCommand := &termCommand{
		vmName:      "Kali2023a",
		osName:      "kali",
		termWindow:  "termKaliWindow.png",
		dsktpSymbol: "Home_Button.png",
		commands:    []string{"ls -alht", "pwd"},
	}

	convertPaths(terminalCommand)
	sendCommands("AutoUi2", terminalCommand)

	/*

		centerX, centerY, err := waitUntil("AutoUi2", "Kali2023a", "/phenix/images/screenshots/Home_Button.png", "1m")

		if err != nil {
			fmt.Printf("%v\n", err)
		}

		if err := doubleClickImage(centerX, centerY, 1, "AutoUi2", "Kali2023a"); err != nil {
			fmt.Printf("Failed to double click:%v\n", err)
		}

		return

		if err := getFileFromHost("192.168.60.2", "/tmp/special_characters.json", "/tmp/test.json"); err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Print("File transfer successful\n")
		}

	*/

}
