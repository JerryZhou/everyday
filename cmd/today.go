package cmd

import (
	daily "jerry.com/everyday/daily"
	"os"
	"strings"
	"time"
)

//***********************************
type Today struct {
}

func (con *Today) Name() string {
	return "today"
}

func (con *Today) Help() string {
	return "usage: today\n" +
		"show me what should be do totay"
}

func (con *Today) Deal(input *daily.InputContext) {
	xnow := time.Now()
	input.Console.Println(input.Console.V.Split)
	input.Console.PrintCenterln(xnow.Format(daily.TimeLayout))

	categorys := listCategory(input.Console)
	mds := []*MdFlatDaily{}
	for _, v := range categorys {
		_, _, dy := input.Console.DailyPath(v, xnow)
		if _, e := os.Stat(dy); e == nil {
			md := &MdFlatDaily{}
			if e0 := md.ReadFrom(dy); e0 == nil {
				mds = append(mds, md)
			}
		}
	}
	for _, md := range mds {
		md.PrintTODO()
	}
}

func (con *Today) AutoCompleteCallback(console *daily.ConsoleDaily, line []byte, pos, key int) (
	newLine []byte, newPos int) {
	newLine = nil
	return
}

//***********************************
// 最近一年的 category
func listCategory(console *daily.ConsoleDaily) (categorys []string) {
	xnow := time.Now()
	xcategory := map[string]string{}
	for i := 0; i < 12; i++ {
		t := xnow.AddDate(0, -1*i, 0)
		dailyPath, _, _ := console.DailyPath("", t)
		if f, e := os.OpenFile(dailyPath, os.O_RDONLY, 0666); e == nil {
			if names, e0 := f.Readdirnames(-1); e0 == nil {
				for _, v := range names {
					if strings.HasPrefix(v, ".") {
						continue
					}
					xcategory[v] = v
				}
			}
		}
	}
	for k, _ := range xcategory {
		categorys = append(categorys, k)
	}
	return
}
