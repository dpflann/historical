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
	QueryUserState
	MainMenuState
	SelectMenuState
	GenerateMenuState
	IncrementCursorState
	DecrementCursorState
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

/*
 Here are the basics:
 1. A main display: listing partial content of the script to generate or nothing / placeholder and then the list of options from Here

 e.g.

    [  No commands selected yet ]

*** Commands ***
    1: preview    2: generate            3: select commands
    4: restart    5: edit selections     6: quit

2. select menu

1.) command 1
2.) command 2
3.) command 3
Showing: 3/3 commands: <-previous, ->next
Select>>3,1-2

1.) command 1
2.) command 2
3.) command 3
Showing: 3/3 commands: <-previous, ->next
Select>>3,1-2

3. selected menu - highlighted what was selected

1.) command 1
2.) command 2
3.) command 3
Showing: 3/3 commands: <-previous, ->next, main menu
Select>>

4. main menu after selections

    [ 3 commands selected ]

*** Commands ***
    1: preview    2: generate            3: select commands
    4: restart    5: edit selections     6: quit

5. generate menu
    [ 3 commands selected ]

Name>>myNewScript

After inputing a name, the file is saved, and the user is taken back to the main menu with a status

6. main menu after generation

Last Generated during session: myNewscript

    [ No commands selected yet ]

*** Commands ***
    1: preview    2: generate            3: select commands
    4: restart    5: edit selections     6: quit
*/

func DisplayMainMenu(selections *[]Command) {
	numberOfCurrentSelections := len(*selections)
	if numberOfCurrentSelections == 0 {
		fmt.Println("[ No commands selected ]")
	} else {
		fmt.Printf("[ %d commands selected ]", numberOfCurrentSelections)
	}

	commands := "*** Commands ***\n\t1: preview\t2: generate\t3: select commands\n\t4: restart\t5: edit script\t6: quit"
	fmt.Println(commands)
}

func DisplaySelectMenu(s *bufio.Scanner, cursor *int, commands *Commands, selections *[]Command) {
	options := "Commands:\t1: previous\t2: next\t3: main menu\t4: quit\n"
	prompt := "Select>>"
	HistoryPage(commands, *cursor, pageLengths)
	fmt.Printf(options)
	fmt.Printf(prompt)
	_ = s.Scan()
}

func ParseSelectQuery(s *bufio.Scanner, commands *Commands, selections *[]Command) int {
	input := s.Text()
	switch input {
	case "p":
		return DecrementCursorState
	case "n":
		return IncrementCursorState
	case "m":
		return SelectMenuState
	case "q":
		return FinishedState
	}
	// Parse the integers provided: comma separated, hyphen separated, mixture
	selectedCommands := ParseSelection(input)
	fmt.Println("Selections!")
	fmt.Println(selectedCommands)
	// Update selections
	return FinishedState
}

func ParseSelection(selection string) []int {
	selections := []int{}
	commaDelimited := strings.Split(selection, ",")
	for _, s := range commaDelimited {
		if strings.Contains(s, "-") {
			hyphenDelimited := strings.Split(s, "-")
			min, _ := strconv.Atoi(hyphenDelimited[0])
			max, _ := strconv.Atoi(hyphenDelimited[1])
			for i := min; i <= max; i++ {
				selections = append(selections, i)
			}
		} else {
			n, _ := strconv.Atoi(s)
			selections = append(selections, n)
		}
	}
	return selections
}

func HistoryPage(cmds *Commands, start, stop int) {
	i := start
	for ; i < len(*cmds) && i < start+stop; i++ {
		cmd := (*cmds)[i]
		if cmd.Selected {
			fmt.Printf(SelectedCmdStrFmt, cmd.Number, cmd.CmdString)
		} else {
			fmt.Printf(UnselectedCmdStrFmt, cmd.Number, cmd.CmdString)
		}
	}
}

func ParseQuery(s *bufio.Scanner) int {
	switch s.Text() {
	case "1", "p":
		return FinishedState
	case "2", "g":
		return FinishedState
	case "3", "s":
		return SelectMenuState
	case "4", "r":
		return FinishedState
	case "5", "e":
		return FinishedState
	case "6", "q":
		return FinishedState
	}
	return FinishedState
}

func main() {
	// parse a flag for silent execution, skip this step
	state := MainMenuState
	scanner := bufio.NewScanner(os.Stdin)

	var commands *Commands
	selections := []Command{}
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
	cursor := 0
	for state != FinishedState {
		switch state {
		case MainMenuState:
			DisplayMainMenu(&selections)
			state = QueryUserState
		case QueryUserState:
			fmt.Print("Select>>")
			_ = scanner.Scan()
			state = ParseQuery(scanner)
		case SelectMenuState:
			DisplaySelectMenu(scanner, &cursor, commands, &selections)
			state = ParseSelectQuery(scanner, commands, &selections)
		case DecrementCursorState:
			cursor -= pageLengths
			if cursor < 0 {
				cursor = 0
			}
			state = SelectMenuState
		case IncrementCursorState:
			cursor += pageLengths
			if cursor > len(*commands) {
				cursor = len(*commands) - pageLengths
			}
			state = SelectMenuState
		}
	}
}
