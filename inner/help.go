package inner

import (
	daily "jerry.com/everyday/daily"
)

//***********************************
type Help struct {
}

func (con *Help) Name() string {
	return "help"
}

func (con *Help) Help() string {
	return "usage: xcd mode\n" +
		"change current mode dir"
}

func (con *Help) Deal(input *daily.InputContext) {
	if len(input.Args) < 1 {
		for _, cmd := range input.Console.Commands {
			//input.Console.Println()
			input.Console.Println(input.Console.V.Split)
			input.Console.Println(cmd.Help())
		}
		return
	} else {
		if cmd, ok := input.Console.Commands[input.Args[0]]; ok {
			input.Console.Println(cmd.Help())
		} else {
			input.Console.Println("没有找到命令", input.Args[0])
		}
	}
}

func (con *Help) AutoCompleteCallback(console *daily.ConsoleDaily, line []byte, pos, key int) (
	newLine []byte, newPos int) {
	newLine = nil
	return
}
