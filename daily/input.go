package daily

import (
	"flag"
)

// 输入模式
type InputMode interface {
	Name() string // 模式的名字

	DealWithInput(s string) // 处理输入

	Scanner() string // 输入符
}

// 输入上下文
type InputContext struct {
	Level int

	Input        string
	EditCommands []string

	Cmd     string
	Args    []string
	FlagSet *flag.FlagSet

	Controls *FlagControl

	// 承载的日志
	Console *ConsoleDaily
}

// 解析命令
func (input *InputContext) ParseCmd(s string) string {
	input.Input = s
	input.Cmd, input.EditCommands, _ = ParseEditorCommand(s)
	if len(input.EditCommands) > 0 {
		input.Args = input.EditCommands[1:]
	} else {
		input.Args = []string{}
	}
	return input.Cmd
}

// 解析命令行
func (input *InputContext) Parsed() bool {
	// Clear Old
	input.Cmd = ""
	input.Args = []string{}
	input.FlagSet = flag.NewFlagSet("controls", flag.ContinueOnError)

	// 命令行控制项
	input.Controls = &FlagControl{
		FlagNewDaily:      true,
		FlagTimeDaily:     "",
		FlagCategoryDaily: "",
		FlagLenDaily:      1,
		FlagKeywordDaily:  "",
		FlagRangeDaily:    1,
		FlagDirDaily:      "",
	}
	Controls := input.Controls

	// 配置了目录
	if Config.DefaultCategory != nil {
		Controls.FlagCategoryDaily = *Config.DefaultCategory
	}

	// nil or empty input
	if len(input.EditCommands) == 0 {
		return false
	}
	// parse cmd and flags
	args := input.EditCommands
	argx := 0
	argy := 0

	// 原始命令和参数
	// 命令行参数
	FlagSet := input.FlagSet

	FlagSet.BoolVar(&Controls.FlagNewDaily,
		"new",
		Controls.FlagNewDaily,
		"是否新的日志")

	FlagSet.StringVar(&Controls.FlagTimeDaily,
		"time",
		Controls.FlagTimeDaily,
		"指定日期2016-01-02")
	FlagSet.StringVar(&Controls.FlagTimeDaily,
		"t",
		Controls.FlagTimeDaily,
		"指定日期2016-01-02")

	FlagSet.StringVar(&Controls.FlagCategoryDaily,
		"category",
		Controls.FlagCategoryDaily,
		"指定任务族")
	FlagSet.StringVar(&Controls.FlagCategoryDaily,
		"c",
		Controls.FlagCategoryDaily,
		"指定任务族")

	FlagSet.IntVar(&Controls.FlagLenDaily,
		"len",
		Controls.FlagLenDaily,
		"指定个数")
	FlagSet.IntVar(&Controls.FlagLenDaily,
		"l",
		Controls.FlagLenDaily,
		"指定个数")
	FlagSet.IntVar(&Controls.FlagLenDaily,
		"n",
		Controls.FlagLenDaily,
		"指定个数")

	FlagSet.IntVar(&Controls.FlagRangeDaily,
		"range",
		Controls.FlagRangeDaily,
		"指定范围")
	FlagSet.IntVar(&Controls.FlagRangeDaily,
		"r",
		Controls.FlagRangeDaily,
		"指定范围")

	FlagSet.StringVar(&Controls.FlagKeywordDaily,
		"keyword",
		Controls.FlagKeywordDaily,
		"指定任务族")
	FlagSet.StringVar(&Controls.FlagKeywordDaily,
		"k",
		Controls.FlagKeywordDaily,
		"指定任务族")

	FlagSet.StringVar(&Controls.FlagDirDaily,
		"dir",
		Controls.FlagDirDaily,
		"指定目录")
	FlagSet.StringVar(&Controls.FlagDirDaily,
		"d",
		Controls.FlagDirDaily,
		"指定目录")

	for i, v := range args {
		if len(v) > 0 && v[0] == '-' {
			FlagSet.Parse(args[i:])
			break
		} else if i == 0 {
			input.Cmd = v
		} else {
			if argy == 0 {
				argx = i
			}
			argy = i + 1
		}
	}
	input.Args = args[argx:argy]
	return true
}
