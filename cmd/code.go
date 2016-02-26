package cmd

import (
	daily "jerry.com/everyday/daily"
)

//***********************************
type Code struct {
}

func (con *Code) Name() string {
	return "code"
}

func (con *Code) Help() string {
	return "usage: code\n" +
		"code current user"
}

func (con *Code) Deal(input *daily.InputContext) {
	daily.RunCommand("", "svnup")
}

func (con *Code) AutoCompleteCallback(console *daily.ConsoleDaily, line []byte, pos, key int) (
	newLine []byte, newPos int) {
	newLine = nil
	return
}
