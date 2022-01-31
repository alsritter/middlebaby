package log

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
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
	debugLog = log.New(os.Stdout, "\033[46m[debug]\033[0m ", log.LstdFlags)
	warnLog  = log.New(os.Stdout, "\033[43m[warn]\033[0m ", log.LstdFlags)
	errorLog = log.New(os.Stderr, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[info ]\033[0m ", log.LstdFlags|log.Lshortfile)
	fatalLog = log.New(os.Stdout, "\033[1;37;41m[fatal]\033[0m ", log.LstdFlags|log.Lshortfile)
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

	Fatal  = fatalLog.Fatal  // call to os.Exit(1).
	Fatalf = fatalLog.Fatalf // call to os.Exit(1).
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
