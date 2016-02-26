package cmd

import (
	daily "jerry.com/everyday/daily"
)

//***********************************
type Mail struct {
}

func (con *Mail) Name() string {
	return "mail"
}

func (con *Mail) Help() string {
	return "usage: mail\n" +
		"clear current screen"
}

func (con *Mail) Deal(input *daily.InputContext) {
	daily.RunCommand("", "clear")
}

func (con *Mail) AutoCompleteCallback(console *daily.ConsoleDaily, line []byte, pos, key int) (
	newLine []byte, newPos int) {
	newLine = nil
	return
}
