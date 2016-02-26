package mode

import (
	cmd "jerry.com/everyday/cmd"
	daily "jerry.com/everyday/daily"
)

// 普通输入模式
type NoteInputMode struct {
	Input *daily.InputContext
}

// 名字
func (im *NoteInputMode) Name() string {
	return "note"
}

// 处理输入
func (im *NoteInputMode) DealWithInput(s string) {
	input := im.Input
	if input.ParseCmd(s) == "" {
		return
	}
	if input.Cmd == "ls" ||
		input.Cmd == "del" ||
		input.Cmd == "edit" {

		input.Args = append([]string{input.Cmd}, input.Args...)

		cmd.Exec_RdDaily(input)
	} else {
		input.Console.LocalCommand(input)
	}
}

// 扫描符
func (im *NoteInputMode) Scanner() string { // 输入符
	return ""
}
