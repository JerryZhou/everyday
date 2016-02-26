package cmd

import (
	"fmt"
	daily "jerry.com/everyday/daily"
)

//***********************************
type Cd struct {
	index     int
	completes []string
	current   string
}

func (con *Cd) Name() string {
	return "cd"
}

func (con *Cd) Help() string {
	return "usage: cd dir\n" +
		"change current scanner dir"
}

func (con *Cd) Deal(input *daily.InputContext) {
	if len(input.Args) < 1 {
		fmt.Println(con.Help())
		return
	}
	input.Console.Cmd_cd(input.Args[0])
}

func (con *Cd) AutoCompleteCallback(console *daily.ConsoleDaily, line []byte, pos, key int) (
	newLine []byte, newPos int) {
	newLine = nil
	return
}
