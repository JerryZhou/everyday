package main

/*
import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	terminal "github.com/carmark/pseudo-terminal-go/terminal"
	color "github.com/fatih/color"
	ui "github.com/gizak/termui"
	rw "github.com/mattn/go-runewidth"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"
)

// 在当前日志文件里面写入
func DailyWrite(dailyFile string,
	daily string) int64 {
	var (
		err  error
		file *os.File
	)
	file, err = os.OpenFile(dailyFile,
		os.O_APPEND|os.O_CREATE|os.O_RDWR,
		0666)
	if err != nil {
		return 0
	}
	defer file.Close()
	fi, _ := file.Stat()

	if len(daily) > 0 {
		file.WriteString(daily)
	}
	return fi.Size()
}

// 读取日志
func DailyRead(dailyFile string) (daily string, err error) {
	var (
		bytes []byte
	)
	bytes, err = ioutil.ReadFile(dailyFile)
	if err == nil {
		daily = string(bytes)
	}
	return
}

// helper
func IsCmd(input *InputContext, cmd bool, parse bool) bool {
	if cmd && parse {
		input.Parsed()
	}
	return cmd
}

// 基础命令
func BasicInput(input *InputContext) {
	// choose command
	if IsCmd(input, input.Cmd == "cd", false) {
		Exec_M_Cd(input) // 基础 cd
	} else if IsCmd(input, input.Cmd == "ls", false) {
		Exec_M_Ls(input) // 基础 ls
	} else if IsCmd(input, input.Cmd == "open", false) {
		Exec_M_Open(input) // 基础 open
	} else if IsCmd(input, input.Cmd == "search" || input.Cmd == "s", true) {
		Exec_SearchDaily(input)
	} else if IsCmd(input, input.Cmd == "" ||
		input.Cmd == "write" ||
		input.Cmd == "w", true) {
		Exec_WriteDaily(input)
	} else if IsCmd(input, input.Cmd == "del" || input.Cmd == "delete", true) {
		Exec_DeleteDaily(input)
	} else if IsCmd(input, input.Cmd == "md" || input.Cmd == "markdown", true) {
		Exec_MdDaily(input)
	} else if IsCmd(input, input.Cmd == "config" || input.Cmd == "cfg", true) {
		Exec_ConfigDaily(input)
	} else if IsCmd(input, input.Cmd == "console" || input.Cmd == "con", true) {
		Exec_ConsoleDaily(input)
	} else if IsCmd(input, input.Cmd == "clear" || input.Cmd == "cls", true) {
		Exec_Cls(input)
	} else if IsCmd(input, input.Cmd == "xopen" || input.Cmd == "edit", true) {
		Exec_OpenDaily(input)
	} else if IsCmd(input, input.Cmd == "xcd", true) {
		Exec_Xcd(input)
	} else if IsCmd(input, input.Cmd == "xls" || input.Cmd == "show", true) {
		Exec_ShowDaily(input)
	} else {
		RunCommand("", input.Cmd, input.Args...)
	}
}

// 搜索日志
func Exec_SearchDaily(input *InputContext) {
	Controls := input.Controls

	// 命令行到控制参数
	if len(input.Args) > 0 {
		if len(Controls.FlagKeywordDaily) == 0 {
			Controls.FlagKeywordDaily = input.Args[0]
		}
		if len(input.Args) > 1 && len(Controls.FlagDirDaily) == 0 {
			Controls.FlagDirDaily = input.Args[1]
		}
	}

	// 选定关键字和目录
	if len(Controls.FlagKeywordDaily) == 0 {
		fmt.Println("关键字-k不能为空")
		return
	}
	dir := input.Console.HomeDir
	if len(Controls.FlagDirDaily) > 0 {
		dir = input.Console.FullDailyPath(Controls.FlagDirDaily)
	}
	RunCommand("", "grep", "-r", Controls.FlagKeywordDaily, dir)
}

// 列出日志
// -t=now
// -l=count
// -c=category
func Exec_ShowDaily(input *InputContext) {
	context := &DailyContext{}
	if err := PrepareContext(input, context); err != nil {
		fmt.Println("prepare error:", err)
		return
	}
	Controls := input.Controls

	// 从参数选项填充控制项
	if len(input.Args) > 0 {
		// 类别
		Controls.FlagCategoryDaily = input.Args[0]

		// 个数
		if len(input.Args) > 1 {
			if i, err := strconv.Atoi(input.Args[1]); err == nil {
				Controls.FlagLenDaily = i
			}
		}
	}

	var (
		dailys []string
		sign   int
	)
	if Controls.FlagLenDaily > 0 {
		sign = 1
	} else {
		sign = -1
	}
	split := strings.Repeat("#", LineLen)
	// 读取日志
	for count := 0; sign*count != Controls.FlagLenDaily; count++ {
		xnow := context.XNow.AddDate(0, 0, -1*sign*count)
		_, _, dailyFile := DailyPath(input, xnow)
		if daily, err := DailyRead(dailyFile); err == nil {
			dailyFile = input.Console.ShortDailyPath(dailyFile)
			empty := strings.Repeat(" ", (LineLen-10-len(dailyFile))/2)
			xtitle := fmt.Sprintf("#####%v%v%v#####", empty, dailyFile, empty)
			dailys = append(dailys, split)
			dailys = append(dailys, xtitle)
			dailys = append(dailys, daily)
		} else {
			dailyFile = input.Console.ShortDailyPath(dailyFile)
			xnowstr := dailyFile + " (Empty)"
			empty := strings.Repeat(" ", (LineLen-10-len(xnowstr))/2)
			xempty := fmt.Sprintf("#####%v%v%v#####\n", empty, xnowstr, empty)
			dailys = append(dailys, split)
			dailys = append(dailys, xempty)

		}
	}
	// 输出日志
	if len(dailys) > 0 {
		for _, daily := range dailys {
			fmt.Println(daily)
		}
	}
}

// 打开
func Exec_M_Open(input *InputContext) {
	input.Console.Mode().Cmd_open(input)
}

// 打开日志
func Exec_OpenDaily(input *InputContext) {
	Controls := input.Controls
	Controls.FlagNewDaily = false

	if len(input.Args) >= 1 {
		daily := input.Console.FullDailyPath(input.Args[0])
		if _, err := os.Stat(daily); err == nil {
			RunCommand("", "vim", "+", daily)
			if Controls.FlagCategoryDaily == "" {
				RunCommand("", "clear")
			}
		} else {
			fmt.Println("文件不存在",
				input.Console.ShortDailyPath(daily))
		}
	} else {
		Exec_WriteDaily(input)
	}
}

// 删除日志
func Exec_DeleteDaily(input *InputContext) {
	Controls := input.Controls
	context := &DailyContext{}
	if err := PrepareContext(input, context); err != nil {
		fmt.Println("prepare error:", err)
		return
	}
	split := strings.Repeat("#", LineLen)

	// 直接删除
	if len(input.Args) >= 1 {
		dailyFile := input.Console.FullDailyPath(input.Args[0])
		if err := os.Remove(dailyFile); err != nil {
			fmt.Println(split)
			fmt.Println("Remove Daily:",
				input.Console.ShortDailyPath(dailyFile),
				" Failed:", err)
		} else {
			fmt.Println(split)
			fmt.Println("Remove Daily:",
				input.Console.ShortDailyPath(dailyFile),
				" Success")
		}
		return
	}

	// 删除指定范围
	var (
		sign int
	)
	if Controls.FlagLenDaily > 0 {
		sign = 1
	} else {
		sign = -1
	}
	// 删除日志
	for count := 0; sign*count != Controls.FlagLenDaily; count++ {
		xnow := context.XNow.AddDate(0, 0, -1*sign*count)
		_, _, dailyFile := DailyPath(input, xnow)
		if err := os.Remove(dailyFile); err != nil {
			fmt.Println(split)
			fmt.Println("Remove Daily:",
				input.Console.ShortDailyPath(dailyFile), " Failed:", err)
		} else {
			fmt.Println(split)
			fmt.Println("Remove Daily:",
				input.Console.ShortDailyPath(dailyFile), " Success")
		}
	}
}

// 执行 MarkDown 的命令操作
// me md add daily group sub xxx
// me md del daily group sub xxx
// me md done daily group sub xxx
// me md mark daily group sub xxx

type MdFlatLine struct {
	Line string
	Idx  int
}

// 打印
func (line *MdFlatLine) Print() {
	fmt.Print(line.Idx, ":", line.Line)
}

// 做标记
func (line *MdFlatLine) Mark(mark string) {
	if strings.Contains(line.Line, mark) {
		return
	}
	line.Line = line.Line[:len(line.Line)-1] + " " + mark + "\n"
}

// 取消标记
func (line *MdFlatLine) Unmark(mark string) {
	if strings.Contains(line.Line, mark) == false {
		return
	}
	line.Line = strings.Replace(line.Line, mark, "", -1)
}

type MdFlatDaily struct {
	Lines []*MdFlatLine
}

// 从文件里面一行一行读取
func (daily *MdFlatDaily) ReadFrom(name string) (err error) {
	var (
		file   *os.File
		reader *bufio.Reader
	)
	for {
		daily.Lines = make([]*MdFlatLine, 0)

		if file, err = os.Open(name); err != nil {
			break
		}
		reader = bufio.NewReader(file)
		defer file.Close()

		for err == nil {
			mdline := &MdFlatLine{}
			mdline.Line, err = reader.ReadString('\n')
			mdline.Idx = len(daily.Lines)

			if err == nil {
				daily.Lines = append(daily.Lines, mdline)
			} else if err == io.EOF {
				err = nil
				if len(mdline.Line) > 0 {
					daily.Lines = append(daily.Lines, mdline)
				}
				break
			} else {
				break
			}
		}

		break
	}
	return
}

// 把md 反写到文件
func (daily *MdFlatDaily) WriteTo(name string) (err error) {
	var (
		file   *os.File
		writer *bufio.Writer
	)
	for {
		if file, err = os.OpenFile(name, os.O_RDWR|os.O_TRUNC, 0); err != nil {
			break
		}
		defer file.Close()
		writer = bufio.NewWriter(file)
		for _, line := range daily.Lines {
			if _, err = writer.WriteString(line.Line); err != nil {
				break
			}
		}
		err = writer.Flush()
		break
	}
	return
}

// 打印
func (daily *MdFlatDaily) Print() {
	for _, line := range daily.Lines {
		line.Print()
	}
	fmt.Print("\n")
}

// 执行Md 文件的相关操作(TODO)
func Exec_MdDaily(input *InputContext) {
	if len(input.Args) < 2 {
		fmt.Println("参数太少")
		return
	}
	rewrite := true
	cmd := input.Args[0]
	idx := 0
	dailyFile := input.Console.FullDailyPath(input.Args[1])
	flatDaily := &MdFlatDaily{}
	if err := flatDaily.ReadFrom(dailyFile); err != nil {
		fmt.Println("Read Md Err:", err)
		return
	}
	if len(input.Args) >= 3 {
		idx, _ = strconv.Atoi(input.Args[2])
	}
	if cmd == "add" {
		line := &MdFlatLine{}
		line.Idx = idx
		line.Line = input.Args[3] + "\n"
		fmt.Println("##### Add Line:")
		line.Print()

		right := append([]*MdFlatLine{}, flatDaily.Lines[idx:]...)
		left := append(flatDaily.Lines[:idx], line)
		flatDaily.Lines = append(left, right...)
	} else if cmd == "del" {
		fmt.Println("##### Delete Line:")
		flatDaily.Lines[idx].Print()

		// sub - 1
		for _, line := range flatDaily.Lines[idx+1:] {
			line.Idx = line.Idx - 1
		}

		flatDaily.Lines = append(
			flatDaily.Lines[:idx],
			flatDaily.Lines[idx+1:]...)
		flatDaily.Print()
	} else if cmd == "done" {
		// [DONE]
		flatDaily.Lines[idx].Mark("*[DONE]*")
		flatDaily.Lines[idx].Print()
	} else if cmd == "undone" {
		// [DONE]
		flatDaily.Lines[idx].Unmark("*[DONE]*")
		flatDaily.Lines[idx].Print()
	} else if cmd == "mark" {
		flatDaily.Lines[idx].Mark(input.Args[3])
		flatDaily.Lines[idx].Print()
	} else if cmd == "unmark" {
		flatDaily.Lines[idx].Unmark(input.Args[3])
		flatDaily.Lines[idx].Print()
	} else if cmd == "view" {
		rewrite = false
		flatDaily.Print()
	} else {
		fmt.Println("Not Imp Md Cmd:", cmd)
		return
	}

	if rewrite {
		if err := flatDaily.WriteTo(dailyFile); err != nil {
			fmt.Println("Write Md Err:", err)
		}
	}
}

// 执行写入操作
func Exec_WriteDaily(input *InputContext) {
	context := &DailyContext{}
	if err := PrepareContext(input, context); err != nil {
		fmt.Println("prepare error:", err)
		return
	}

	Controls := input.Controls

	if Controls.FlagNewDaily {
		now := time.Now()
		xtime := now.Format(TimeLayout)
		header := "------------------------------------------------------------\n" +
			xtime +
			strings.Repeat(" ", LineLen-len(xtime)-len(input.Console.Author)) +
			input.Console.Author + "\n"
		DailyWrite(context.Daily, header)
	}

	RunCommand("", "vim", "+", context.Daily)

	if Controls.FlagNewDaily {
		endnow := time.Now()
		endtime := endnow.Format(TimeLayout)
		bootstrap := endtime +
			strings.Repeat("-", LineLen-len(endtime)-len(EndMark)) +
			EndMark + "\n\n"

		DailyWrite(context.Daily, bootstrap)
	}

	// cls
	if Controls.FlagCategoryDaily == "" {
		RunCommand("", "clear")
	}
}

// 进入新的模式
func Exec_Xcd(input *InputContext) {
	if len(input.Args) < 1 {
		fmt.Println("参数不够: xcd ..")
		return
	}
	if input.Args[0] == ".." {
		if input.Console.Depth() > 1 {
			input.Console.Pop()
		}
	} else if input.Args[0] == "md" {
		md := &MdInputMode{}
		md.Input = &InputContext{}
		md.Input.Console = input.Console
		input.Console.Push(md)
	}
}

// 编辑配置文件
func Exec_ConfigDaily(input *InputContext) {
	// 解析配置文件
	current, _ := user.Current()
	rc := path.Join(current.HomeDir, EveryDay, ".rc")

	if len(input.Args) >= 2 {
		configfiled := input.Args[0]
		configvalue := input.Args[1]

		if configfiled == "category" || configfiled == "c" {
			// copy it
			Config.DefaultCategory = &configvalue
		}

	} else {
		RunCommand("", "vim", "+", rc)
	}
}


// 普通输入模式
type NormalInputMode struct {
	Input *InputContext
}

// 名字
func (normal *NormalInputMode) Name() string {
	return EveryDayName
}

func (normal *NormalInputMode) Cmd_cd(input *InputContext) {
	if len(input.Args) <= 0 {
		return
	}
	input.Console.Cmd_cd(input.Args[0])
}

func (normal *NormalInputMode) Cmd_ls(input *InputContext) {
	RunCommand("", "ls", input.Args...)
}

func (normal *NormalInputMode) Cmd_open(input *InputContext) {
	RunCommand("", "open", input.Args...)
}

// 自我阐述
func (normal *NormalInputMode) Me(s string) {
	fmt.Println("普通模式", s)
}

// 处理输入
func (normal *NormalInputMode) DealWithInput(s string) {
	input := normal.Input
	if input.ParseCmd(s) == "" {
		return
	}
	BasicInput(input)
}

// 扫描符
func (normal *NormalInputMode) Scanner() string { // 输入符
	return ""
}

// 针对处理 Md
type MdInputMode struct {
	Input *InputContext

	Mdfile string // 当前的Md 文件
	Dir    string
}

// 名字
func (md *MdInputMode) Name() string {
	return "md"
}

func (md *MdInputMode) Cmd_cd(input *InputContext) {
	if len(input.Args) <= 0 {
		return
	}
	input.Console.Cmd_cd(input.Args[0])

}

func (md *MdInputMode) Cmd_ls(input *InputContext) {
	RunCommand("", "ls", input.Args...)
}

func (md *MdInputMode) Cmd_open(input *InputContext) {
	RunCommand("", "open", input.Args...)
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

// 自我阐述
func (md *MdInputMode) Me(s string) {
	fmt.Println("当前的Md 文件: ", md.Mdfile, " ", s)
	fmt.Println("可用的命令有: use view add del done mark unmark")
}

// 日志文件是否存在
func (md *MdInputMode) DailyExits(daily string) (bool, error) {
	if _, err := os.Stat(daily); err != nil {
		return false, err
	} else {
		return true, nil
	}
}

// 处理命令
func (md *MdInputMode) DealWithInput(s string) {
	input := md.Input
	if input.ParseCmd(s) == "" {
		return
	}

	if input.Cmd == "use" {
		input.Parsed()

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
		input.Cmd == "mark" ||
		input.Cmd == "unmark" {
		input.Parsed()

		daily := input.Console.FullDailyPath(md.Mdfile)
		if exits, err := md.DailyExits(daily); exits == false {
			fmt.Println("文件不存在",
				input.Console.ShortDailyPath(daily), err)
			return
		}
		input.Args = append([]string{input.Cmd, md.Mdfile}, input.Args...)

		Exec_MdDaily(input)
	} else if input.Cmd == "me" {
		md.Me("")
	} else {
		BasicInput(input)
	}
}

// 进入everyday 命令行
func Exec_ConsoleDaily(input *InputContext) {
	input.Console.Start()
}

// 清理 console
func Exec_Cls(input *InputContext) {
	RunCommand("", "clear")
}

// 目录操作命令
func Exec_M_Cd(input *InputContext) {
	input.Console.Mode().Cmd_cd(input)
}

// 目录穷举操作
func Exec_M_Ls(input *InputContext) {
	input.Console.Mode().Cmd_ls(input)
}
*/
