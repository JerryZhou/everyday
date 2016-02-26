package cmd

import (
	daily "jerry.com/everyday/daily"
)

// console#Command
type Con struct {
}

func (con *Con) Name() string {
	return "con"
}

func (con *Con) Help() string {
	return "usage: me con\n" +
		"enter the everyday terminal"
}

func (con *Con) Deal(input *daily.InputContext) {
	input.Console.Start()
}

func (con *Con) AutoCompleteCallback(console *daily.ConsoleDaily, line []byte, pos, key int) (
	newLine []byte, newPos int) {
	newLine = nil
	return
}
