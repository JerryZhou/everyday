package daily

import (
	"encoding/json"
	"errors"
	"fmt"
	rw "github.com/mattn/go-runewidth"
	"os"
	"os/exec"
)

//***********************************
// 搞一个Json 串出来 Marshal Json
func GiveMeJson(req interface{}) string {
	if req == nil {
		return "nil"
	}

	if jsonret, err := json.Marshal(req); err == nil {
		return string(jsonret)
	}

	return "error"
}

//***********************************
// 搞一个对象给我， UnMarshal Json
func GiveMeJsonObject(jsonstr string, req interface{}) error {
	if req == nil {
		return errors.New("无能序列化到nil")
	}
	if jsonstr == "" {
		return errors.New("空的json串")
	}

	// wrong format
	if err := json.Unmarshal([]byte(jsonstr), req); err != nil {
		return err
	}

	return nil
}

//***********************************
// 文本宽度
func StringWidth(s string) (width int) {
	width = rw.StringWidth(s)
	return
}

//***********************************
// 从文件里面读取一个结构体
func ReadJson(obj interface{}, jsonpath string) error {
	config, err := os.Open(jsonpath)
	if err != nil {
		fmt.Println("file open err:", err)
		return err
	}
	defer config.Close()
	jsonparser := json.NewDecoder(config)
	if err = jsonparser.Decode(obj); err != nil {
		fmt.Println("json decode err:", err)
		return err
	}
	return nil
}

//***********************************
// Returns the executable path and arguments
func ParseEditorCommand(editorCmd string) (
	string, []string, error) {
	var args []string
	state := "start"
	current := ""
	quote := "\""
	for _, c := range editorCmd {
		if state == "quotes" {
			if string(c) != quote {
				current += string(c)
			} else {
				args = append(args, current)
				current = ""
				state = "start"
			}
			continue
		}

		if c == '"' || c == '\'' {
			state = "quotes"
			quote = string(c)
			continue
		}

		if state == "arg" {
			if c == ' ' || c == '\t' {
				args = append(args, current)
				current = ""
				state = "start"
			} else {
				current += string(c)
			}
			continue
		}

		if c != ' ' && c != '\t' {
			state = "arg"
			current += string(c)
		}
	}

	if state == "quotes" {
		return "", []string{}, errors.New(fmt.Sprintf("Unclosed quote in command line: %s", editorCmd))
	}

	if current != "" {
		args = append(args, current)
	}

	if len(args) <= 0 {
		return "", []string{}, errors.New("Empty command line")
	}

	return args[0], args, nil
}

//***********************************
// 解析 flag set 并且返回 flag set 之外的参数
func ParseFlagSet(args []string) (vargs []string, vsets []string) {
	// parse cmd and flags
	argx := 0
	argy := 0
	for i, v := range args {
		if len(v) > 0 && v[0] == '-' {
			vsets = args[i:]
			break
		} else {
			if argy == 0 {
				argx = i
			}
			argy = i + 1
		}
	}

	vargs = args[argx:argy]
	return
}

//***********************************
// 执行命令
func RunCommand(ch string, name string, arg ...string) (cmdOut []byte, err error) {
	for {
		if len(ch) > 0 {
			// 跳转目录
			if err = os.Chdir(ch); err != nil {
				fmt.Println("os ch dir err:", err)
				break
			}
		}

		cmd := exec.Command(name, arg...)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		if len(ch) > 0 {
			cmdOut, err = cmd.Output()
		} else {
			cmd.Stdout = os.Stdout
			err = cmd.Run()
		}

		if err != nil {
			if name == "grep" {
				if err.Error() == "exit status 1" {
					fmt.Println("没有找到任何匹配的内容")
				} else {
				}
			} else {
				fmt.Println(err, name, arg)
			}
			break
		}
		if len(cmdOut) > 0 {
			fmt.Println(string(cmdOut))
		}
		break
	}
	return
}
