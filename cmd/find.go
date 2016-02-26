package cmd

import (
	daily "jerry.com/everyday/daily"
)

//***********************************
type Find struct {
}

func (con *Find) Name() string {
	return "find"
}

func (con *Find) Help() string {
	return "usage: find keyword\n" +
		"clear current screen"
}

func (con *Find) Deal(input *daily.InputContext) {
	if len(input.Args) < 1 {
		input.Console.Println(con.Help())
		return
	}
	// 选定关键字和目录
	dir := input.Console.Cmd_pwd()
	daily.RunCommand("", "grep", "-r", input.Args[0], dir)
}

func (con *Find) AutoCompleteCallback(console *daily.ConsoleDaily, line []byte, pos, key int) (
	newLine []byte, newPos int) {
	newLine = nil
	return
}
