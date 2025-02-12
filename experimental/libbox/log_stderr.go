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
			content := fmt.Sprintf("%v\n%s", e, string(debug.Stack()))
			err = E.Cause(E.New(content), "panic: StderrRedirect")
			SentryCaptureException(&SentryPanicError{Err: err.Error()})
		}
	}()
	if len(path) == 0 {
		return nil
	}
	stderrLogFile, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return E.Cause(err, "StderrRedirect:")
	}
	content, _ := readFile(stderrLogFile, 4*1000)
	if len(content) > 0 {
		go func() {
			var stack []string
			index := strings.Index(content, "panic")
			if index >= 0 {
				lines := strings.Split(content[index:], "\n")
				findStack := false
				for _, line := range lines {
					line = strings.Trim(line, "\r\t\n")
					if strings.HasPrefix(line, "goroutine ") {
						if findStack {
							break
						}
						findStack = true
					} else{
						stack = append(stack, line)
					}
				}
				if len(stack) > 0 {
					SentryCaptureException(&SentryPanicError{Err: strings.Join(stack,"\n")})
				}
			}
		}()
	}
	stderrLogFile.Truncate(0)
	stderrLogFile.Seek(0, 0)
	stderrLogFile.Sync()
	return stderrRedirect(stderrLogFile)
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
