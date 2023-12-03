package logger

import (
	"fmt"
	"go.uber.org/zap/zapcore"
	"math/rand"
	"time"
)

const maxLogMessageLen int = 10000

func init() {
	rand.Seed(time.Now().Unix())
}

func preCheckMsgLen(template string, fmtArgs []interface{}) []string {
	msg := template
	if msg == "" && len(fmtArgs) > 0 {
		msg = fmt.Sprint(fmtArgs...)
	} else if msg != "" && len(fmtArgs) > 0 {
		msg = fmt.Sprintf(template, fmtArgs...)
	}

	var splitMsg = make([]string, 0)
	var l = len(msg)
	var id = rand.Int31()
	for i := 0; i <= len(msg)/maxLogMessageLen; i++ {
		var chunkLen = (i + 1) * maxLogMessageLen
		if l < chunkLen {
			chunkLen = l
		}

		var chunkMsg = msg[i*maxLogMessageLen : chunkLen]
		if i == 0 || len(chunkMsg) > 0 {
			if l > maxLogMessageLen {
				chunkMsg = fmt.Sprint(id, ":", i, " ", chunkMsg)
			}

			splitMsg = append(splitMsg, chunkMsg)
		}
	}

	return splitMsg
}

func (l *HPLog) Debug(args ...interface{}) {
	chunks := preCheckMsgLen("", args)
	for _, c := range chunks {
		l.SugaredLogger.Debug(c)
	}
}

func (l *HPLog) Info(args ...interface{}) {
	chunks := preCheckMsgLen("", args)

	for _, c := range chunks {
		l.SugaredLogger.Info(c)
	}
}

func (l *HPLog) Warn(args ...interface{}) {
	chunks := preCheckMsgLen("", args)
	for _, c := range chunks {
		l.SugaredLogger.Warn(c)
	}
}

func (l *HPLog) Error(args ...interface{}) {
	chunks := preCheckMsgLen("", args)
	for _, c := range chunks {
		l.SugaredLogger.Error(c)
	}
}

func (l *HPLog) Panic(args ...interface{}) {
	l.SugaredLogger.Panic(args...)

}

func (l *HPLog) Fatal(args ...interface{}) {
	l.SugaredLogger.Fatal(args...)
}

func (l *HPLog) Debugf(template string, args ...interface{}) {
	chunks := preCheckMsgLen(template, args)
	for _, c := range chunks {
		l.SugaredLogger.Debug(c)
	}
}

func (l *HPLog) Infof(template string, args ...interface{}) {
	chunks := preCheckMsgLen(template, args)
	for _, c := range chunks {
		l.SugaredLogger.Info(c)
	}
}

func (l *HPLog) Warnf(template string, args ...interface{}) {
	chunks := preCheckMsgLen(template, args)
	for _, c := range chunks {
		l.SugaredLogger.Warn(c)
	}
}

func (l *HPLog) Errorf(template string, args ...interface{}) {
	chunks := preCheckMsgLen(template, args)
	for _, c := range chunks {
		l.SugaredLogger.Error(c)
	}
}

func (l *HPLog) Panicf(template string, args ...interface{}) {
	l.SugaredLogger.Panicf(template, args...)
}

func (l *HPLog) SetLevel(level zapcore.Level) {
	l.logLevel.SetLevel(level)
}

func (l *HPLog) GetLevel() string {
	return l.logLevel.Level().String()
}
