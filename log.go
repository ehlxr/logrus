package logrus

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/ehlxr/logrus/crash"

	"github.com/rifflock/lfshook"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
)

var (
	crashLog           = flag.String("log-crash", "./crash.log", "The crash log file.")
	logFile            = flag.String("log-file", "", "The external log file. Default log to console.")
	logLevel           = flag.String("log-level", "info", "The log level, default is info")
	logLineNumber      = flag.Bool("log-ln", true, "Print log line number, default is true")
	logTimestamp       = flag.Bool("log-ts", true, "Print log timestamp, default is true")
	logColors          = flag.Bool("log-cl", true, "Enable colors, default is true")
	logLevelTruncation = flag.Bool("log-level-tc", true, "Enable the truncation of the level text to 4 characters, default is true")
)

var Log *logrus.Logger

func init() {
	Log = logrus.New()
	Log.Formatter = &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
}

func callerPrettyfier(f *runtime.Frame) (string, string) {
	// filename := f.File[strings.Index(f.File, "gateway"):]
	filename := path.Base(f.File)

	// fn := f.Function
	// fn = path.Base(f.Function)
	// fn = fn[strings.Index(fn, ".")+1:]

	// return fmt.Sprintf("%s() -", fn), fmt.Sprintf(" %s:%d", filename, f.Line)
	return "-", fmt.Sprintf(" %s:%d", filename, f.Line)
}

func InitLog() {
	writeCrashLog(*crashLog)

	formatter := &logrus.TextFormatter{
		QuoteEmptyFields:       true,
		DisableTimestamp:       !*logTimestamp,
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:04:05",
		ForceColors:            true,
		DisableColors:          !*logColors,
		DisableLevelTruncation: !*logLevelTruncation,
		CallerPrettyfier:       callerPrettyfier,
	}

	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("config logger level error. %v", errors.WithStack(err))
	}

	Log.Formatter = formatter
	Log.SetReportCaller(*logLineNumber)
	Log.Level = level
	Log.Out = os.Stdout

	if "" != *logFile {
		// Log.Out = ioutil.Discard
		err := os.MkdirAll(path.Dir(*logFile), os.ModePerm)
		if err != nil {
			log.Fatalf("make log dir error. %v", errors.WithStack(err))
		}

		_, err = os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
		if err != nil {
			log.Fatalf("open log file error. %v", errors.WithStack(err))
		}

		// Copy
		// lfFormatter := *formatter
		// writeLogFile(&lfFormatter)
		// writeErrorLogFile(&lfFormatter)

		lfFormatter := &TextFormatter{CallerPrettyfier: callerPrettyfier}

		Log.Formatter = lfFormatter
		writeLogFile(lfFormatter)
		writeErrorLogFile(lfFormatter)
	} /*else {
		Log.Out = os.Stdout
	}*/

}

func writeLogFile(formatter logrus.Formatter) {
	writer, err := rotatelogs.New(
		*logFile+".%Y%m%d",
		rotatelogs.WithLinkName(*logFile),         // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),     // 文件最大保存时间
		rotatelogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
	)
	if err != nil {
		log.Fatalf("config normal logger file error. %v", errors.WithStack(err))
	}

	Log.AddHook(lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, formatter))
}

func writeErrorLogFile(formatter logrus.Formatter) {
	errorFile := "error.log"
	writer, err := rotatelogs.New(
		path.Join(path.Dir(*logFile), errorFile+".%Y%m%d"),
		rotatelogs.WithLinkName(path.Join(path.Dir(*logFile), errorFile)),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		log.Fatalf("config error log file error. %v", errors.WithStack(err))
	}

	Log.AddHook(lfshook.NewHook(lfshook.WriterMap{
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, formatter))
}

func writeCrashLog(file string) {
	err := os.MkdirAll(path.Dir(file), os.ModePerm)
	if err != nil {
		log.Fatalf("make crash log dir error. %v", errors.WithStack(err))
	}

	crash.CrashLog(file)
}
