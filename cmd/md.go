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
	"regexp"
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
// 行的状态
type MdLineState struct {
	Depth int
	State int
	Path  []int

	Child        int // 孩子总数
	Todo         int // 需要做的个数
	MaxChildLine int // 最后一个孩子的行号
}

// 具体行
type MdFlatLine struct {
	Line   string
	Idx    int
	MdLine *string

	State *MdLineState
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

// 打印
func (line *MdFlatLine) PrintWithState() {
	if line.MdLine == nil {
		outs := string(depends.MdOutput([]byte(line.Line)))
		if len(outs) > 0 && outs[0] == '\n' {
			outs = outs[1:]
		}
		line.MdLine = &outs
	}
	fmt.Printf("%02d:%-.18s \t\t(%v)\n", line.Idx, *line.MdLine, daily.GiveMeJson(line.State))
}

var (
	DepthWidth = 4
)

// 列表等级
// 从 1开始
// TODO: refract this
func (line *MdFlatLine) Depth() (depth int) {
	strip := strings.TrimSpace(line.Line)
	if len(strip) == 0 {
		depth = 0
		return
	}
	if strings.HasPrefix(line.Line, "# ") ||
		(strings.HasPrefix(line.Line, "---") && line.Idx == 0) {
		depth = 1 * DepthWidth
	} else if strings.HasPrefix(line.Line, "## ") {
		depth = 2 * DepthWidth
	} else if strings.HasPrefix(line.Line, "*") {
		depth = 3 * DepthWidth
	} else if b, _ := regexp.MatchString("^[1-9]\\. [^\n]*", line.Line); b {
		depth = 3 * DepthWidth
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
			depth = depth + 3*DepthWidth
		}
	}

	depth = (depth / DepthWidth)
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

	MaxValidIdx int // 最大的有效内容行
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
				} else {
					break
				}
			} else {
				break
			}
		}

		// 解析行的状态
		daily.ParseLineState()
		break
	}
	return
}

// depth [1, ]
func cal_parent(
	depth int,
	visits []int,
	lines []*MdFlatLine) (
	parent *MdFlatLine) {
	for n := depth - 1; n > 0; n-- {
		m := visits[n]
		if m != -1 {
			parent = lines[m]
			break
		}
	}
	return
}

// depth [d, len(pp)]
func cal_end_parent(
	depth int,
	parent *MdFlatLine,
	parents []*MdFlatLine,
	lines []*MdFlatLine,
	idx int) {
	parents[depth] = parent
	for n := depth + 1; n < len(parents); n++ {
		p := parents[n]
		if p == nil {
			break
		}
		parents[n] = nil
		p.State.MaxChildLine = idx

		// 追溯父亲的完成状态
		cal_backtrace_parent(p, lines)
	}
}

// 追溯父亲的完成状态
func cal_backtrace_parent(p *MdFlatLine, lines []*MdFlatLine) {
	if p.State.Child > 0 &&
		p.State.Todo == 0 &&
		(p.State.State&RStateDone == 0) {
		// 把节点从undone 变成 done
		p.State.State = p.State.State | RStateDone
		// 没有done 也没有skip 则把父亲上面的todo -1
		if p.State.State&RStateSkip == 0 {
			// 父节点的todo 要减1
			// path = [..., parent, self]
			for i := len(p.State.Path) - 2; i >= 0; i-- {
				parentidx := p.State.Path[i]
				if parentidx == -1 {
					continue
				}
				parent := lines[parentidx]
				parent.State.Todo--
			}
		}
	}
}

// 填写访问记录
func cal_assign_visit(
	depth int,
	visits []int,
	idx int) {
	visits[depth] = idx
	for n := depth + 1; n < len(visits); n++ {
		if visits[n] == -1 {
			break
		}
		visits[n] = -1
	}
}

// 解析行的状态
func (md *MdFlatDaily) ParseLineState() {
	maxline := len(md.Lines) - 1
	if maxline < 0 {
		return
	}
	visits := make([]int, 8) // 访问记录
	parents := make([]*MdFlatLine, 8)
	sliceAssign(visits, -1)
	validIdx := 0 // 最后一个有效行

	for _, line := range md.Lines {
		line.State = &MdLineState{}
		line.State.Depth = line.Depth()
		if line.State.Depth == 0 {
			continue
		}
		// done
		if line.Has(MarkDone) {
			line.State.State = line.State.State | RStateDone
		}
		// skip
		if line.Has(MarkSkip) {
			line.State.State = line.State.State | RStateSkip
		}
		// remmind
		if line.Has(MarkRemind) {
			line.State.State = line.State.State | RStateRemind
		}

		depth := line.State.Depth
		// 天上访问记录
		cal_assign_visit(depth, visits, line.Idx)
		// 路径, TODO 跳过中间的-1
		line.State.Path = append([]int{},
			visits[1:line.State.Depth+1]...)
		// 查找父亲
		parent := cal_parent(depth, visits, md.Lines)

		// 更新父亲的状态，并把父状态带给孩子
		if parent != nil {
			line.State.State = line.State.State |
				parent.State.State

			parent.State.Child++
			if line.State.State&RStateDone == 0 &&
				line.State.State&RStateSkip == 0 {
				parent.State.Todo++
			}
		}
		// 结束那些已经计算好的父节点
		cal_end_parent(depth, parent, parents, md.Lines, validIdx)

		validIdx = line.Idx
	}

	// 最后一行的结束计算 parent
	cal_end_parent(1, nil, parents, md.Lines, validIdx)

	// 最大的有效行
	md.MaxValidIdx = validIdx
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
func (daily *MdFlatDaily) PrintWithState() {
	for _, line := range daily.Lines {
		line.PrintWithState()
	}
	fmt.Print("\n")
}

// 打印
func (daily *MdFlatDaily) PrintTODO() {
	for i := 0; i < len(daily.Lines) && i <= daily.MaxValidIdx; {
		line := daily.Lines[i]
		if line.State.Child > 0 && line.State.Todo == 0 {
			i = line.State.MaxChildLine + 1
			continue
		}
		if line.State.State&RStateSkip == 0 &&
			line.State.State&RStateDone == 0 {
			line.Print()
		}
		i++
	}
}

// 打印
func (flat *MdFlatDaily) PrintHead(console *daily.ConsoleDaily) {
	console.Println(console.V.Split)
	console.PrintCenterln(console.ShortDailyPath(flat.File))
	console.Println("")
}

// 是否有未做的事情
func (daily *MdFlatDaily) HasTodo() (has bool) {
	has = false
	for i := 0; i < len(daily.Lines) && i <= daily.MaxValidIdx; {
		line := daily.Lines[i]
		if line.State.Child > 0 && line.State.Todo == 0 {
			i = line.State.MaxChildLine + 1
			continue
		}
		if line.State.State&RStateSkip == 0 &&
			line.State.State&RStateDone == 0 {
			has = true
			break
		}
		i++

	}
	return
}

// 执行Md 文件的相关操作
func Exec_MdDaily(input *daily.InputContext) {
	cmd := input.Args[0]
	if cmd == "ls" {
		Exec_MdDaily_ls(input)
	} else if cmd == "add" {
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
	Count    int
	Range    int
	Category string
}

// 列出所有未完成项目
func Exec_MdDaily_todo(input *daily.InputContext) {
	// 扫描出所有日志
	flats := md_ls(input)

	split := color.MagentaString(strings.Repeat("#", daily.LineLen))
	// 输出日志
	for _, daily := range flats {
		if daily.HasTodo() {
			input.Console.Println(split)
			input.Console.Println(color.MagentaString(input.Console.SprintfCenter(
				input.Console.ShortDailyPath(daily.File))))

			daily.PrintTODO()
		}
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

// 扫描本地的类别日志
func Exec_MdDaily_ls(input *daily.InputContext) {
	flats := md_ls(input)
	for _, v := range flats {
		v.PrintHead(input.Console)
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

// 把 slice 里面的所有值都赋值为value
func sliceAssign(a []int, value int) {
	for i, _ := range a {
		a[i] = value
	}
}

// 列出所有日志
func md_ls(input *daily.InputContext) (flats []*MdFlatDaily) {
	control := &MdFlag_todo{
		Ts:       daily.DefaultTime(),
		Range:    365,
		Count:    7,
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
	set.IntVar(&control.Count, "n", control.Count, "指定个数")
	set.IntVar(&control.Range, "r", control.Range, "指定范围(天数)")
	set.StringVar(&control.Category, "c", control.Category, "指定类别")

	set.Parse(vsets)

	var (
		sign int
		xnow time.Time = time.Now()
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

		if len(flats) >= control.Count {
			break
		}

		flatDaily := &MdFlatDaily{}
		if err := flatDaily.ReadFrom(dailyFile); err == nil {
			flats = append(flats, flatDaily)
		}
	}
	return
}
