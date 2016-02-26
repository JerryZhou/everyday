package cmd

import (
	daily "jerry.com/everyday/daily"
	"strings"
)

//***********************************
type Config struct {
}

func (con *Config) Name() string {
	return "config"
}

func (con *Config) Help() string {
	return "usage: config [key=value]\n" +
		"config everyday rc"
}

func (con *Config) Deal(input *daily.InputContext) {
	if len(input.Args) == 0 {
		daily.RunCommand("", "vim", "+", daily.ConfigRc())
	} else {
		if input.Args[0] == "see" {
			input.Console.Println(daily.GiveMeJson(daily.Config))
		} else {
			for _, v := range input.Args {
				vargs := strings.Split(v, "=")
				if len(vargs) != 2 {
					continue
				}
				daily.Config.Set(vargs[0], vargs[1])
			}
		}

	}
}

func (con *Config) AutoCompleteCallback(console *daily.ConsoleDaily, line []byte, pos, key int) (
	newLine []byte, newPos int) {
	newLine = nil
	return
}
