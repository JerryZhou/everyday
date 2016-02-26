package cmd

import (
	"fmt"
	//"errors"
	heap "container/heap"
	gosxnotifier "github.com/deckarep/gosx-notifier"
	//color "github.com/fatih/color"
	daily "jerry.com/everyday/daily"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 提醒队列的操作参数
const (
	SeqOpCodeAdd = iota
	SeqOpCodeDel
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
}

// 打开
func (seq *ReminderSequence) Open() {
	seq.Sequence = []*Reminder{}
	seq.Ops = make(chan *SeqOp)
	seq.Duration = 0
	seq.IsExit = false
}

// 开启loop
func (seq *ReminderSequence) Start() {
	go seq.StartLoop()
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

// 时间到了
func (seq *ReminderSequence) when_reach(r *Reminder) {
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

// 提醒的唯一key
type ReminderKey struct {
	Category string
	Name     string
}

// 文本
func (key *ReminderKey) String() string {
	return daily.GiveMeJson(key)
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

// 一篇日志里面的所有提醒项目
type ReminderDaily struct {
	File      string      // 日志路径
	Reminders []*Reminder // 提醒项目
}

// 从平板文件读取
func (rd *ReminderDaily) ReadFrom(md *MdFlatDaily, skip bool) (err error) {
	rd.File = md.File

	visit := make([]int, 32) // 父级路径
	stats := make([]int, 32) // 级别状态
	dh := 0
	for _, line := range md.Lines {
		/*
			if line.State == nil || line.State.Depth == 0 {
				continue
			}
		*/
		if dh = line.Depth(); dh == 0 {
			continue
		}
		stats[dh] = stats[dh-1]
		visit[dh] = line.Idx

		// 父亲已经done了
		if skip && (stats[dh]&RStateDone) != 0 {
			continue
		}
		// 父亲已经skip了
		if skip && (stats[dh]&RStateSkip) != 0 {
			continue
		}
		// 自己已经done
		if line.Has(MarkDone) {
			stats[dh] = stats[dh] | RStateDone
			if skip {
				continue
			}
		}
		// 自己已经skip
		if line.Has(MarkSkip) {
			stats[dh] = stats[dh] | RStateSkip
			if skip {
				continue
			}
		}
		// 没有 reminder
		if !line.Has(MarkRemind) {
			continue
		}

		// *[REMIND]*("ts:":"", "way":"")
		//\\*\\[REMIND\\]\\*\\([^\\(]*\\)
		r := &Reminder{}
		r.Key = &ReminderKey{}
		r.Key.Category = rd.File
		r.Key.Name = strings.Join(sliceItoa(visit[1:dh+1]), ",")
		r.Line = line
		r.Daily = rd
		r.State = stats[dh]
		rdparam := RemindRegxp.FindString(line.Line)
		if len(rdparam) > 14 {
			// NB!! 11 is hard code for *[REMIND]* __( param )__
			r.Param = &ReminderParam{}
			if err := daily.GiveMeJsonObject(
				string(rdparam[14:len(rdparam)-3]), r.Param); err != nil {
				//fmt.Println("参数填写错误", err)
				continue
			}
			r.Ts = make([]time.Time, 0)
		}
		// TODO Ts
		rd.Reminders = append(rd.Reminders, r)
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
	if d, ok := con.Dailys[daily]; ok {
		// con.Console.Println("已经存在了")
		rd = d
		return
	}
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
	con.Dailys[rd.File] = rd
	for _, r := range rd.Reminders {
		con.Reminders[r.Key.String()] = r
		con.Sequence.Add(r)
	}
}

// 从系统移除一个日志
func (con *Remind) Dettach(rd *ReminderDaily) {
	delete(con.Dailys, rd.File)
	for _, r := range rd.Reminders {
		delete(con.Reminders, r.Key.String())
		con.Sequence.Del(r)
	}
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
	if len(input.Args) < 2 {
		input.Console.Println(con.Help())
		return
	}
	con.Console = input.Console
	if input.Args[0] == "see" {
		for _, dy := range input.Args[1:] {
			con.Console.Println(con.Console.V.Split)
			con.Console.PrintCenterln(con.Console.ShortDailyPath(dy))

			rd := con.AddDaily(input.Console.FullDailyPath(dy))
			if rd != nil && len(rd.Reminders) > 0 {
				for _, r := range rd.Reminders {
					r.Line.Print()
				}
			} else {
				con.Console.PrintCenterln("没有提醒")
				con.Console.Println("")
			}
		}
	} else if input.Args[0] == "try" {
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
		con.Sequence.Add(r)
	}
}

func (con *Remind) AutoCompleteCallback(console *daily.ConsoleDaily, line []byte, pos, key int) (
	newLine []byte, newPos int) {
	newLine = nil
	return
}

// slice I to A
func sliceItoa(sliceint []int) (slice []string) {
	for _, i := range sliceint {
		slice = append(slice, strconv.Itoa(i))
	}
	return

}
