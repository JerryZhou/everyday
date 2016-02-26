package daily

import (
	"strings"
)

// 常量
var (
	EveryDayVer    = "1.0.0 beta"
	EveryDay       = ".everyday"
	EveryDayName   = "everyday"
	EveryDayFormat = ".md"
	EveryDayRc     = ".rc"

	LineLen = 60

	EndMark         = "[END OF EDIT]"
	TimeLayout      = "2006-01-02 15:04:05"
	TimeDailyLayout = "2006-01-02"
)

// 命令行常量
type ConsoleConst struct {
	Spot  string // 分割符
	Space string // 空白符

	Split string // 分割行

	Start    string // 行起始
	StartLen int

	End    string // 行结束
	EndLen int

	LineLen int // 行宽
}

// 初始化
func (cconst *ConsoleConst) InitWith(linewidth int) {
	// init the const vars
	cconst.LineLen = LineLen // 60

	cconst.Space = " "

	cconst.Spot = "#"
	cconst.Split = strings.Repeat(
		cconst.Spot, cconst.LineLen)

	cconst.StartLen = 5
	cconst.Start = strings.Repeat(
		cconst.Spot, cconst.StartLen)

	cconst.EndLen = 5
	cconst.End = strings.Repeat(
		cconst.Spot, cconst.EndLen)
}
