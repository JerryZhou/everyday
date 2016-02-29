package daily

import (
	"bufio"
	"errors"
	"fmt"
	terminal "github.com/carmark/pseudo-terminal-go/terminal"
	color "github.com/fatih/color"
	ui "github.com/gizak/termui"
	"io"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"
)

// 本地命令
type Command interface {
	Deal(input *InputContext)
	Name() string
	Help() string
	AutoCompleteCallback(console *ConsoleDaily, line []byte, pos, key int) (newLine []byte, newPos int)
}

// 负责console 的管理
type ConsoleDaily struct {
	LineLen int          // 行宽
	V       ConsoleConst // 常量

	Modes []InputMode

	Term *terminal.Terminal

	// 作者
	Author string
	// 存储目录
	HomeDir string

	// 本地路径
	VPaths []string

	// 所有本地命令
	Commands map[string]Command
}

// 简单的命令处理
func (console *ConsoleDaily) LocalCommand(input *InputContext) {
	cmd, ok := console.Commands[input.Cmd]
	if ok {
		cmd.Deal(input)
	} else {
		RunCommand("", input.Cmd, input.Args...)
	}
}

// 增加命令处理
func (console *ConsoleDaily) AddCommand(cmd Command) {
	console.Commands[cmd.Name()] = cmd
}

// 去掉根路径的短路径
func (console *ConsoleDaily) ShortDailyPath(daily string) string {
	return strings.TrimPrefix(daily, console.HomeDir)
}

// 补齐跟路径的全路径
func (console *ConsoleDaily) FullDailyPath(daily string) string {
	if strings.HasPrefix(daily, console.HomeDir) {
		return daily
	}
	return path.Join(console.HomeDir, daily)
}

// 初始化 console
func (console *ConsoleDaily) Open() {
	console.Commands = make(map[string]Command)

	current, _ := user.Current()
	// 作者
	if Config.Author != nil {
		console.Author = *Config.Author
	} else {
		console.Author = current.Username
	}

	// 存储目录
	if Config.HomeDir != nil {
		console.HomeDir = *Config.HomeDir
	} else {
		console.HomeDir = path.Join(current.HomeDir, EveryDay)
	}

	// 当前工作目录
	if err := os.Chdir(console.HomeDir); err != nil {
		panic(err)
	}

	// 默认的模式和输入参数接收器
	console.Modes = []InputMode{}

	// 常量
	console.V.InitWith(LineLen)
}

// 关闭 console
func (console *ConsoleDaily) Close() {
}

// 打印首页参数
func (console *ConsoleDaily) PrintScanner() {
	if console.Term == nil {
		console.Print(console.Scanner())
	} else {
		console.Term.SetPrompt(console.Scanner())
	}
}

// 打印
func (console *ConsoleDaily) Print(a ...interface{}) {
	fmt.Print(a...)
}

// 打印
func (console *ConsoleDaily) Println(a ...interface{}) {
	fmt.Println(a...)
}

// 格式化
func (console *ConsoleDaily) Sprintf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

// 格式化到中间输出
func (console *ConsoleDaily) SprintfCenter(format string, a ...interface{}) string {
	s := console.Sprintf(format, a...)
	slen := StringWidth(s)
	sspace := console.V.LineLen -
		slen -
		console.V.StartLen -
		console.V.EndLen
	spaceleft := strings.Repeat(console.V.Space, sspace/2)
	spaceright := strings.Repeat(console.V.Space,
		sspace-sspace/2)

	s = fmt.Sprintf("%v%v%v%v%v",
		console.V.Start,
		spaceleft, s, spaceright,
		console.V.End)
	return s
}

// 在行中央打印
func (console *ConsoleDaily) PrintCenterln(a ...interface{}) {
	s := fmt.Sprint(a...)
	slen := StringWidth(s)
	sspace := console.V.LineLen -
		slen -
		console.V.StartLen -
		console.V.EndLen
	spaceleft := strings.Repeat(console.V.Space, sspace/2)
	spaceright := strings.Repeat(console.V.Space,
		sspace-sspace/2)

	fmt.Printf("%v%v%v%v%v\n",
		console.V.Start,
		spaceleft, color.GreenString(s), spaceright,
		console.V.End)
}

// 处理输入参数
func (console *ConsoleDaily) Enter(s string) {
	console.Mode().DealWithInput(s)
}

// 返回深度
func (console *ConsoleDaily) Depth() int {
	return len(console.Modes)
}

// 当前当前模式
func (console *ConsoleDaily) Mode() InputMode {
	return console.Modes[len(console.Modes)-1]
}

// 回退一层
func (console *ConsoleDaily) Pop() int {
	console.Modes = console.Modes[:len(console.Modes)-1]

	return console.Depth()
}

// 前进一层
func (console *ConsoleDaily) Push(input InputMode) int {
	console.Modes = append(console.Modes, input)
	return console.Depth()
}

// 目录操作
func (console *ConsoleDaily) Cmd_cd(arg string) {
	if arg == ".." {
		depth := len(console.VPaths)
		if depth > 0 {
			console.VPaths = console.VPaths[:depth-1]
		}
	} else {
		ospath := append(console.VPaths, arg)
		if fi, _ := os.Stat(path.Join(console.HomeDir, path.Join(ospath...))); fi != nil && fi.IsDir() {
			console.VPaths = append(console.VPaths, arg)
		} else {
			console.Println("目录不存在")
			return
		}
	}

	os.Chdir(console.Cmd_pwd())
}

// 当前目录
func (console *ConsoleDaily) Cmd_pwd() string {
	return path.Join(console.HomeDir, path.Join(
		console.VPaths...))
}

// scanner : cmds + paths + mode.scanner
func (console *ConsoleDaily) Scanner() string {
	cmd := []string{}
	for _, m := range console.Modes {
		cmd = append(cmd, m.Name())
	}
	// cmds
	cmds := color.BlueString("(" + strings.Join(cmd, ":") + ")")
	// paths
	paths := color.GreenString("%v%v",
		path.Join(console.VPaths...), "> ")
	// mode
	mode := console.Modes[len(console.Modes)-1].Scanner()

	return cmds + paths + mode
}

// 第三方的loop
// NB! 不可以处理中文输入
func (console *ConsoleDaily) TerminalLoop() {
	term, err := terminal.NewWithStdInOut()
	if err != nil {
		panic(err)
	}
	defer term.ReleaseFromStdInOut()

	// 自动完成
	completeinput := &InputContext{}
	completeinput.Console = console
	term.AutoCompleteCallback = func(line []byte, pos, key int) (newLine []byte, newPos int) {
		if key == '\t' {
			input := string(line)
			cmd := completeinput.ParseCmd(input)
			command, ok := console.Commands[cmd]
			if ok {
				newLine, newPos = command.AutoCompleteCallback(console, line, pos, key)
				return
			}
		}
		newLine = nil
		newPos = 0
		return
	}

	console.Term = term
	for {
		console.PrintScanner()
		// scaning input
		line, err := console.Term.ReadLine()
		if err == io.EOF {
			break
		}
		if (err != nil && strings.Contains(err.Error(), "control-c break")) || len(line) == 0 {
			continue
		}
		if line == "q" || line == "exit" {
			console.Println("Bye!!")
			break
		}
		console.Enter(line)
	}
}

// 通过buf io 进行loop
// NB! 可以处理中文输入
func (console *ConsoleDaily) BufIOLoop() {
	console.PrintScanner()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		if input == "q" || input == "exit" {
			console.Println("Bye!!")
			break
		}
		console.Enter(input)
		console.PrintScanner()
	}
}

// 进入 console
func (console *ConsoleDaily) Start() {
	// clear to top
	RunCommand("", "clear")

	// print version ： 用 FgCyan
	console.Println(console.V.Split)
	console.PrintCenterln("  ")
	console.PrintCenterln("welcome to " + EveryDay + " " + EveryDayVer)
	console.PrintCenterln(" 按q 或者exit 退出 ")
	console.PrintCenterln("  ")
	console.Println(console.V.Split)

	// 可以处理 历史记录，可以处理退格等
	console.BufIOLoop()
}

// 获取日志详细的路径
func (console *ConsoleDaily) DailyPath(category string, now time.Time) (
	dailyPath, dailyName, daily string) {
	// 输入参数
	year, month, day := now.Date()
	// 日志目录
	dailyPath = path.Join(console.HomeDir,
		strconv.Itoa(year),
		fmt.Sprintf("%02d", int(month)),
	)
	// 日志文件
	dailyName = strconv.Itoa(day) +
		"-" +
		now.Weekday().String() +
		EveryDayFormat
	// 不同目录的日志记录在不同文件夹
	if len(category) > 0 {
		dailyPath = path.Join(dailyPath, category)
	}

	// 确保父目录存在
	daily = path.Join(dailyPath, dailyName)
	return
}

// 从路径解析到响应的类别，日期
func (console *ConsoleDaily) DailyDetail(daily string) (
	category string, now time.Time, err error) {

	daily = console.ShortDailyPath(daily)
	// path/2016/02/category/14-WeakDay.md
	divs := strings.Split(daily, string(os.PathSeparator))
	divl := len(divs)
	if len(divs) < 4 {
		err = errors.New("路径错误:" + daily)
		return
	}

	var (
		year, month, day int
	)
	year, err = strconv.Atoi(divs[divl-4])
	if err != nil {
		return
	}
	month, err = strconv.Atoi(divs[divl-3])
	if err != nil {
		return
	}
	category = divs[divl-2]
	if _, err = strconv.Atoi(category); err == nil {
		return
	} else {
		err = nil
	}
	divdays := strings.Split(divs[divl-1], "-")
	if len(divdays) != 2 {
		err = errors.New("个数错误应该是: day-weekday.md")
		return
	}
	day, err = strconv.Atoi(divdays[0])
	if err != nil {
		err = errors.New("格式错误-日期")
		return
	}
	checkt := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	if (checkt.Weekday().String() + ".md") != divdays[1] {
		err = errors.New("格式错误:" + checkt.String() + " != " + divdays[1])
		return
	}

	// 日期
	now = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	return
}

// 基于 termui 的展示UI
func (console *ConsoleDaily) Termui() {
	if err := ui.Init(); err != nil {
		console.Println("UI Init Err:", err)
		return
	}
	defer ui.Close()

	ui.NewLineChart()

	g := ui.NewPar("")
	g.Width = 60
	g.Height = 4
	g.Text = EveryDay + " " + EveryDayVer + "\n" +
		"日志管理- 子命令: ls, open, del, md"
	g.BorderLabel = EveryDay
	g.BorderFg = ui.ColorYellow

	strs := []string{
		"[0] github.com/gizak/termui",
		"[1] [你好，世界](fg-blue)",
		"[2] [こんにちは世界](fg-red)",
		"[3] [color output](fg-white,bg-green)",
		"[4] output.go",
		"[5] random_out.go",
		"[6] dashboard.go",
		"[7] nsf/termbox-go"}

	ls := ui.NewList()
	ls.Items = strs
	ls.ItemFgColor = ui.ColorYellow
	ls.BorderLabel = "List"
	ls.Height = 7
	ls.Width = 35
	ls.Y = 4

	ui.Render(g, ls)

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		ui.StopLoop()
	})

	ui.Loop()

}
