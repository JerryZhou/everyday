package mode

import (
	daily "jerry.com/everyday/daily"
)

// 普通输入模式
type NormalInputMode struct {
	Input *daily.InputContext
}

// 名字
func (normal *NormalInputMode) Name() string {
	return daily.EveryDayName
}

// 处理输入
func (normal *NormalInputMode) DealWithInput(s string) {
	input := normal.Input
	if input.ParseCmd(s) == "" {
		return
	}
	input.Console.LocalCommand(input)
}

// 扫描符
func (normal *NormalInputMode) Scanner() string { // 输入符
	return ""
}
