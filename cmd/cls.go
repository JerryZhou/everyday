package cmd

import (
	daily "jerry.com/everyday/daily"
)

//***********************************
type Cls struct {
}

func (con *Cls) Name() string {
	return "cls"
}

func (con *Cls) Help() string {
	return "usage: cls\n" +
		"clear current screen"
}

func (con *Cls) Deal(input *daily.InputContext) {
	daily.RunCommand("", "clear")
}

func (con *Cls) AutoCompleteCallback(console *daily.ConsoleDaily, line []byte, pos, key int) (
	newLine []byte, newPos int) {
	newLine = nil
	return
}
