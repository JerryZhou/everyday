package cmd

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	color "github.com/fatih/color"
	"io"
	daily "jerry.com/everyday/daily"
	depends "jerry.com/everyday/depends"
	"os"
	"strconv"
	"strings"
	"time"
	//"unicode"
)

//***********************************
var (
	MarkDone   string = "*[DONE]*"   // 标记已做
	MarkSkip   string = "*[SKIP]*"   // 跳过段落
	MarkRemind string = "*[REMIND]*" // 提醒
)

//***********************************
type Md struct {
}

func (con *Md) Name() string {
	return "md"
}

func (con *Md) Help() string {
	return "usage: md cmd \n" +
		"md todo [-c category] [-n count] [-t 2016-02-14]\n" +
		"md add mdfile 1 content\n" +
		"md del mdfile 1\n" +
		"md done mdfile 1\n" +
		"md undone mdfile 1\n" +
		"md mark mdfile 1 *[DONE]*\n" +
		"md unmark mdfile 1 *[DONE]*\n" +
		"md view mdfile\n" +
		"对md 文件进行相关操作"
}

func (con *Md) Deal(input *daily.InputContext) {
	if len(input.Args) < 1 {
		fmt.Println(con.Help())
		return
	}
	Exec_MdDaily(input)
}

func (con *Md) AutoCompleteCallback(console *daily.ConsoleDaily, line []byte, pos, key int) (
	newLine []byte, newPos int) {
	newLine = nil
	return
}

// 执行 MarkDown 的命令操作
// me md add daily group sub xxx
// me md del daily group sub xxx
// me md done daily group sub xxx
// me md mark daily group sub xxx

//***********************************
type MdFlatLine struct {
	Line   string
	Idx    int
	MdLine *string
}

// 打印
func (line *MdFlatLine) Print() {
	if line.MdLine == nil {
		outs := string(depends.MdOutput([]byte(line.Line)))
		if len(outs) > 0 && outs[0] == '\n' {
			outs = outs[1:]
		}
		line.MdLine = &outs
	}
	fmt.Printf("%02d:%v\n", line.Idx, *line.MdLine)
}

// 列表等级
// 从 1开始
func (line *MdFlatLine) Depth() (depth int) {
	strip := strings.TrimSpace(line.Line)
	if len(strip) == 0 {
		depth = 0
		return
	}
	if strings.HasPrefix(line.Line, "## ") {
		depth = 4
	} else if strings.HasPrefix(line.Line, "*") {
		depth = 8
	} else {
		for _, v := range line.Line {
			if v == rune('\t') {
				depth = depth + 4
			} else if v == rune(' ') {
				depth++
			} else {
				break
			}
		}
		if depth > 0 {
			depth = depth + 8
		}
	}

	depth = (depth / 4)
	return
}

// 是否有标记
func (line *MdFlatLine) Has(mark string) bool {
	return strings.Contains(line.Line, mark)
}

// 做标记
func (line *MdFlatLine) Mark(mark string) {
	if line.Has(mark) {
		return
	}
	line.Line = line.Line + " " + mark
}

// 取消标记
func (line *MdFlatLine) Unmark(mark string) {
	if line.Has(mark) == false {
		return
	}
	line.Line = strings.Replace(line.Line, mark, "", -1)
}

//***********************************
type MdFlatDaily struct {
	Lines []*MdFlatLine
	File  string
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
		daily.File = name

		for err == nil {
			mdline := &MdFlatLine{}
			mdline.Line, err = reader.ReadString('\n')
			mdline.Idx = len(daily.Lines)
			// skip last '\n'
			if mdline.Line != "" && mdline.Line[len(mdline.Line)-1] == '\n' {
				mdline.Line = mdline.Line[:len(mdline.Line)-1]
			}

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
			if _, err = writer.WriteString(line.Line + "\n"); err != nil {
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

// 打印
func (daily *MdFlatDaily) PrintTODO() {
	var (
		part int = 0
	)
	for _, line := range daily.Lines {
		if strings.HasPrefix(line.Line, "## ") {
			if line.Has(MarkSkip) {
				part = 1
				continue
			} else {
				part = 0
			}
		}
		if line.Has(MarkDone) {
			continue
		}
		if line.Has(MarkSkip) {
			continue
		}
		if part == 0 {
			line.Print()
		}
	}
	fmt.Print("\n")
}

// 执行Md 文件的相关操作
func Exec_MdDaily(input *daily.InputContext) {
	cmd := input.Args[0]
	if cmd == "add" {
		Exec_MdDaily_add(input)
	} else if cmd == "del" {
		Exec_MdDaily_del(input)
	} else if cmd == "done" {
		Exec_MdDaily_done(input)
	} else if cmd == "undone" {
		Exec_MdDaily_undone(input)
	} else if cmd == "mark" {
		Exec_MdDaily_mark(input)
	} else if cmd == "unmark" {
		Exec_MdDaily_unmark(input)
	} else if cmd == "view" {
		Exec_MdDaily_view(input)
	} else if cmd == "todo" {
		Exec_MdDaily_todo(input)
	} else if cmd == "skip" {
		Exec_MdDaily_skip(input)
	} else if cmd == "unskip" {
		Exec_MdDaily_unskip(input)
	} else {
		fmt.Println("Not Imp Md Cmd:", cmd)
		return
	}
}

// 控制参数
type MdFlag_todo struct {
	Ts       string
	Range    int
	Category string
}

// 列出所有未完成项目
func Exec_MdDaily_todo(input *daily.InputContext) {
	control := &MdFlag_todo{
		Ts:       daily.DefaultTime(),
		Range:    7,
		Category: daily.DefaultCategory(),
	}

	// [cmd args flags]
	vargs, vsets := daily.ParseFlagSet(input.Args)
	if len(vargs) > 1 {
		control.Category = vargs[1]
		if len(vargs) > 2 {
			control.Ts = vargs[2]
		}
	}
	set := flag.NewFlagSet("todo", flag.ContinueOnError)
	set.StringVar(&control.Ts, "t", control.Ts, "指定日期")
	set.IntVar(&control.Range, "n", control.Range, "指定个数")
	set.StringVar(&control.Category, "c", control.Category, "指定类别")

	set.Parse(vsets)

	var (
		flats []*MdFlatDaily
		sign  int
		xnow  time.Time = time.Now()
	)

	if xnow0, err0 := time.Parse(daily.TimeDailyLayout, control.Ts); err0 == nil {
		xnow = xnow0
	}
	if control.Range > 0 {
		sign = 1
	} else {
		sign = -1
	}
	// 读取日志
	for count := 0; sign*count != control.Range; count++ {
		xnow := xnow.AddDate(0, 0, -1*sign*count)
		_, _, dailyFile := input.Console.DailyPath(control.Category, xnow)

		flatDaily := &MdFlatDaily{}
		if err := flatDaily.ReadFrom(dailyFile); err == nil {
			flats = append(flats, flatDaily)
		}
	}
	split := color.MagentaString(strings.Repeat("#", daily.LineLen))
	// 输出日志
	for _, daily := range flats {
		input.Console.Println(split)
		input.Console.Println(color.MagentaString(input.Console.SprintfCenter(
			input.Console.ShortDailyPath(daily.File))))

		daily.PrintTODO()
	}
}

// 做标记
func Exec_MdDaily_mark(input *daily.InputContext) {
	if len(input.Args) < 4 {
		input.Console.Println("参数太少: mark mdfile line **")
		return
	}

	// 解析日志和行号
	flatDaily, idx, err := mdDaily(input.Console, input.Args[1:])
	if err != nil {
		input.Console.Println(err)
		return
	}

	flatDaily.Lines[idx].Mark(input.Args[3])
	flatDaily.Lines[idx].Print()

	if err := flatDaily.WriteTo(flatDaily.File); err != nil {
		fmt.Println("Write Md Err:", err)
	}
}

// 取消标记
func Exec_MdDaily_unmark(input *daily.InputContext) {
	if len(input.Args) < 4 {
		input.Console.Println("参数太少: mark mdfile line **")
		return
	}

	// 解析日志和行号
	flatDaily, idx, err := mdDaily(input.Console, input.Args[1:])
	if err != nil {
		input.Console.Println(err)
		return
	}

	flatDaily.Lines[idx].Unmark(input.Args[3])
	flatDaily.Lines[idx].Print()
	if err := flatDaily.WriteTo(flatDaily.File); err != nil {
		fmt.Println("Write Md Err:", err)
	}
}

// 在指定的地方增加一行
func Exec_MdDaily_add(input *daily.InputContext) {
	if len(input.Args) < 4 {
		input.Console.Println("参数太少: add mdfile 2 `	*美人鱼首播`")
		return
	}

	// 解析日志和行号
	flatDaily, idx, err := mdDaily(input.Console, input.Args[1:])
	if err != nil {
		input.Console.Println(err)
		return
	}

	line := &MdFlatLine{}
	line.Idx = idx
	line.Line = input.Args[3] + "\n"
	fmt.Println("##### Add Line:")
	line.Print()

	right := append([]*MdFlatLine{}, flatDaily.Lines[idx:]...)
	// sub + 1
	for _, line := range right {
		line.Idx = line.Idx + 1
	}
	left := append(flatDaily.Lines[:idx], line)
	flatDaily.Lines = append(left, right...)
	if err := flatDaily.WriteTo(flatDaily.File); err != nil {
		fmt.Println("Write Md Err:", err)
	}
}

// 删除某行
func Exec_MdDaily_del(input *daily.InputContext) {
	if len(input.Args) < 3 {
		input.Console.Println("参数太少: del mdfile line")
		return
	}

	// 解析日志和行号
	flatDaily, idx, err := mdDaily(input.Console, input.Args[1:])
	if err != nil {
		input.Console.Println(err)
		return
	}

	fmt.Println("##### Delete Line:")
	flatDaily.Lines[idx].Print()

	// sub - 1
	for _, line := range flatDaily.Lines[idx+1:] {
		line.Idx = line.Idx - 1
	}

	flatDaily.Lines = append(
		flatDaily.Lines[:idx],
		flatDaily.Lines[idx+1:]...)
	if err := flatDaily.WriteTo(flatDaily.File); err != nil {
		fmt.Println("Write Md Err:", err)
	}
}

// 标记已做
func Exec_MdDaily_done(input *daily.InputContext) {
	input.Args[0] = "mark"
	input.Args = append(input.Args, MarkDone)
	Exec_MdDaily_mark(input)
}

// 标记已做
func Exec_MdDaily_undone(input *daily.InputContext) {
	input.Args[0] = "unmark"
	input.Args = append(input.Args, MarkDone)
	Exec_MdDaily_unmark(input)
}

// 标记已做
func Exec_MdDaily_skip(input *daily.InputContext) {
	input.Args[0] = "mark"
	input.Args = append(input.Args, MarkSkip)
	Exec_MdDaily_mark(input)
}

// 标记已做
func Exec_MdDaily_unskip(input *daily.InputContext) {
	input.Args[0] = "unmark"
	input.Args = append(input.Args, MarkSkip)
	Exec_MdDaily_unmark(input)
}

// 展示
func Exec_MdDaily_view(input *daily.InputContext) {
	if len(input.Args) < 2 {
		input.Console.Println("参数太少: view mdfile")
		return
	}

	dailyFile := input.Console.FullDailyPath(input.Args[1])
	flatDaily := &MdFlatDaily{}
	if err := flatDaily.ReadFrom(dailyFile); err != nil {
		input.Console.Println("Read Md Err:", err)
		return
	}
	flatDaily.Print()
}

// 解析日志文件和行号 (arg[0] = mdfile ; arg[1] = line)
func mdDaily(console *daily.ConsoleDaily, args []string) (
	flatDaily *MdFlatDaily, idx int, err error) {
	if len(args) < 2 {
		err = errors.New("参数太少")
		return
	}
	flatDaily = &MdFlatDaily{}
	dailyFile := console.FullDailyPath(args[0])
	if err = flatDaily.ReadFrom(dailyFile); err != nil {
		return
	}

	i := 0
	if i, err = strconv.Atoi(args[1]); err == nil {
		idx = i
	} else {
		err = errors.New(fmt.Sprintf("第三个参数必须是行号:%v", args[1]))
		return
	}
	if idx < 0 || idx >= len(flatDaily.Lines) {
		err = errors.New(fmt.Sprintf("行号:%d 超出范围了[0, %d]", idx, len(flatDaily.Lines)))
	}

	return
}
