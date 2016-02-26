package cmd

import (
	"fmt"
	//"errors"
	heap "container/heap"
	gosxnotifier "github.com/deckarep/gosx-notifier"
	//color "github.com/fatih/color"
	daily "jerry.com/everyday/daily"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 提醒队列的操作参数
const (
	SeqOpCodeAdd = iota
	SeqOpCodeDel
	SeqOpCodeLs
	SeqOpCodeExit
)

// 操作参数
type SeqOp struct {
	Op int
	R  *Reminder
}

// Heap
// 提醒队列
type ReminderSequence struct {
	// 按照提醒时间排序的队列
	Sequence []*Reminder

	Duration time.Duration
	Ops      chan *SeqOp

	IsExit bool

	// 提醒队列
	Notes chan *Reminder
}

// 打开
func (seq *ReminderSequence) Open() {
	seq.Sequence = []*Reminder{}
	seq.Ops = make(chan *SeqOp, 20)
	seq.Duration = 0
	seq.IsExit = false

	seq.Notes = make(chan *Reminder, 100)
}

// 开启loop
func (seq *ReminderSequence) Start() {
	go seq.StartLoop()
	go seq.StartNoteLoop()
}

//{{ for heap
func (seq *ReminderSequence) Len() int { return len(seq.Sequence) }
func (seq *ReminderSequence) Less(i, j int) bool {
	return seq.Sequence[i].BoltTs() < seq.Sequence[j].BoltTs()
}
func (seq *ReminderSequence) Swap(i, j int) {
	seq.Sequence[i], seq.Sequence[j] = seq.Sequence[j], seq.Sequence[i]

	seq.Sequence[i].Seat = i
	seq.Sequence[j].Seat = j
}
func (seq *ReminderSequence) Push(x interface{}) {
	r := x.(*Reminder)
	r.Seat = seq.Len()
	r.Seq = seq

	seq.Sequence = append(seq.Sequence, r)
}
func (seq *ReminderSequence) Pop() interface{} {
	n := seq.Len()
	x := seq.Sequence[n-1]
	x.Seat = 0 // 移除了
	x.Seq = nil

	seq.Sequence = seq.Sequence[:n-1]
	return x
}

// }} end for heap

// 新增一个序列
func (seq *ReminderSequence) addop(r *Reminder) (success bool) {
	// 已经跳过的，已经完成的就不加入通知序列了
	if r.State&RStateSkip != 0 || r.State&RStateDone != 0 {
		success = false
		return
	}
	// 已经加过一次了
	if r.Seq != nil {
		success = false
		return
	}
	success = true
	heap.Push(seq, r)
	return
}

// 从seq 里面移除一个
func (seq *ReminderSequence) delop(r *Reminder) (success bool) {
	if r.Seq != seq {
		success = false
		return
	}
	success = true
	heap.Remove(seq, r.Seat)
	return
}

// 最近需要提醒的一个
func (seq *ReminderSequence) peekop() (r *Reminder) {
	if seq.Len() > 0 {
		r = seq.Sequence[0]
	}
	return
}

// 退出循环
func (seq *ReminderSequence) exitop() {
	seq.IsExit = true
}

// 可排序的列表
type RdArray struct {
	Rds []*Reminder
}

func (rds *RdArray) Len() int           { return len(rds.Rds) }
func (rds *RdArray) Less(i, j int) bool { return rds.Rds[i].BoltTs() < rds.Rds[j].BoltTs() }
func (rds *RdArray) Swap(i, j int)      { rds.Rds[i], rds.Rds[j] = rds.Rds[j], rds.Rds[i] }

// 按照实际顺序列出来
func (seq *ReminderSequence) lsop() {
	rds := &RdArray{}
	rds.Rds = append([]*Reminder{}, seq.Sequence...) // copy
	go func() {
		sort.Sort(rds)
		for _, r := range rds.Rds {
			r.Print()
		}
	}()
}

// 退出
func (seq *ReminderSequence) Exit() {
	op := &SeqOp{Op: SeqOpCodeExit}
	seq.Ops <- op
}

// 加入
func (seq *ReminderSequence) Add(r *Reminder) {
	op := &SeqOp{Op: SeqOpCodeAdd, R: r}
	seq.Ops <- op
}

// 删除
func (seq *ReminderSequence) Del(r *Reminder) {
	op := &SeqOp{Op: SeqOpCodeDel, R: r}
	seq.Ops <- op
}

// 穷举
func (seq *ReminderSequence) Ls() {
	op := &SeqOp{Op: SeqOpCodeLs}
	seq.Ops <- op
}

// 时间到了
func (seq *ReminderSequence) when_reach(r *Reminder) {
	seq.Notes <- r
}

// 决定谁是下一个
func (seq *ReminderSequence) which_next(r *Reminder) {
	seq.Duration = 0
	now := time.Now()
	// 外部不知道
	for {
		r = seq.peekop()
		if r == nil {
			// 已经没有了
			break
		}
		d := r.Time().Sub(now)
		if d <= 0 {
			// 时间到了移除
			seq.delop(r)
			seq.when_reach(r)
		} else {
			// 下一个就是r
			seq.Duration = d
			break
		}
	}
}

// 操作到了
func (seq *ReminderSequence) when_op(op *SeqOp) {
	success := false
	switch op.Op {
	case SeqOpCodeAdd:
		success = seq.addop(op.R)
	case SeqOpCodeDel:
		success = seq.delop(op.R)
	case SeqOpCodeExit:
		seq.exitop()
	case SeqOpCodeLs:
		success = true
		seq.lsop()
	default:
	}

	// 下一个reminder
	if success {
		seq.which_next(nil)
	}
}

// 时间到了定时
func (seq *ReminderSequence) when_tick(t time.Time) {

	for {
		r := seq.peekop()
		if r == nil {
			// 已经没有了
			break
		}
		if t.After(r.Time()) {
			seq.delop(r)
			seq.when_reach(r)
		} else {
			// 决定下一个
			seq.which_next(r)
			break
		}
	}
}

// 开启服务
func (seq *ReminderSequence) StartLoop() {
	for {
		if seq.Duration > 0 {
			select {
			case op := <-seq.Ops:
				seq.when_op(op)
			case t := <-time.After(seq.Duration):
				seq.when_tick(t)
			}
		} else {
			op := <-seq.Ops
			seq.when_op(op)
		}

		if seq.IsExit {
			break
		}
	}
}

// 无线循环，处理一个休息一下
func (seq *ReminderSequence) StartNoteLoop() {
	for {
		r := <-seq.Notes
		if r == nil {
			break
		}
		seq.exec_note(r)
		if seq.IsExit {
			break
		}
		time.Sleep(time.Second)
	}
}

// 执行具体的通知
func (seq *ReminderSequence) exec_note(r *Reminder) {
	// TODO: to real op channel
	// TODO: reached r
	//fmt.Println(color.GreenString("\n提醒到点了------"))
	//fmt.Println(color.RedString("现在时间:" + time.Now().Format(daily.TimeLayout)))
	//fmt.Println(color.RedString("提醒时间:" + r.Time().Format(daily.TimeLayout)))
	//r.Line.Print()

	//At a minimum specifiy a message to display to end-user.
	note := gosxnotifier.NewNotification(r.Line.Line)

	//Optionally, set a title
	note.Title = r.Key.Category

	//Optionally, set a subtitle
	note.Subtitle = r.Key.Name

	//Optionally, set a sound from a predefined set.
	note.Sound = gosxnotifier.Basso

	//Optionally, set a group which ensures only one notification is ever shown replacing previous notification of same group id.
	note.Group = "com.unique.yourapp.identifier"

	//Optionally, set a sender (Notification will now use the Safari icon)
	note.Sender = "com.apple.Safari"

	//Optionally, specifiy a url or bundleid to open should the notification be
	//clicked.
	note.Link = "http://www.google.com" //or BundleID like: com.apple.Terminal

	//Optionally, an app icon (10.9+ ONLY)
	// xnow.String()
	// /Users/jerryzhou/.everyday/.config
	note.AppIcon = "/Users/jerryzhou/.everyday/.config/icon.png"

	//Optionally, a content image (10.9+ ONLY)
	note.ContentImage = "/Users/jerryzhou/.everyday/.config/icon.png"

	//Then, push the notification
	err := note.Push()

	//If necessary, check error
	if err != nil {
		fmt.Println("Uh oh!")
	}
}

// 提醒的唯一key
type ReminderKey struct {
	Category string
	Name     string
}

// 文本
func (key *ReminderKey) String() string {
	//return daily.GiveMeJson(key)
	return key.Category + "-" + key.Name
}

// *[REMIND]*__("ts":"", "way":"", "count":"")__
// 匹配提醒的正则表达式
var (
	RemindRegxp *regexp.Regexp = regexp.MustCompile("\\*\\[REMIND\\]\\* __\\([^\\(]*\\)__")
)

// 提醒控制参数
type ReminderParam struct {
	Ts    *string `json:"ts,omitempty"`
	Way   *string `json:"way,omitempty"`
	Count *int    `json:"count,omitempty"`
}

// 提醒的状态
// 所有的reminder 状态
const (
	RStateDone = 1 << iota
	RStateSkip
	RStateRemind
)

// 一个提醒事项
type Reminder struct {
	Key   *ReminderKey // 关键字
	Line  *MdFlatLine
	Param *ReminderParam // 参数

	Ts []time.Time // 提醒时间

	State int // 状态

	Daily *ReminderDaily // 来自哪里

	// 由 sequence 进行维护的数据，外部不能修改
	Seat int               // 在堆栈的位置
	Seq  *ReminderSequence // 提醒序列
}

// 下一个提醒时间点
func (r *Reminder) BoltTs() int64 {
	return r.Time().Unix()
}

// 需要的提醒时间
func (r *Reminder) Time() time.Time {
	if len(r.Ts) > 0 {
		return r.Ts[len(r.Ts)-1]
	}
	if r.Param != nil && r.Param.Ts != nil {
		if t, e := time.ParseInLocation(daily.TimeLayout, *r.Param.Ts, time.Local); e == nil {
			r.Ts = append(r.Ts, t)
			return t
		} else if t, e := time.ParseInLocation(daily.TimeLayout_Seconds, *r.Param.Ts, time.Local); e == nil {
			r.Ts = append(r.Ts, t)
			return t
		} else {
			fmt.Println("时间格式错误", *r.Param.Ts)
			r.Param.Ts = nil
		}
	}
	return time.Now()
}

// 打印自己
func (r *Reminder) Print() {
	fmt.Println("")
	fmt.Println(strings.Repeat("#", 60))
	fmt.Println("Key:", r.Key.String())
	fmt.Print("Content:")
	r.Line.Print()
	fmt.Println("Time:", r.Time().Format(daily.TimeLayout))
}

// 一篇日志里面的所有提醒项目
type ReminderDaily struct {
	File      string      // 日志路径
	Reminders []*Reminder // 提醒项目
}

// 创建一个提醒
// *[REMIND]*("ts:":"", "way":"")
//\\*\\[REMIND\\]\\*\\([^\\(]*\\)
func (rd *ReminderDaily) MarkRemind(line *MdFlatLine) (r *Reminder) {
	r = &Reminder{}
	r.Key = &ReminderKey{}
	r.Key.Category = rd.File
	r.Key.Name = strings.Join(sliceItoa(line.State.Path), ",")
	r.Line = line
	r.Daily = rd
	r.State = line.State.State
	rdparam := RemindRegxp.FindString(line.Line)
	if len(rdparam) > 14 {
		// NB!! 11 is hard code for *[REMIND]* __( param )__
		r.Param = &ReminderParam{}
		if err := daily.GiveMeJsonObject(
			string(rdparam[14:len(rdparam)-3]), r.Param); err != nil {
			//fmt.Println("参数填写错误", err)
			r = nil
			return
		}
		r.Ts = make([]time.Time, 0)
	}
	return
}

// 从平板文件读取
func (rd *ReminderDaily) ReadFrom(md *MdFlatDaily, skip bool) (err error) {
	rd.File = md.File
	for _, line := range md.Lines {
		if line.State == nil || line.State.Depth == 0 {
			continue
		}
		if skip {
			if line.State.State&RStateSkip != 0 ||
				line.State.State&RStateDone != 0 {
				continue
			}
		}
		if line.State.State&RStateRemind == 0 {
			continue
		}

		if r := rd.MarkRemind(line); r != nil {
			rd.Reminders = append(rd.Reminders, r)
		}
	}

	return
}

//***********************************
type Remind struct {
	// 所有的提醒事项
	Reminders map[string]*Reminder
	// 所有日志项目
	Dailys map[string]*ReminderDaily
	// 提醒序列
	Sequence *ReminderSequence

	// 控制台
	Console *daily.ConsoleDaily

	HaveOpen bool
}

// 初始化
func (con *Remind) Open() {
	con.HaveOpen = true
	con.Sequence = &ReminderSequence{}
	con.Sequence.Open()
	con.Sequence.Start()

	con.Dailys = map[string]*ReminderDaily{}
	con.Reminders = map[string]*Reminder{}
}

// 扫描日志加入提醒系统
func (con *Remind) AddDaily(daily string) (rd *ReminderDaily) {
	flat := &MdFlatDaily{}
	if err := flat.ReadFrom(daily); err != nil {
		con.Console.Println(err)
		return
	}

	xrd := &ReminderDaily{}
	if err := xrd.ReadFrom(flat, false); err == nil {
		con.Attach(xrd)
		rd = xrd
	}
	return
}

// 附加一个提醒日志到系统
func (con *Remind) Attach(rd *ReminderDaily) {
	// 已经存在则移除以前的
	if d, ok := con.Dailys[rd.File]; ok {
		con.Dettach(d)
	}
	// 加入系统
	con.Dailys[rd.File] = rd
	for _, r := range rd.Reminders {
		con.Add(r)
	}
}

// 从系统移除一个日志
func (con *Remind) Dettach(rd *ReminderDaily) {
	delete(con.Dailys, rd.File)
	for _, r := range rd.Reminders {
		con.Del(r)
	}
}

// 查找
func (con *Remind) Find(key string) (r *Reminder) {
	r, _ = con.Reminders[key]
	return
}

// 增加
func (con *Remind) Add(r *Reminder) {
	key := r.Key.String()
	if r0 := con.Find(key); r0 != nil {
		con.Del(r0)
	}
	con.Reminders[key] = r
	con.Sequence.Add(r)
}

// 删除
func (con *Remind) Del(r *Reminder) {
	delete(con.Reminders, r.Key.String())
	con.Sequence.Del(r)
}

func (con *Remind) Name() string {
	return "remind"
}

func (con *Remind) Help() string {
	return "usage: remind\n" +
		"clear current screen"
}

func (con *Remind) Deal(input *daily.InputContext) {
	if !con.HaveOpen {
		con.Open()
	}
	if len(input.Args) < 1 {
		input.Console.Println(con.Help())
		return
	}
	con.Console = input.Console
	cmd := input.Args[0]
	if cmd == "see" {
		con.exec_remind_see(input)
	} else if cmd == "try" {
		con.exec_remind_try(input)
	} else if cmd == "ls" {
		con.exec_remind_ls(input)
	} else if cmd == "del" {
		con.exec_remind_del(input)
	} else if cmd == "add" {
		con.exec_remind_add(input)
	} else if cmd == "collect" {
		con.exec_remind_collect(input)
	}
}

func (con *Remind) AutoCompleteCallback(console *daily.ConsoleDaily, line []byte, pos, key int) (
	newLine []byte, newPos int) {
	newLine = nil
	return
}

// 把指定的文档加入可见系统, 如果重复加会移除以前的
func (con *Remind) exec_remind_see(input *daily.InputContext) {
	if len(input.Args) < 2 {
		input.Console.Println(con.Help())
		return
	}
	for _, dy := range input.Args[1:] {
		con.Console.Println(con.Console.V.Split)
		con.Console.PrintCenterln(con.Console.ShortDailyPath(dy))

		rd := con.remind_see(input.Console, dy)
		if rd != nil && len(rd.Reminders) > 0 {
			for _, r := range rd.Reminders {
				r.Line.Print()
			}
		} else {
			con.Console.PrintCenterln("没有提醒")
			con.Console.Println("")
		}
	}
}

// 列出当前的提醒队列
func (con *Remind) exec_remind_ls(input *daily.InputContext) {
	con.Sequence.Ls()
}

// 加一个提醒进去
// add key content ts
func (con *Remind) exec_remind_add(input *daily.InputContext) {
	if len(input.Args) < 4 {
		input.Console.Println(con.Help())
		return
	}
	key := input.Args[1]
	line := input.Args[2]
	ts := input.Args[3]

	r := &Reminder{}
	r.Param = &ReminderParam{}
	r.Param.Ts = &ts
	r.Key = &ReminderKey{}
	r.Key.Category = "jerry"
	r.Key.Name = key
	r.Line = &MdFlatLine{}
	r.Line.Idx = 0
	r.Line.Line = line
	r.Daily = nil
	r.State = 0
	r.Ts = make([]time.Time, 0)

	con.Add(r)
}

// 删除一个提醒
func (con *Remind) exec_remind_del(input *daily.InputContext) {
	if len(input.Args) < 2 {
		input.Console.Println(con.Help())
		return
	}
	key := input.Args[1]
	if r := con.Find(key); r != nil {
		con.Del(r)
	} else {
		fmt.Println("没有找到提醒:", key)
	}
}

// 在多久以后通知一下我
func (con *Remind) exec_remind_try(input *daily.InputContext) {
	if len(input.Args) < 2 {
		input.Console.Println(con.Help())
		return
	}
	i, _ := strconv.Atoi(input.Args[1])
	xnow := time.Now().Add(time.Duration(int64(i) * int64(time.Second)))

	ts := xnow.Format(daily.TimeLayout) // local format
	r := &Reminder{}
	r.Param = &ReminderParam{}
	r.Param.Ts = &ts
	r.Key = &ReminderKey{}
	r.Key.Category = "try"
	r.Key.Name = ts
	r.Line = &MdFlatLine{}
	r.Line.Idx = 0
	r.Line.Line = "* try tick " + r.Key.Name
	r.Daily = nil
	r.State = 0
	r.Ts = make([]time.Time, 0)

	con.Add(r)
}

// 收集提醒事项
func (con *Remind) exec_remind_collect(input *daily.InputContext) {
	input.Args[0] = "ls"
	notes := rd_ls(input, false)
	for _, v := range notes {
		dy := v.FullPath
		rd := con.remind_see(input.Console, dy)
		if rd != nil && len(rd.Reminders) > 0 {
			con.Console.Println(con.Console.V.Split)
			con.Console.PrintCenterln(con.Console.ShortDailyPath(dy))

			for _, r := range rd.Reminders {
				r.Line.Print()
			}
		}
	}
}

// slice I to A
func sliceItoa(sliceint []int) (slice []string) {
	for _, i := range sliceint {
		slice = append(slice, strconv.Itoa(i))
	}
	return
}

func (con *Remind) remind_see(console *daily.ConsoleDaily, dy string) (rd *ReminderDaily) {
	rd = con.AddDaily(console.FullDailyPath(dy))
	return
}
