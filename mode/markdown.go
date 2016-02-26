package mode

import (
	"fmt"
	color "github.com/fatih/color"
	cmd "jerry.com/everyday/cmd"
	daily "jerry.com/everyday/daily"
	"os"
)

// 针对处理 Md
type MdInputMode struct {
	Input *daily.InputContext

	Mdfile string // 当前的Md 文件
	Dir    string
}

// 名字
func (md *MdInputMode) Name() string {
	return "md"
}

// 扫描符
func (md *MdInputMode) Scanner() string { // 输入符
	// input
	if len(md.Mdfile) > 0 {
		return color.CyanString(md.Mdfile + "$ ")
	} else {
		return ""
	}
}

// 处理命令
func (md *MdInputMode) DealWithInput(s string) {
	input := md.Input
	if input.ParseCmd(s) == "" {
		return
	}

	if input.Cmd == "use" {
		if len(input.Args) < 1 {
			fmt.Println("参数太少: use mdfile")
			return
		}
		daily := input.Console.FullDailyPath(input.Args[0])
		if exits, err := md.DailyExits(daily); exits == false {
			fmt.Println("文件不存在",
				input.Console.ShortDailyPath(daily), err)
		} else {
			md.Mdfile = input.Console.ShortDailyPath(daily)
		}
	} else if input.Cmd == "view" ||
		input.Cmd == "add" ||
		input.Cmd == "del" ||
		input.Cmd == "done" ||
		input.Cmd == "undone" ||
		input.Cmd == "skip" ||
		input.Cmd == "unskip" ||
		input.Cmd == "mark" ||
		input.Cmd == "unmark" {

		daily := input.Console.FullDailyPath(md.Mdfile)
		if exits, err := md.DailyExits(daily); exits == false {
			fmt.Println("文件不存在",
				input.Console.ShortDailyPath(daily), err)
			return
		}
		if md.Mdfile != "" {
			input.Args = append([]string{input.Cmd, md.Mdfile}, input.Args...)
		} else {
			input.Args = append([]string{input.Cmd}, input.Args...)
		}

		cmd.Exec_MdDaily(input)
	} else if input.Cmd == "todo" {
		input.Args = append([]string{input.Cmd}, input.Args...)
		cmd.Exec_MdDaily(input)
	} else {
		input.Console.LocalCommand(input)
	}
}

// 日志文件是否存在
func (md *MdInputMode) DailyExits(daily string) (bool, error) {
	if _, err := os.Stat(daily); err != nil {
		return false, err
	} else {
		return true, nil
	}
}
