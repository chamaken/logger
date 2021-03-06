// sys/syslog.h
//	Copyright (c) 1982, 1986, 1988, 1993
//	The Regents of the University of California.  All rights reserved.
//
// This package only provides a simple priority logging by wrapping
// original log package.

// これはオリジナルパッケージ log のラッパーです。NewLogger() で作成したインスタ
// ンスの SetLevel() で出力レベルを変更します。グローバルな logger に対しては
// SetLevel() あるいは環境変数 GOLOGLEVEL を debug や err に設定することで出力レ
// ベルを変更することができます。Facility は定義してあるだけで使っていません。
//
// This is a wrapper of Go original log package. Log level can be changed by
// SetLevel() method when the instance is created NewLogger(). Global Logger level can
// also be changed by SetLevel() function or setting GOLOGLEVEL env value to
// debug, err and stuff like that.

package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type Level int
const (
	LOG_EMERG	= Level(0)	// system is unusable
	LOG_ALERT	= Level(1)	// action must be taken immediately
	LOG_CRIT	= Level(2)	// critical conditions
	LOG_ERR		= Level(3)	// error conditions
	LOG_WARNING	= Level(4)	// warning conditions
	LOG_NOTICE	= Level(5)	// normal but significant condition
	LOG_INFO	= Level(6)	// informational
	LOG_DEBUG	= Level(7)	// debug-level messages

	LOG_PRIMASK	= 0x07		// mask to extract priority part (internal)
					// extract priority
	INTERNAL_NOPRI	= 0x10		// the "no priority" priority
	INTERNAL_MARK	= (LOG_NFACILITIES << 3) | 0
)

type Facility int
const (
	LOG_KERN	= Facility(0 << 3)	// kernel messages
	LOG_USER	= Facility(1 << 3)	// random user-level messages
	LOG_MAIL	= Facility(2 << 3)	// mail system
	LOG_DAEMON	= Facility(3 << 3)	// system daemons
	LOG_AUTH	= Facility(4 << 3)	// security/authorization messages
	LOG_SYSLOG	= Facility(5 << 3)	// messages generated internally by syslogd
	LOG_LPR		= Facility(6 << 3)	// line printer subsystem
	LOG_NEWS	= Facility(7 << 3)	// network news subsystem
	LOG_UUCP	= Facility(8 << 3)	// UUCP subsystem
	LOG_CRON	= Facility(9 << 3)	// clock daemon
	LOG_AUTHPRIV	= Facility(10 << 3)	// security/authorization messages (private)
	LOG_FTP		= Facility(11 << 3)	// ftp daemon

	// other codes through 15 reserved for system use
	LOG_LOCAL0	= Facility(16 << 3)	// reserved for local use
	LOG_LOCAL1	= Facility(17 << 3)	// reserved for local use
	LOG_LOCAL2	= Facility(18 << 3)	// reserved for local use
	LOG_LOCAL3	= Facility(19 << 3)	// reserved for local use
	LOG_LOCAL4	= Facility(20 << 3)	// reserved for local use
	LOG_LOCAL5	= Facility(21 << 3)	// reserved for local use
	LOG_LOCAL6	= Facility(22 << 3)	// reserved for local use
	LOG_LOCAL7	= Facility(23 << 3)	// reserved for local use

	LOG_NFACILITIES	= 24		// current number of facilities
	LOG_FACMASK	= 0x03f8	// mask to extract facility part
					// facility of pri
)

func LOG_PRI(p int) Level {
	return Level((p) & LOG_PRIMASK)
}

func LOG_MAKEPRI(fac Facility, pri Level) int {
	return ((int(fac) << 3) | int(pri))
}

var Levels = map[Level]string {
	LOG_ALERT:	"alert",
	LOG_CRIT:	"crit",	
	LOG_DEBUG:	"debug",
	LOG_EMERG:	"emerg",
	LOG_ERR:	"err",
	LOG_INFO:	"info",
	INTERNAL_NOPRI:	"none",
	LOG_NOTICE:	"notice",
	LOG_WARNING:   	"warning",
}

var Facilities = map[Facility]string {
	LOG_AUTH:	"auth",
	LOG_AUTHPRIV:	"authpriv",
	LOG_CRON:	"cron",
	LOG_DAEMON:	"daemon",
	LOG_FTP:	"ftp",
	LOG_KERN:	"kern",
	LOG_LPR:	"lpr",
	LOG_MAIL:	"mail",
	INTERNAL_MARK:	"mark",     	// INTERNAL
	LOG_NEWS:	"news",
	// LOG_AUTH:	"security", 	// DEPRECATED
	LOG_SYSLOG:	"syslog",
	LOG_USER:	"user",
	LOG_UUCP:	"uucp",
	LOG_LOCAL0:	"local0",
	LOG_LOCAL1:	"local1",
	LOG_LOCAL2:	"local2",
	LOG_LOCAL3:	"local3",
	LOG_LOCAL4:	"local4",
	LOG_LOCAL5:	"local5",
	LOG_LOCAL6:	"local6",
	LOG_LOCAL7:	"local7",
}

func LOG_FAC(p int) Facility {
	return Facility(((p) & LOG_FACMASK) >> 3)
}

func LOG_MASK(pri Level) int {
	return (1 << uint(pri))		// mask for one priority
}

func LOG_UPTO(pri Level) int {
	return ((1 << (uint(pri) + 1)) - 1)	// all priorities through pri
}

type Logger struct {
	logger *log.Logger
	upto int
}

func NewLogger(out io.Writer, prefix string, flag int, priority Level) *Logger {
	return &Logger{log.New(out, prefix, flag), LOG_UPTO(priority)}
}

func (l *Logger) Flags() int {
	return l.logger.Flags()
}
func (l *Logger) Prefix() string {
	return l.logger.Prefix()
}
func (l *Logger) Priority() Level {
	// return Level(math.Log2(float64(^l.upto & (l.upto + 1))))
	// http://stackoverflow.com/questions/2380728/getting-the-number-of-trailing-1-bits
	b := int32(^l.upto & (l.upto + 1))            // this gives a 1 to the left of the trailing 1's
	b--                                           // this gets us just the trailing 1's that need counting
	b = (b & 0x55555555) + ((b>>1)  & 0x55555555) // 2 bit sums of 1 bit numbers
	b = (b & 0x33333333) + ((b>>2)  & 0x33333333) // 4 bit sums of 2 bit numbers
	b = (b & 0x0f0f0f0f) + ((b>>4)  & 0x0f0f0f0f) // 8 bit sums of 4 bit numbers
	b = (b & 0x00ff00ff) + ((b>>8)  & 0x00ff00ff) // 16 bit sums of 8 bit numbers
	b = (b & 0x0000ffff) + ((b>>16) & 0x0000ffff) // sum of 16 bit numbers
	return Level(b)
}
func (l *Logger) SetFlags(flag int) {
	l.logger.SetFlags(flag)
}
func (l *Logger) SetPrefix(prefix string) {
	l.logger.SetPrefix(prefix)
}
func (l *Logger) SetPriority(priority Level) {
	l.upto = LOG_UPTO(priority)
}

func (l *Logger) Panic(format string, v ...interface{}) {
	s := fmt.Sprintf("[panic] " + format, v...)
	l.logger.Output(3, s)
	panic(s)
}
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.logger.Output(3, fmt.Sprintf("[fatal] " + format, v...))
	os.Exit(1)
}
func (l *Logger) printf(format string, prio Level, v ...interface{}) {
	l.logger.Output(3, fmt.Sprintf(fmt.Sprintf("[%s] %s", Levels[prio], format), v...))
}
func (l *Logger) Emerg(format string, v ...interface{}) {
	if l.upto & LOG_MASK(LOG_EMERG) != 0 { l.printf(format, LOG_EMERG, v...) }
}
func (l *Logger) Alert(format string, v ...interface{}) {
	if l.upto & LOG_MASK(LOG_ALERT) != 0 { l.printf(format, LOG_ALERT, v...) }
}
func (l *Logger) Crit(format string, v ...interface{}) {
	if l.upto & LOG_MASK(LOG_CRIT) != 0 { l.printf(format, LOG_CRIT, v...) }
}
func (l *Logger) Error(format string, v ...interface{}) {
	if l.upto & LOG_MASK(LOG_ERR) != 0 { l.printf(format, LOG_ERR, v...) }
}
func (l *Logger) Warning(format string, v ...interface{}) {
	if l.upto & LOG_MASK(LOG_WARNING) != 0 { l.printf(format, LOG_WARNING, v...) }
}
func (l *Logger) Notice(format string, v ...interface{}) {
	if l.upto & LOG_MASK(LOG_NOTICE) != 0 { l.printf(format, LOG_NOTICE, v...) }
}
func (l *Logger) Info(format string, v ...interface{}) {
	if l.upto & LOG_MASK(LOG_INFO) != 0 { l.printf(format, LOG_INFO, v...) }
}
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.upto & LOG_MASK(LOG_DEBUG) != 0 {  l.printf(format, LOG_DEBUG, v...) }
}

// function
func init() {
	s := os.Getenv("GOLOGLEVEL")
	for k, v := range(Levels) {
		if strings.ToLower(s) == v {
			SetPriority(k)
			break
		}
	}
}

func SetOutput(out io.Writer) {
	std = NewLogger(out, Prefix(), Flags(), Priority())
}
func Flags() int {
	return std.Flags()
}
func Prefix() string {
	return std.Prefix()
}
func Priority() Level {
	return std.Priority()
}
func SetFlags(flag int) {
	std.SetFlags(flag)
}
func SetPrefix(prefix string) {
	std.SetPrefix(prefix)
}
func SetPriority(priority Level) {
	std.SetPriority(priority)
}

var std = NewLogger(os.Stderr, "", log.LstdFlags, LOG_ERR)

func Panic(format string, v ...interface{}) {
	s := fmt.Sprintf("[panic] " + format, v...)
	std.logger.Output(3, s)
	panic(s)
}
func Fatal(format string, v ...interface{}) {
	std.logger.Output(3, fmt.Sprintf("[fatal] " + format, v...))
	os.Exit(1)
}
func Emerg(format string, v ...interface{}) {
	if std.upto & LOG_MASK(LOG_EMERG) != 0 { std.printf(format, LOG_EMERG, v...) }
}
func Alert(format string, v ...interface{}) {
	if std.upto & LOG_MASK(LOG_ALERT) != 0 { std.printf(format, LOG_ALERT, v...) }
}
func Crit(format string, v ...interface{}) {
	if std.upto & LOG_MASK(LOG_CRIT) != 0 { std.printf(format, LOG_CRIT, v...) }
}
func Error(format string, v ...interface{}) {
	if std.upto & LOG_MASK(LOG_ERR) != 0 { std.printf(format, LOG_ERR, v...) }
}
func Warning(format string, v ...interface{}) {
	if std.upto & LOG_MASK(LOG_WARNING) != 0 { std.printf(format, LOG_WARNING, v...) }
}
func Notice(format string, v ...interface{}) {
	if std.upto & LOG_MASK(LOG_NOTICE) != 0 { std.printf(format, LOG_NOTICE, v...) }
}
func Info(format string, v ...interface{}) {
	if std.upto & LOG_MASK(LOG_INFO) != 0 { std.printf(format, LOG_INFO, v...) }
}
func Debug(format string, v ...interface{}) {
	if std.upto & LOG_MASK(LOG_DEBUG) != 0 {  std.printf(format, LOG_DEBUG, v...) }
}
