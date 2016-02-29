package cmd

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	goterminal "github.com/apoorvam/goterminal"
	daily "jerry.com/everyday/daily"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

//***********************************
// record it
type Rd struct {
}

func (con *Rd) Name() string {
	return "rd"
}

func (con *Rd) Help() string {
	return "usage: rd cmd\n" +
		"rd ls [-c jerry] [-t 2016-1] [-n 5]\n" +
		"rd edit [record] \n" +
		"rd del record \n" +
		"the main daily rd(record) command"
}

func (con *Rd) Deal(input *daily.InputContext) {
	if len(input.Args) < 1 {
		input.Console.Println(con.Help())
		return
	}
	Exec_RdDaily(input)
}

func (con *Rd) AutoCompleteCallback(console *daily.ConsoleDaily, line []byte, pos, key int) (
	newLine []byte, newPos int) {
	newLine = nil
	return
}

//***********************************
type Note struct {
	// 属性
	Author   string
	Category string
	Ts       string

	// 路径
	FullPath string
	ModTime  time.Time

	// 窗口
	Console *daily.ConsoleDaily
}

// 从扫描信息获取日志
func MakeNote(console *daily.ConsoleDaily, fullpath string, fileInfo os.FileInfo) (note *Note, err error) {
	note = nil
	// 合法的note
	if fileInfo.IsDir() == true ||
		strings.HasPrefix(fileInfo.Name(), ".") == true ||
		strings.HasSuffix(fileInfo.Name(), ".md") == false {
		return
	}
	var (
		xcnow    time.Time
		category string
	)
	if category, xcnow, err = console.DailyDetail(fullpath); err != nil {
		return
	}

	note = &Note{}
	note.FullPath = fullpath
	note.Console = console
	note.Category = category
	note.Ts = xcnow.Format(daily.TimeDailyLayout)
	note.ModTime = fileInfo.ModTime()
	return
}

// 打印
func (note *Note) Print() {
	note.Console.Println(note.Console.V.Split)
	note.Console.PrintCenterln(note.Console.ShortDailyPath(note.FullPath))
	note.Console.Println("")
}

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

// 打开日志编辑
func (note *Note) Edit() {
	var (
		header, tail string
		shouldtail   bool = false
	)
	// begin edit
	if fileInfo, err := os.Stat(note.FullPath); err != nil || fileInfo.Size() == 0 {
		now := time.Now()
		xtime := now.Format(daily.TimeLayout)
		header = "------------------------------------------------------------\n" +
			xtime +
			strings.Repeat(" ", daily.LineLen-len(xtime)-len(note.Author)) +
			note.Author + "\n"
		DailyWrite(note.FullPath, header)

		shouldtail = true
	}
	daily.RunCommand("", "vi", "+", note.FullPath)

	if fileInfo, err := os.Stat(note.FullPath); err == nil {
		if fileInfo.Size() == int64(len(header)) {
			f, _ := os.OpenFile(note.FullPath, os.O_RDONLY, 0666)
			defer f.Close()
			r := bufio.NewReader(f)
			line0, _, _ := r.ReadLine()
			line1, _, _ := r.ReadLine()
			v := string(line0) + "\n" + string(line1) + "\n"
			if v == header {
				note.Del()
				shouldtail = false
			}
		}

		note.ModTime = fileInfo.ModTime()
	}
	if shouldtail {
		endnow := time.Now()
		endtime := endnow.Format(daily.TimeLayout)
		tail = endtime +
			strings.Repeat("-", daily.LineLen-len(endtime)-len(daily.EndMark)) +
			daily.EndMark + "\n\n"

		DailyWrite(note.FullPath, tail)
	}
}

// 打开日志查看
func (note *Note) Open() {
}

// 删除日志
func (note *Note) Del() error {
	return os.Remove(note.FullPath)
}

// 操作
func Exec_RdDaily(input *daily.InputContext) {
	cmd := input.Args[0]

	if cmd == "ls" {
		Exec_RdDaily_Ls(input)
	} else if cmd == "edit" {
		Exec_RdDaily_Edit(input)
	} else if cmd == "del" {
		Exec_RdDaily_Del(input)
	} else {
		input.Console.Println("Not Imp Cmd:", cmd)
	}
}

type RdFlag_Ls struct {
	Ts       string
	Count    int
	Category string
}

// 操作 - 枚举
func Exec_RdDaily_Ls(input *daily.InputContext) {
	notes := rd_ls(input, true)
	for _, v := range notes {
		v.Print()
	}
}

// 操作 - 编辑
func Exec_RdDaily_Edit(input *daily.InputContext) {
	rd_edit(input)
}

// 操作 - 删除
func Exec_RdDaily_Del(input *daily.InputContext) {
	var (
		err       error
		dailyFile string
	)
	split := strings.Repeat("#", daily.LineLen)
	if dailyFile, err = rd_del(input); err != nil {
		input.Console.Println(split)
		input.Console.Println("Remove Daily:",
			input.Console.ShortDailyPath(dailyFile),
			" Failed:", err)
	} else {
		input.Console.Println(split)
		input.Console.Println("Remove Daily:",
			input.Console.ShortDailyPath(dailyFile),
			" Success")
	}

	// 直接删除
	if len(input.Args) >= 1 {
		dailyFile := input.Console.FullDailyPath(input.Args[1])
		if err := os.Remove(dailyFile); err != nil {
			input.Console.Println(split)
			input.Console.Println("Remove Daily:",
				input.Console.ShortDailyPath(dailyFile),
				" Failed:", err)
		} else {
			input.Console.Println(split)
			input.Console.Println("Remove Daily:",
				input.Console.ShortDailyPath(dailyFile),
				" Success")
		}
	}
}

func rd_ls(input *daily.InputContext, dogress bool) (notes []*Note) {
	// 当前路径
	current := input.Console.Cmd_pwd()

	control := &RdFlag_Ls{
		Ts:       "", // time.Now().Format(daily.TimeLayout)
		Count:    -1,
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
	set := flag.NewFlagSet("ls", flag.ContinueOnError)
	set.StringVar(&control.Ts, "t", control.Ts, "指定日期")
	set.IntVar(&control.Count, "n", control.Count, "指定个数")
	set.StringVar(&control.Category, "c", control.Category, "指定类别")

	set.Parse(vsets)

	progress := goterminal.New(os.Stdout)
	notes = []*Note{}
	filepath.Walk(current, func(path string, info os.FileInfo, err error) error {
		if dogress {
			progress.Clear()
			fmt.Fprintf(progress, "扫描到(%d) 个符合要求的日志 ...\n", len(notes))
		}

		if err == nil {
			note, _ := MakeNote(input.Console, path, info)
			for {
				// 不是笔记
				if note == nil {
					break
				}
				// 个数已经够了
				if control.Count > 0 && len(notes) >= control.Count {
					return errors.New("已经够个数了")
				}
				// 时间不符合
				if control.Ts != "" &&
					strings.Contains(note.Ts, control.Ts) == false {
					break
				}
				// 类别不符合
				if control.Category != "" &&
					strings.Contains(control.Category, note.Category) == false {
					break
				}

				notes = append(notes, note)
				break
			}
		}
		if dogress {
			progress.Print()
		}
		if dogress && len(notes) < 10 {
			time.Sleep(time.Millisecond * 20)
		}
		return nil
	})
	return
}

func rd_edit(input *daily.InputContext) {
	control := &RdFlag_Ls{
		Ts:       daily.DefaultTime(),
		Count:    -1,
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
	set := flag.NewFlagSet("ls", flag.ContinueOnError)
	set.StringVar(&control.Ts, "t", control.Ts, "指定日期")
	set.IntVar(&control.Count, "n", control.Count, "指定个数")
	set.StringVar(&control.Category, "c", control.Category, "指定类别")

	set.Parse(vsets)

	if len(vargs) == 2 {
		current := input.Console.Cmd_pwd()
		dy := path.Join(current, vargs[1])
		if fileInfo, err := os.Stat(dy); err == nil &&
			fileInfo.IsDir() == false {
			note, _ := MakeNote(input.Console, dy, fileInfo)
			if note != nil {
				note.Edit()
				if daily.Config.Private != nil &&
					strings.Contains(*daily.Config.Private, note.Category) {
					daily.RunCommand("", "clear")
				}
			}

		} else {
			input.Console.Println("文件不存在",
				input.Console.ShortDailyPath(dy))
		}
	} else {
		_, _, dy := input.Console.DailyPath(control.Category, time.Now())
		fileInfo, err := os.Stat(dy)
		if err != nil && os.IsNotExist(err) {
			// make sure directory exits
			os.MkdirAll(path.Dir(dy), 0777)
			// touch file
			f, e := os.OpenFile(dy, os.O_CREATE, 0666)
			if e != nil {
				err = e
			} else {
				defer f.Close()
				fileInfo, err = os.Stat(dy)
			}
		}
		if err == nil {
			note, _ := MakeNote(input.Console, dy, fileInfo)
			if note != nil {
				note.Edit()
				if daily.Config.Private != nil &&
					strings.Contains(*daily.Config.Private, note.Category) {
					daily.RunCommand("", "clear")
				}
			}
		} else {
			input.Console.Println(err)
		}
	}
}

func rd_del(input *daily.InputContext) (dailyFile string, err error) {
	dailyFile = input.Console.FullDailyPath(input.Args[1])
	err = os.Remove(dailyFile)
	return
}
