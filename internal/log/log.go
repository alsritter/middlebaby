package log

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"alsritter.icu/middlebaby/internal/event"
)

// log levels
const (
	TraceLevel = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	Disabled
)

func init() {
	levelMap = make(map[string]int)
	levelMap["TRACE"] = TraceLevel
	levelMap["DEBUG"] = DebugLevel
	levelMap["INFO"] = InfoLevel
	levelMap["WARN"] = WarnLevel
	levelMap["ERROR"] = ErrorLevel
}

var (
	traceLog = log.New(os.Stderr, "\033[45m[trace]\033[0m ", log.LstdFlags|log.Lshortfile)
	debugLog = log.New(os.Stdout, "\033[36m[debug]\033[0m ", log.LstdFlags)
	warnLog  = log.New(os.Stdout, "\033[43m[warn]\033[0m ", log.LstdFlags)
	errorLog = log.New(os.Stderr, "\033[41m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[info ]\033[0m ", log.LstdFlags|log.Lshortfile)
	fatalLog = log.New(os.Stderr, "\033[1;37;41m[fatal]\033[0m ", log.LstdFlags|log.Lshortfile)
	loggers  = []*log.Logger{traceLog, errorLog, infoLog, debugLog, warnLog}
	levelMap map[string]int
	mu       sync.Mutex

	currentLevel = 1
)

// log methods
var (
	Trace  = traceLog.Println
	Tracef = traceLog.Printf

	Debug  = debugLog.Println
	Debugf = debugLog.Printf

	Error  = errorLog.Println
	Errorf = errorLog.Printf

	Info  = infoLog.Println
	Infof = infoLog.Printf

	Warn  = warnLog.Print
	Warnf = warnLog.Printf

	Fatal  = printFatal()  // call to os.Exit(1).
	Fatalf = printFatalf() // call to os.Exit(1).
)

func SetLevel(levelStr string) {
	level, ok := levelMap[levelStr]
	if !ok {
		level = 1
	}

	mu.Lock()
	defer mu.Unlock()
	currentLevel = level

	// Reset output level.
	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}

	if level > TraceLevel {
		traceLog.SetOutput(ioutil.Discard)
	}

	if level > DebugLevel {
		debugLog.SetOutput(ioutil.Discard)
	}

	if level > InfoLevel {
		infoLog.SetOutput(ioutil.Discard)
	}

	if level > WarnLevel {
		warnLog.SetOutput(ioutil.Discard)
	}

	if level > ErrorLevel {
		errorLog.SetOutput(ioutil.Discard)
	}
}

func GetCurrentLevel() int {
	return currentLevel
}

// TODO: need close all child process.
func printFatal() func(v ...interface{}) {
	return func(v ...interface{}) {
		fatalLog.Println(v...)
		kill()
	}
}

func printFatalf() func(format string, v ...interface{}) {
	return func(format string, v ...interface{}) {
		fatalLog.Printf(format, v...)
		kill()
	}
}

func kill() {
	defer func() {
		time.Sleep(1 * time.Second)
		os.Exit(1)
	}()

	event.Bus.Publish(event.CLOSE)
}
