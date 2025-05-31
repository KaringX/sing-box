// karing
package debug

import (
	"runtime"
	"strconv"
	"strings"
)

var MainGoroutineId int

func StackCurrent() string {
	buf := make([]byte, 2048)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}
	return string(buf)
}

func Stacks(all bool, includeStackBody bool) map[int]string {
	buf := make([]byte, 2048)
	for {
		n := runtime.Stack(buf, all)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}
	stacks := make(map[int]string)
	split := strings.Split(string(buf), "\n")

	head := ""
	body := ""
	for i := range split {
		if strings.HasPrefix(split[i], "goroutine ") {
			if len(head) != 0 {
				stk := strings.TrimPrefix(head, "goroutine")
				idField := strings.Fields(stk)[0]
				id, err := strconv.Atoi(idField)
				if err == nil {
					stacks[id] = body
				}
				head = ""
				body = ""
			}
			head = split[i]
		} else if len(head) != 0 {
			if includeStackBody {
				body += split[i] + "\n"
			}
		}
	}

	if len(head) != 0 {
		stk := strings.TrimPrefix(head, "goroutine")
		idField := strings.Fields(stk)[0]
		id, err := strconv.Atoi(idField)
		if err == nil {
			stacks[id] = body
		}
	}

	return stacks
}

func GetCurrentGoroutineId() int {
	stacks := Stacks(false, false)
	if len(stacks) > 0 {
		for key := range stacks {
			return key
		}
	}
	return 0
}

func GetGoroutineStack(id int) string {
	if id == 0 {
		return ""
	}
	stacks := Stacks(true, true)
	stackBody, ok := stacks[id]
	if ok {
		return stackBody
	}
	return ""
}
