package log

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
)

// log levels
const (
	DebugLevel = iota
	InfoLevel
	ErrorLevel
	Disabled
)

func init() {
	levelMap = make(map[string]int32)
	levelMap["DEBUG"] = DebugLevel
	levelMap["INFO"] = InfoLevel
	levelMap["ERROR"] = ErrorLevel
}

var (
	debugLevel = log.New(os.Stderr, "\033[37m[debug]\033[0m ", log.LstdFlags|log.Lshortfile)
	errorLog   = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog    = log.New(os.Stdout, "\033[34m[info ]\033[0m ", log.LstdFlags|log.Lshortfile)
	loggers    = []*log.Logger{errorLog, infoLog, debugLevel}
	levelMap   map[string]int32
	mu         sync.Mutex
)

// log methods
var (
	Debug  = debugLevel.Println
	Debugf = debugLevel.Printf

	Error  = errorLog.Println
	Errorf = errorLog.Printf

	Info  = infoLog.Println
	Infof = infoLog.Printf
)

func SetLevel(levelStr string) {
	level, ok := levelMap[levelStr]
	if !ok {
		level = 1
	}

	mu.Lock()
	defer mu.Unlock()

	// 这里是重置输出级别
	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}

	if level > DebugLevel {
		debugLevel.SetOutput(ioutil.Discard)
	}

	if level > InfoLevel {
		infoLog.SetOutput(ioutil.Discard)
	}

	if level > ErrorLevel {
		errorLog.SetOutput(ioutil.Discard)
	}
}
