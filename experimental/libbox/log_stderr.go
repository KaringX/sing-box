package libbox

//karing

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	E "github.com/sagernet/sing/common/exceptions"
)

var stderrLogFile *os.File

func StderrRedirect(path string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			panicErrMessage := fmt.Sprintf("%v", e)
			SentryCaptureErrorMessage(panicErrMessage, "panic: StderrRedirect", SentryTrim(string(debug.Stack())))
		}
	}()
	if len(path) == 0 {
		return nil
	}
	stderrLogFile, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return E.Cause(err, "StderrRedirect:")
	}
	StderrCheckAndCapture()
	return stderrRedirect(stderrLogFile)
}

func StderrCheckAndCapture() {
	if stderrLogFile == nil {
		return
	}
	content, _ := readFile(stderrLogFile, 4*1000)
	stderrLogFile.Truncate(0)
	stderrLogFile.Seek(0, 0)
	stderrLogFile.Sync()
	if len(content) > 0 {
		go func() {
			var stack []string
			index := strings.Index(content, "panic")
			if index >= 0 {
				lines := strings.Split(content[index:], "\n")
				panicErrMessage := ""
				findStack := false
				for i, line := range lines {
					line = strings.Trim(line, "\r\t\n")
					if i == 0 {
						panicErrMessage += line
						continue
					}
					if strings.HasPrefix(line, "goroutine ") {
						if findStack {
							break
						}
						findStack = true
					} else {
						stack = append(stack, line)
					}
				}
				if len(stack) > 0 {
					SentryCaptureErrorMessage(panicErrMessage, "panic: stderrLog", strings.Join(stack, "\n"))
				}
			}
		}()
	}
}
func StderrWrite(content string) {
	if stderrLogFile == nil {
		return
	}
	stderrLogFile.WriteString(content)
	stderrLogFile.Sync()
}

func readFile(file *os.File, maxLen int64) (string, error) {
	var step int64 = 1000
	var offset int64 = 0
	buf := make([]byte, step)
	result := ""
	for offset <= maxLen {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return result, err
		}
		if err == io.EOF {
			break
		}
		result += string(buf[:n])
	}
	return result, nil
}
