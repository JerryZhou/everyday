package main

import (
	cmd "jerry.com/everyday/cmd"
	daily "jerry.com/everyday/daily"
	inner "jerry.com/everyday/inner"
	mode "jerry.com/everyday/mode"
	"os"
	"strings"
)

// default config "~/.everyday"
func main() {
	daily.PrepareConfig()

	console := &daily.ConsoleDaily{}
	console.Open()
	defer console.Close()

	// input
	normal := &mode.NormalInputMode{}
	normal.Input = &daily.InputContext{}
	normal.Input.Console = console
	console.Push(normal)

	// cmds
	console.AddCommand(&cmd.Con{})
	console.AddCommand(&cmd.Cls{})
	console.AddCommand(&cmd.Cd{})
	console.AddCommand(&cmd.Md{})
	console.AddCommand(&cmd.Rd{})
	console.AddCommand(&cmd.Find{})
	console.AddCommand(&cmd.Today{})
	console.AddCommand(&cmd.Config{})
	console.AddCommand(&cmd.Remind{})
	console.AddCommand(&cmd.Mail{})
	console.AddCommand(&cmd.Code{})

	console.AddCommand(&inner.Xcd{})

	console.Enter(strings.Join(os.Args[1:], " "))
}
