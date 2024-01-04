package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"phenix/util/mm/mmcli"
)

var (
	extraDelay           = 500
	defaultThreshold     = float32(0.9)
	screenshotDirectory  = getDirectory("tmpScreenshots")
	playbackDirectory    = "/phenix/images"
	refImageDirectory    = getDirectory("refImages")
	specialCharacters, _ = readJson(filepath.Join(getParentDirectory(), "lookup", "special_characters.json"))
	normalCharactersRe   = regexp.MustCompile(`(?i)[a-z0-9]`)
	noShiftList          = map[string]struct{}{"bracketright": struct{}{}, "bracketleft": struct{}{}, "semicolon": struct{}{}, "apostrophe": struct{}{}, "comma": struct{}{},
		"slash": struct{}{}, "period": struct{}{}, "equal": struct{}{}, "minus": struct{}{}}
)

type termCommand struct {
	vmName      string
	osName      string
	termWindow  string
	termBtn     string
	dsktpSymbol string
	timeout     string
	threshold   string
	commands    []string
}

func buildInjectString(vmName, sendText string, options ...int) ([]string, error) {

	if len(specialCharacters) == 0 {
		return nil, fmt.Errorf("Unable to build inject string:%v", sendText)
	}

	var (
		commands []string
		sendChar string
		chr      string
		delay    = 10000000
	)

	if len(options) > 0 {
		delay = options[0]
	}

	for _, s := range sendText {
		chr = string(s)
		sendChar = chr
		if !normalCharactersRe.Match([]byte(chr)) {
			if _, ok := specialCharacters[chr]; ok {
				sendChar = specialCharacters[chr].(string)
			}

			if _, ok := noShiftList[sendChar]; !ok {
				commands = append(commands, fmt.Sprintf("%d:KeyEvent,%s,%s\n", delay, "true", "Shift_L"))
				commands = append(commands, fmt.Sprintf("%d:KeyEvent,%s,%s\n", delay, "true", sendChar))
				commands = append(commands, fmt.Sprintf("%d:KeyEvent,%s,%s\n", delay, "false", "Shift_L"))
				commands = append(commands, fmt.Sprintf("%d:KeyEvent,%s,%s\n", delay, "false", sendChar))

			} else {
				commands = append(commands, fmt.Sprintf("%d:KeyEvent,%s,%s\n", delay, "true", sendChar))
				commands = append(commands, fmt.Sprintf("%d:KeyEvent,%s,%s\n", delay, "false", sendChar))
			}

		} else {
			commands = append(commands, fmt.Sprintf("%d:KeyEvent,%s,%s\n", delay, "true", sendChar))
			commands = append(commands, fmt.Sprintf("%d:KeyEvent,%s,%s\n", delay, "false", sendChar))
		}

	}

	return commands, nil

}

func buildKeysInjectString(vmName, sendKeys string) ([]string, error) {

	keyList := strings.Split(sendKeys, " ")

	var commands []string

	for _, key := range keyList {
		commands = append(commands, fmt.Sprintf("vnc inject %s KeyEvent,true,%s", vmName, key))
	}

	return commands, nil

}

func sendCharacters(expName, vmName, characters string, options ...int) error {

	var delay = 10000000

	if len(options) > 0 {
		delay = options[0]
	}

	injectList, err := buildInjectString(vmName, characters, delay)

	if err != nil {
		return fmt.Errorf("sendCharacters Error:%v\n", err)
	}

	playbackFile := filepath.Join(screenshotDirectory, fmt.Sprintf("%s_%s_keyboard_send.kb", expName, vmName))

	if err := writeFile(playbackFile, injectList); err != nil {
		return fmt.Errorf("clickImage:%v\n", err)
	}

	cmd := mmcli.NewNamespacedCommand(expName)
	cmd.Command = fmt.Sprintf("vnc play %s %s", vmName, playbackFile)

	InfoLogger.Println(cmd.Command)

	if err := mmcli.ErrorResponse(mmcli.Run(cmd)); err != nil {
		ErrorLogger.Printf("%v", err)
		return fmt.Errorf("clickImage error:%v\n", err)
	}

	time.Sleep(time.Duration(len(injectList)*delay) * time.Nanosecond)
	time.Sleep(time.Duration(extraDelay) * time.Millisecond)

	return nil
}

func sendSpecialKeys(expName, vmName, keys string) error {

	injectList, err := buildKeysInjectString(vmName, keys)

	if err != nil {
		return fmt.Errorf("sendSpecialKeys Error:%v", err)
	}

	cmd := mmcli.NewNamespacedCommand(expName)

	for _, injectStr := range injectList {

		InfoLogger.Printf("Command:%s", injectStr)

		cmd.Command = injectStr

		if err := mmcli.ErrorResponse(mmcli.Run(cmd)); err != nil {
			return fmt.Errorf("sendSpecialKeys error:%v", err)
		}

		time.Sleep(200 * time.Millisecond)

	}

	return nil

}

func waitUntil(expName, vmName, srcImg string, options ...string) (int, int, error) {

	screenshotTimer := time.NewTicker(5 * time.Second)

	var timeout = "30m"
	var threshold = defaultThreshold

	if len(options) > 0 {
		if len(options[0]) > 0 {
			_, err := time.ParseDuration(options[0])
			if err == nil {
				timeout = options[0]
			}
		}

	}

	if len(options) > 1 {
		if len(options[1]) > 0 {
			tmp, _ := strconv.ParseFloat(options[1], 32)
			threshold = float32(tmp)
		}
	}

	quitTimer, _ := time.ParseDuration(timeout)
	InfoLogger.Printf("QuitTimer:%v", quitTimer)

	defer screenshotTimer.Stop()

	for {
		select {
		case <-screenshotTimer.C:

			InfoLogger.Println("Taking Screenshot")
			refImg := filepath.Join(screenshotDirectory, fmt.Sprintf("%s_%s_Ref.png", expName, vmName))

			quitTimer -= (5 * time.Second)

			if err := takeScreenshot(expName, vmName, refImg); err != nil {
				return -1, -1, fmt.Errorf("Screenshot failed %v", err)
			}

			centerX, centerY, err := getXY(srcImg, refImg, threshold)

			if err == nil {
				return centerX, centerY, nil
			}

		case <-time.After(quitTimer):
			return -1, -1, fmt.Errorf("Failed to find image:%v in %v\n", srcImg, quitTimer)
		}

		InfoLogger.Printf("Timer is %v", quitTimer)
		quitTimer -= time.Second

	}

	return -1, -1, fmt.Errorf("Failed to find image:%v in %v\n", srcImg, quitTimer)

}

func waitUntilGone(expName, vmName, srcImg string, options ...string) (int, int, error) {

	screenshotTimer := time.NewTicker(5 * time.Second)

	var timeout = "30m"

	if len(options) > 0 {
		timeout = options[0]
	}

	quitTimer, _ := time.ParseDuration(timeout)
	InfoLogger.Printf("QuitTimer:%v", quitTimer)

	defer screenshotTimer.Stop()

	for {
		select {
		case <-screenshotTimer.C:

			fmt.Println("Taking Screenshot")
			refImg := filepath.Join(screenshotDirectory, fmt.Sprintf("%s_%s_Ref.png", expName, vmName))

			quitTimer -= (5 * time.Second)

			if err := takeScreenshot(expName, vmName, refImg); err != nil {
				return -1, -1, fmt.Errorf("Screenshot failed %v", err)
			}

			centerX, centerY, err := getXY(srcImg, refImg)

			if err != nil {
				return centerX, centerY, err
			}

		case <-time.After(quitTimer):
			return -1, -1, fmt.Errorf("Failed to find image:%v in %v\n", srcImg, quitTimer)
		}

		InfoLogger.Printf("Timer is %v", quitTimer)
		quitTimer -= time.Second

	}

	return -1, -1, fmt.Errorf("Failed to find image:%v in %v\n", srcImg, quitTimer)

}

func waitUntilChange(expName, vmName, srcImg string) error {

	screenshotTimer := time.NewTicker(5 * time.Second)
	quitTimer, _ := time.ParseDuration("30m")
	InfoLogger.Printf("QuitTimer:%v", quitTimer)

	defer screenshotTimer.Stop()

	for {
		select {
		case <-screenshotTimer.C:

			InfoLogger.Println("waitUntilChange:Taking Screenshot")
			refImg := filepath.Join(screenshotDirectory, fmt.Sprintf("%s_%s_Ref_UntilChange.png", expName, vmName))

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

		InfoLogger.Printf("Timer is %v", quitTimer)
		quitTimer -= time.Second

	}

	return fmt.Errorf("Failed to identify an image change:%v in %v\n", srcImg, quitTimer)

}

func rightClickImage(expName, vmName, imgName string) error {

	if err := clickImage(2, expName, vmName, imgName); err != nil {
		ErrorLogger.Printf("rightClickImage:%v", err)
		return fmt.Errorf("rightClickImage:%v\n", err)
	}

	return nil

}

func clickImage(btn uint8, expName, vmName, imgName string) error {

	var (
		commands          []string
		clickPlaybackFile string
		delay             = 100000000
	)

	x, y, err := waitUntil(expName, vmName, filepath.Join(refImageDirectory, imgName))

	if err != nil {
		ErrorLogger.Printf("clickImage:%v", err)
		return fmt.Errorf("clickImage:%v\n", err)
	}

	// Build the list of commands to properly emulate a click
	for i := 0; i < 4; i++ {

		if i == 1 {
			commands = append(commands, fmt.Sprintf("%d:PointerEvent,%d,%d,%d\n", delay, btn, x, y))
		} else {
			commands = append(commands, fmt.Sprintf("%d:PointerEvent,%d,%d,%d\n", delay, 0, x, y))
		}

	}

	clickPlaybackFile = filepath.Join(screenshotDirectory, fmt.Sprintf("%s_%s_click.kb", expName, vmName))

	if err := writeFile(clickPlaybackFile, commands); err != nil {
		ErrorLogger.Printf("clickImage:%v", err)
		return fmt.Errorf("clickImage:%v\n", err)
	}

	cmd := mmcli.NewNamespacedCommand(expName)
	cmd.Command = fmt.Sprintf("vnc play %s %s", vmName, clickPlaybackFile)

	//fmt.Println(cmd.Command)

	if err := mmcli.ErrorResponse(mmcli.Run(cmd)); err != nil {
		ErrorLogger.Printf("clickImage:%v", err)
		return fmt.Errorf("clickImage error:%v\n", err)
	}

	time.Sleep(time.Duration(len(commands)*delay) * time.Nanosecond)
	time.Sleep(time.Duration(extraDelay) * time.Millisecond)

	return nil

}

func doubleClickImage(btn uint8, expName, vmName, imgName string) error {

	var (
		commands          []string
		clickPlaybackFile string
		delay             = 100000000
	)

	x, y, err := waitUntil(expName, vmName, filepath.Join(refImageDirectory, imgName))

	if err != nil {
		ErrorLogger.Printf("doubleClickImage:%v", err)
		return fmt.Errorf("doubleClickImage:%v\n", err)
	}

	// Build the list of commands to properly emulate a click
	for i := 0; i < 8; i++ {

		if i == 1 || i == 5 {
			commands = append(commands, fmt.Sprintf("%d:PointerEvent,%d,%d,%d\n", delay, btn, x, y))
		} else {
			commands = append(commands, fmt.Sprintf("%d:PointerEvent,%d,%d,%d\n", delay, 0, x, y))
		}

	}

	clickPlaybackFile = filepath.Join(screenshotDirectory, fmt.Sprintf("%s_%s_doubleclick.kb", expName, vmName))

	if err := writeFile(clickPlaybackFile, commands); err != nil {
		ErrorLogger.Printf("doubleClickImage:%v", err)
		return fmt.Errorf("doubleClickImage:%v\n", err)
	}

	cmd := mmcli.NewNamespacedCommand(expName)
	cmd.Command = fmt.Sprintf("vnc play %s %s", vmName, clickPlaybackFile)

	if err := mmcli.ErrorResponse(mmcli.Run(cmd)); err != nil {
		ErrorLogger.Printf("doubleClickImage error:%v", err)
		return fmt.Errorf("doubleClickImage error:%v\n", err)
	}

	time.Sleep(time.Duration(len(commands)*delay) * time.Nanosecond)
	time.Sleep(time.Duration(extraDelay) * time.Millisecond)

	return nil

}

func moveMouseTo(expName, vmName, imgName string) error {

	if err := clickImage(0, expName, vmName, imgName); err != nil {
		ErrorLogger.Printf("moveMouseTo error:%v", err)
		fmt.Errorf("moveMouseTo error:%v\n", err)
	}

	return nil
}

func sendGUITermCommands(expName string, termCommand *termCommand) error {

	osName := strings.ToLower(termCommand.osName)
	var srcImg string

	// Check to see if the terminal is opened
	if !isTerminalOpen(expName, termCommand) {

		InfoLogger.Println("Terminal Window is not open")
		switch osName {
		case "kali":
			if err := openKaliTerminal(expName, termCommand); err != nil {
				ErrorLogger.Printf("%v", err)
				return fmt.Errorf("%v\n", err)
			}

		case "winxp", "win7":
			if err := openWinTerminal(expName, termCommand); err != nil {
				ErrorLogger.Printf("%v", err)
				return fmt.Errorf("%v\n", err)
			}

		default:
			ErrorLogger.Printf("sendUITermCommands: operating system %s not supported", osName)
			return fmt.Errorf("sendUITermCommands: operating system %s not supported\n", osName)

		}

	}

	for _, command := range termCommand.commands {

		InfoLogger.Printf("Sending Command:%v", command)

		// Send the command followed by the "enter" key
		if err := sendCharacters(expName, termCommand.vmName, command); err != nil {
			ErrorLogger.Printf("%v", err)
			return fmt.Errorf("%v\n", err)
		}

		// Take a reference screenshot
		srcImg = filepath.Join(screenshotDirectory, fmt.Sprintf("%s_%s_Src_UntilChange.png", expName, termCommand.vmName))

		InfoLogger.Println(srcImg)

		if err := takeScreenshot(expName, termCommand.vmName, srcImg); err != nil {
			return fmt.Errorf("Screenshot failed %v", err)
		}

		if err := sendSpecialKeys(expName, termCommand.vmName, "Return"); err != nil {
			ErrorLogger.Printf("%v", err)
			return fmt.Errorf("%v\n", err)
		}

		// Wait for the command to complete
		if err := waitUntilChange(expName, termCommand.vmName, srcImg); err != nil {
			ErrorLogger.Printf("%v", err)
			return fmt.Errorf("%v\n", err)
		}

	}

	if err := closeTerminal(expName, termCommand); err != nil {
		ErrorLogger.Printf("%v", err)
		return fmt.Errorf("%v\n", err)
	}

	return nil

}

func sendLinuxTermCommands(expName, vmName string, commands []string) error {

	var srcImg string

	for _, command := range commands {

		InfoLogger.Printf("Sending Command:%v", command)

		// Send the command followed by the "enter" key
		if err := sendCharacters(expName, vmName, command); err != nil {
			ErrorLogger.Printf("%v", err)
			return fmt.Errorf("%v\n", err)
		}

		// Take a reference screenshot
		srcImg = filepath.Join(screenshotDirectory, fmt.Sprintf("%s_%sSrc_UntilChange.png", expName, vmName))

		InfoLogger.Println(srcImg)

		if err := takeScreenshot(expName, vmName, srcImg); err != nil {
			return fmt.Errorf("Screenshot failed %v", err)
		}

		if err := sendSpecialKeys(expName, vmName, "Return"); err != nil {
			ErrorLogger.Printf("%v", err)
			return fmt.Errorf("%v\n", err)
		}

		// Wait for the command to complete
		if err := waitUntilChange(expName, vmName, srcImg); err != nil {
			ErrorLogger.Printf("%v", err)
			return fmt.Errorf("%v\n", err)
		}

	}

	return nil

}

func sendKBShortcut(expName string, vmName, shortcut string) error {

	var (
		commands     []string
		playbackFile string
		delay        = 100000000
	)

	keys := strings.Split(shortcut, "+")

	for i := 0; i < 2; i++ {
		for _, key := range keys {
			if i == 1 {
				commands = append(commands, fmt.Sprintf("%d:KeyEvent,%s,%s\n", delay, "false", key))
			} else {
				commands = append(commands, fmt.Sprintf("%d:KeyEvent,%s,%s\n", delay, "true", key))
			}
		}
	}

	playbackFile = filepath.Join(screenshotDirectory, fmt.Sprintf("%s_%s_shortcut.kb", expName, vmName))

	if err := writeFile(playbackFile, commands); err != nil {
		ErrorLogger.Printf("sendKBShortcut:%v", err)
		return fmt.Errorf("sendKBShortcut:%v\n", err)
	}

	cmd := mmcli.NewNamespacedCommand(expName)
	cmd.Command = fmt.Sprintf("vnc play %s %s", vmName, playbackFile)

	//fmt.Println(cmd.Command)

	if err := mmcli.ErrorResponse(mmcli.Run(cmd)); err != nil {
		ErrorLogger.Printf("sendKBShortcut:%v", err)
		return fmt.Errorf("sendKBShortcut error:%v\n", err)
	}

	time.Sleep(time.Duration(len(commands)*delay) * time.Nanosecond)
	time.Sleep(time.Duration(extraDelay) * time.Millisecond)

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
