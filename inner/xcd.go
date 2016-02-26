package inner

import (
	daily "jerry.com/everyday/daily"
	mode "jerry.com/everyday/mode"
)

//***********************************
type Xcd struct {
}

func (con *Xcd) Name() string {
	return "xcd"
}

func (con *Xcd) Help() string {
	return "usage: xcd mode\n" +
		"change current mode dir"
}

func (con *Xcd) Deal(input *daily.InputContext) {
	if len(input.Args) < 1 {
		input.Console.Println(con.Help())
		return
	}
	if input.Args[0] == ".." {
		if input.Console.Depth() > 1 {
			input.Console.Pop()
		}
	} else if input.Args[0] == "md" {
		md := &mode.MdInputMode{}
		md.Input = &daily.InputContext{}
		md.Input.Console = input.Console
		input.Console.Push(md)
	} else if input.Args[0] == "note" {
		im := &mode.NoteInputMode{}
		im.Input = &daily.InputContext{}
		im.Input.Console = input.Console
		input.Console.Push(im)
	} else if input.Args[0] == "ed" {
		ed := &mode.NormalInputMode{}
		ed.Input = &daily.InputContext{}
		ed.Input.Console = input.Console
		input.Console.Push(ed)
	} else {
		input.Console.Println("Not Support Mode:", input.Args[0])
	}
}

func (con *Xcd) AutoCompleteCallback(console *daily.ConsoleDaily, line []byte, pos, key int) (
	newLine []byte, newPos int) {
	newLine = nil
	return
}
