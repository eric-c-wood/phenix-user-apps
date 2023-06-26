package main

import (
	"fmt"

	"path/filepath"
	"regexp"
	"strings"
	"time"

	"phenix/util/mm/mmcli"
)

var (
	screenshotDirectory  = getDirectory("tmpScreenshots")
	refImageDirectory    = getDirectory("refImages")
	specialCharacters, _ = readJson(filepath.Join(getParentDirectory(), "lookup", "special_characters.json"))
	normalCharactersRe   = regexp.MustCompile(`(?i)[a-z0-9]`)
)

type termCommand struct {
	vmName      string
	osName      string
	termWindow  string
	termBtn     string
	dsktpSymbol string
	commands    []string
}

func buildInjectString(vmName, sendText string) ([]string, error) {

	if len(specialCharacters) == 0 {
		return nil, fmt.Errorf("Unable to build inject string:%v", sendText)
	}

	var (
		commands []string
		sendChar string
		chr      string
	)

	for _, s := range sendText {
		chr = string(s)
		sendChar = chr
		if !normalCharactersRe.Match([]byte(chr)) {
			if _, ok := specialCharacters[chr]; ok {
				sendChar = specialCharacters[chr].(string)
			}

		}

		commands = append(commands, fmt.Sprintf("vnc inject %s KeyEvent,true,%s", vmName, sendChar))
		//commands = append(commands, fmt.Sprintf("vnc inject %s KeyEvent,false,%s", vmName, sendChar))
	}

	return commands, nil

}

func buildKeysInjectString(vmName, sendKeys string) ([]string, error) {

	keyList := strings.Split(sendKeys, " ")

	var commands []string

	for _, key := range keyList {
		commands = append(commands, fmt.Sprintf("vnc inject %s KeyEvent,true,%s", vmName, key))
		//commands = append(commands, fmt.Sprintf("vnc inject %s KeyEvent,false,%s", vmName, key))
	}

	return commands, nil

}

func sendCharacters(expName, vmName, characters string) error {

	injectList, err := buildInjectString(vmName, characters)

	if err != nil {
		return fmt.Errorf("sendCharacters Error:%v", err)
	}

	cmd := mmcli.NewNamespacedCommand(expName)

	for _, injectStr := range injectList {

		cmd.Command = injectStr

		if err := mmcli.ErrorResponse(mmcli.Run(cmd)); err != nil {
			return fmt.Errorf("sendCharacters error:%v", err)
		}

		time.Sleep(250 * time.Millisecond)

	}

	return nil
}

func sendSpecialKeys(expName, vmName, keys string) error {

	injectList, err := buildKeysInjectString(vmName, keys)

	if err != nil {
		return fmt.Errorf("sendSpecialKeys Error:%v", err)
	}

	cmd := mmcli.NewNamespacedCommand(expName)

	for _, injectStr := range injectList {

		fmt.Printf("Command:%s\n", injectStr)

		cmd.Command = injectStr

		if err := mmcli.ErrorResponse(mmcli.Run(cmd)); err != nil {
			return fmt.Errorf("sendSpecialKeys error:%v", err)
		}

		time.Sleep(200 * time.Millisecond)

	}

	return nil

}

func waitUntil(expName, vmName, srcImg, timeout string) (int, int, error) {

	screenshotTimer := time.NewTicker(5 * time.Second)
	quitTimer, _ := time.ParseDuration(timeout)
	fmt.Printf("QuitTimer:%v\n", quitTimer)

	defer screenshotTimer.Stop()

	for {
		select {
		case <-screenshotTimer.C:

			fmt.Println("Taking Screenshot")
			refImg := filepath.Join(screenshotDirectory, fmt.Sprintf("%sRef.png", vmName))

			quitTimer -= (5 * time.Second)

			if err := takeScreenshot(expName, vmName, refImg); err != nil {
				return -1, -1, fmt.Errorf("Screenshot failed %v", err)
			}

			centerX, centerY, err := getXY(srcImg, refImg)

			if err == nil {
				return centerX, centerY, nil
			}

		case <-time.After(quitTimer):
			return -1, -1, fmt.Errorf("Failed to find image:%v in %v\n", srcImg, quitTimer)
		}

		fmt.Printf("Timer is %v\n", quitTimer)
		quitTimer -= time.Second

	}

	return -1, -1, fmt.Errorf("Failed to find image:%v in %v\n", srcImg, quitTimer)

}

func waitUntilChange(expName, vmName string) error {

	// Take an initial screenshot
	srcImg := filepath.Join(screenshotDirectory, fmt.Sprintf("%sSrc_UntilChange.png", vmName))

	fmt.Println(srcImg)

	if err := takeScreenshot(expName, vmName, srcImg); err != nil {
		return fmt.Errorf("Screenshot failed %v", err)
	}

	screenshotTimer := time.NewTicker(5 * time.Second)
	quitTimer, _ := time.ParseDuration("5m")
	fmt.Printf("QuitTimer:%v\n", quitTimer)

	defer screenshotTimer.Stop()

	for {
		select {
		case <-screenshotTimer.C:

			fmt.Println("waitUntilChange:Taking Screenshot")
			refImg := filepath.Join(screenshotDirectory, fmt.Sprintf("%sRef_UntilChange.png", vmName))

			quitTimer -= (5 * time.Second)

			if err := takeScreenshot(expName, vmName, refImg); err != nil {
				return fmt.Errorf("Screenshot failed %v\n", err)
			}

			if imageChanged(srcImg, refImg) {
				return nil
			}

		case <-time.After(quitTimer):
			return fmt.Errorf("Failed to identify an image change:%v in %v\n", srcImg, quitTimer)
		}

		fmt.Printf("Timer is %v\n", quitTimer)
		quitTimer -= time.Second

	}

	return fmt.Errorf("Failed to identify an image change:%v in %v\n", srcImg, quitTimer)

}

func clickImage(x, y int, btn uint8, expName, vmName string) error {

	var (
		commands          []string
		clickPlaybackFile string
	)

	// Build the list of commands to properly emulate a click
	for i := 0; i < 4; i++ {

		if i == 1 {
			commands = append(commands, fmt.Sprintf("%d:PointerEvent,%d,%d,%d\n", 100000000, btn, x, y))
		} else {
			commands = append(commands, fmt.Sprintf("%d:PointerEvent,%d,%d,%d\n", 100000000, 0, x, y))
		}

	}

	clickPlaybackFile = filepath.Join(screenshotDirectory, fmt.Sprintf("%s_click.kb", vmName))

	if err := writeFile(clickPlaybackFile, commands); err != nil {
		return fmt.Errorf("clickImage:%v\n", err)
	}

	cmd := mmcli.NewNamespacedCommand(expName)
	cmd.Command = fmt.Sprintf("vnc play %s %s", vmName, clickPlaybackFile)

	//fmt.Println(cmd.Command)

	if err := mmcli.ErrorResponse(mmcli.Run(cmd)); err != nil {
		return fmt.Errorf("clickImage error:%v", err)
	}

	return nil

}

func doubleClickImage(x, y int, btn uint8, expName, vmName string) error {

	var (
		commands          []string
		clickPlaybackFile string
	)

	// Build the list of commands to properly emulate a click
	for i := 0; i < 8; i++ {

		if i == 1 || i == 5 {
			commands = append(commands, fmt.Sprintf("%d:PointerEvent,%d,%d,%d\n", 100000000, btn, x, y))
		} else {
			commands = append(commands, fmt.Sprintf("%d:PointerEvent,%d,%d,%d\n", 100000000, 0, x, y))
		}

	}

	clickPlaybackFile = filepath.Join(screenshotDirectory, fmt.Sprintf("%s_doubleclick.kb", vmName))

	if err := writeFile(clickPlaybackFile, commands); err != nil {
		return fmt.Errorf("clickImage:%v\n", err)
	}

	cmd := mmcli.NewNamespacedCommand(expName)
	cmd.Command = fmt.Sprintf("vnc play %s %s", vmName, clickPlaybackFile)

	//fmt.Println(cmd.Command)

	if err := mmcli.ErrorResponse(mmcli.Run(cmd)); err != nil {
		return fmt.Errorf("clickImage error:%v", err)
	}

	return nil

}

func moveMouseTo(x, y int, expName, vmName string) error {

	if err := clickImage(x, y, 0, expName, vmName); err != nil {
		fmt.Printf("moveMouseTo error:%v", err)
	}

	return nil
}

func sendCommands(expName string, termCommand *termCommand) error {

	osName := strings.ToLower(termCommand.osName)

	// Check to see if the terminal is opened
	if !isTerminalOpen(expName, termCommand) {

		fmt.Println("Terminal Window is not open")
		switch osName {
		case "kali":
			if err := openKaliTerminal(expName, termCommand); err != nil {
				return fmt.Errorf("sendCommand:%v\n", err)
			}

		case "winxp":
			if err := openWinXPTerminal(expName, termCommand); err != nil {
				return fmt.Errorf("sendCommand:%v\n", err)
			}

		default:
			return fmt.Errorf("sendCommand: operating system %s not supported", osName)

		}

	}

	for _, command := range termCommand.commands {

		// Send the command followed by the "enter" key
		if err := sendCharacters(expName, termCommand.vmName, command); err != nil {
			return fmt.Errorf("sendCommand:%v\n", err)
		}

		time.Sleep(1 * time.Second)

		if err := sendSpecialKeys(expName, termCommand.vmName, "Return"); err != nil {
			return fmt.Errorf("sendCommand:%v\n", err)
		}

		// Wait for the command to complete
		if err := waitUntilChange(expName, termCommand.vmName); err != nil {
			return fmt.Errorf("sendCommand:%v\n", err)
		}

		time.Sleep(2 * time.Second)
	}

	if err := closeTerminal(expName, termCommand); err != nil {
		return fmt.Errorf("sendCommand:%v\n", err)
	}

	return nil

}

func sendKBShortcut(expName string, vmName, shortcut string) error {

	var (
		commands     []string
		playbackFile string
	)

	keys := strings.Split(shortcut, "+")

	for i := 0; i < 2; i++ {
		for _, key := range keys {
			if i == 1 {
				commands = append(commands, fmt.Sprintf("%d:KeyEvent,%s,%s\n", 100000000, "false", key))
			} else {
				commands = append(commands, fmt.Sprintf("%d:KeyEvent,%s,%s\n", 100000000, "true", key))
			}
		}
	}

	playbackFile = filepath.Join(screenshotDirectory, fmt.Sprintf("%s_shortcut.kb", vmName))

	if err := writeFile(playbackFile, commands); err != nil {
		return fmt.Errorf("sendKBShortcut:%v\n", err)
	}

	cmd := mmcli.NewNamespacedCommand(expName)
	cmd.Command = fmt.Sprintf("vnc play %s %s", vmName, playbackFile)

	//fmt.Println(cmd.Command)

	if err := mmcli.ErrorResponse(mmcli.Run(cmd)); err != nil {
		return fmt.Errorf("sendKBShortcut error:%v", err)
	}

	return nil

}

func convertPaths(termCommand *termCommand) {

	if len(termCommand.termBtn) > 0 {
		termCommand.termBtn = filepath.Join(refImageDirectory, termCommand.termBtn)
	}

	if len(termCommand.termWindow) > 0 {
		termCommand.termWindow = filepath.Join(refImageDirectory, termCommand.termWindow)
	}

	if len(termCommand.dsktpSymbol) > 0 {
		termCommand.dsktpSymbol = filepath.Join(refImageDirectory, termCommand.dsktpSymbol)
	}

}
