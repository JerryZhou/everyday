package daily

import (
	"io/ioutil"
	"os/user"
	"path"
	"time"
)

// 全局配置项
var (
	Config *DailyConfig
)

// 配置项
type DailyConfig struct {
	HomeDir *string `json:"homedir,omitempty"`
	Author  *string `json:"author,omitempty"`

	DefaultCategory *string `json:"category,omitempty"`

	Private *string `json:"private,omitempty"`
}

// 设置值
func (config *DailyConfig) Set(key, value string) (setted bool) {
	setted = true
	xvalue := value
	if key == "home" {
		config.HomeDir = &xvalue
	} else if key == "author" {
		config.Author = &xvalue
	} else if key == "category" {
		config.DefaultCategory = &xvalue
	} else if key == "private" {
		config.Private = &xvalue
	} else {
		setted = false
	}
	if setted {
		ioutil.WriteFile(ConfigRc(), []byte(GiveMeJson(config)), 0666)
	}
	return
}

// 配置文件位置
func ConfigRc() string {
	// 解析命令行
	current, _ := user.Current()
	rc := path.Join(current.HomeDir, EveryDay, ".rc")

	return rc
}

// 读取配置项
func ReadConfig() *DailyConfig {
	config := &DailyConfig{}

	// 解析命令行
	current, _ := user.Current()
	rc := path.Join(current.HomeDir, EveryDay, ".rc")

	// 读取配置文件
	ReadJson(config, rc)
	return config
}

// 默认的日期
func DefaultCategory() (category string) {
	if Config.DefaultCategory != nil {
		category = *Config.DefaultCategory
	}
	if category == "" {
		category = "everyday"
	}
	return
}

// 默认的时间
func DefaultTime() (ts string) {
	ts = time.Now().Format(TimeDailyLayout)
	return
}

func PrepareConfig() {
	Config = ReadConfig()
}

// 控制项
type FlagControl struct {
	FlagNewDaily      bool
	FlagTimeDaily     string // 指定时间
	FlagCategoryDaily string // 指定类别
	FlagLenDaily      int    // 个数

	FlagKeywordDaily string // 关键字
	FlagDirDaily     string // 目录

	FlagRangeDaily int // 范围
}

// 当前用户上下文
type DailyContext struct {
	Daily string // 当前日志文件

	XNow time.Time // 配置的起始时间
}

/*
// 准备上下文
func (context *DailyContext) Prepare(Controls *FlagControl) error {
	// 根据日期定为一个具体文件
	context.XNow = time.Now()
	if len(Controls.FlagTimeDaily) > 0 {
		if xnow, err := time.Parse(TimeDailyLayout, Controls.FlagTimeDaily); err == nil {
			context.XNow = xnow
		}
	}

	var (
		dailyPath string
	)
	dailyPath, _, context.Daily = DailyPath(input, context.XNow)
	// 确保父目录存在
	if err := os.MkdirAll(dailyPath, 0777); err != nil {
		return err
	}

	return nil
}

*/
