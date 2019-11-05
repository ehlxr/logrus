package logrus

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

type TextFormatter struct {
	CallerPrettyfier       func(*runtime.Frame) (function string, file string)
	DisableLevelTruncation bool
}

func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var funcVal, fileVal string
	if entry.HasCaller() {
		funcVal = entry.Caller.Function
		fileVal = fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)

		if f.CallerPrettyfier != nil {
			funcVal, fileVal = f.CallerPrettyfier(entry.Caller)
		}
	}

	levelText := strings.ToUpper(entry.Level.String())
	if !f.DisableLevelTruncation {
		levelText = levelText[0:4]
	}

	return []byte(fmt.Sprintf("[%s][%s]%s %s %s\n",
		levelText,
		entry.Time.Format("2006-01-02 15:04:05"),
		fileVal,
		funcVal,
		entry.Message,
	)), nil
}
