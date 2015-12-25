// don't want to conflict with stdlib's `log` package
package atc_log

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
)

var (
	// Set to true to enable debugging output
	DEBUG bool = false
)

func Syslog() *log.Logger {
	log, err := syslog.NewLogger(syslog.LOG_USER|syslog.LOG_INFO, 0)
	if err != nil {
		log.Println("warning: Could not create syslog logger:", err)
	}
	return log
}

func Stdlog() *log.Logger {
	return log.New(os.Stderr, "", log.Ldate|log.Ltime)
}

type LogMux struct {
	loggers []*log.Logger
}

func NewMux(loggers ...*log.Logger) *LogMux {
	return &LogMux{loggers}
}

func (l *LogMux) Fatal(v ...interface{}) {
	for _, l := range l.loggers {
		l.Print(v...)
	}
	os.Exit(1)
}

func (l *LogMux) Fatalf(format string, v ...interface{}) {
	for _, l := range l.loggers {
		l.Printf(format, v...)
	}
	os.Exit(1)
}

func (l *LogMux) Fatalln(v ...interface{}) {
	for _, l := range l.loggers {
		l.Println(v...)
	}
	os.Exit(1)
}

func (l *LogMux) Panic(v ...interface{}) {
	for _, l := range l.loggers {
		l.Print(v...)
	}
	panic(fmt.Sprint(v...))
}

func (l *LogMux) Panicf(format string, v ...interface{}) {
	for _, l := range l.loggers {
		l.Printf(format, v...)
	}
	panic(fmt.Sprintf(format, v...))
}

func (l *LogMux) Panicln(v ...interface{}) {
	for _, l := range l.loggers {
		l.Println(v...)
	}
	panic(fmt.Sprint(v...))
}

func (l *LogMux) Print(v ...interface{}) {
	for _, l := range l.loggers {
		l.Print(v...)
	}
}

func (l *LogMux) Printf(format string, v ...interface{}) {
	for _, l := range l.loggers {
		l.Printf(format, v...)
	}
}

func (l *LogMux) Println(v ...interface{}) {
	for _, l := range l.loggers {
		l.Println(v...)
	}
}

func (l *LogMux) Debug(v ...interface{}) {
	if DEBUG {
		for _, l := range l.loggers {
			l.Print(v...)
		}
	}
}

func (l *LogMux) Debugf(format string, v ...interface{}) {
	if DEBUG {
		for _, l := range l.loggers {
			l.Printf(format, v...)
		}
	}
}

func (l *LogMux) Debugln(v ...interface{}) {
	if DEBUG {
		for _, l := range l.loggers {
			l.Println(v...)
		}
	}
}
