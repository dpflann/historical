package main

import (
	"fmt"
	//"io/ioutil"
	"bufio"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	StartState = iota
	ParseHistoryState
	DisplayHistoryState
	FinishedState
)

type (
	// Command struct will contain the command number and the command string.
	Command struct {
		Number    int    // The number indicating the place in the history that this command occurred
		CmdString string // The command executed
		Selected  bool   // True if this command is selected for the output script
		Order     int    // Indicates the order that this command will occur in the script, this allows reordering of commands
	}

	Commands []Command
)

var CommandRegex = regexp.MustCompile(`^\s+(\d+) (.*)$`)
var pageLengths = 50
var ColorBlue = "\033[94m"
var ColorGreen = "\033[92m"
var ColorWarning = "\033[93m"
var ColorFail = "\033[91m"
var ColorEnd = "\033[0m"
var UnselectedCmdStrFmt = "%d.) %s\n"
var SelectedCmdStrFmt = ColorGreen + "%d.) %s" + ColorEnd + "\n"
var bashStart = "#!/bin/bash\n"

func ParseHistory(history string) (*Commands, error) {
	var commands Commands
	splits := strings.Split(history, "\n")
	for _, v := range splits {
		c := CommandRegex.FindStringSubmatch(v)
		if len(c) > 1 {
			number, err := strconv.Atoi(c[1])
			if err != nil {
				fmt.Println("Unable to parse to int", c[1])
				return nil, err
			}
			cmd := Command{
				Number:    number,
				CmdString: c[2],
			}
			commands = append(commands, cmd)
		}
	}
	return &commands, nil
}

func DisplayHistoryPage(cmds *Commands, start, stop int) int {
	for ; start < stop; start++ {
		cmd := (*cmds)[start]
		fmt.Printf(UnselectedCmdStrFmt, cmd.Number, cmd.CmdString)
	}
	return stop
}

func ParseCommandSelection() {}
func WriteScript()           {}
func CreateScript()          {}

func main() {
	// parse a flag for silent execution, skip this step
	fmt.Println("Welcome to historical, where you can manipulate the past to make your future more productive!")
	fmt.Println("Shall we proceed?")
	state := ParseHistoryState
	scanner := bufio.NewScanner(os.Stdin)
	_ = scanner.Scan()
	if scanner.Text() == "n" {
		state = FinishedState
	}
	var commands *Commands
	for state != FinishedState {
		switch state {
		case ParseHistoryState:
			out, err := exec.Command("bash", "-i", "-c", "history -r; history").Output()
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
			commands, err = ParseHistory(string(out))
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
			fmt.Printf("Parsed %d commands from history\n", len(*commands))
			state = DisplayHistoryState
		case DisplayHistoryState:
			DisplayHistoryPage(commands, 0, 25)
			state = FinishedState
		}
	}
}
