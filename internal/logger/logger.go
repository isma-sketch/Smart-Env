package logger

import (
	"fmt"
	"os"
	"time"
)

var verbose bool

// SetVerbose enables verbose logging.
func SetVerbose(v bool) { verbose = v }

// IsVerbose reports verbose mode.
func IsVerbose() bool { return verbose }

func ts() string { return time.Now().Format("15:04:05") }

// Debug prints when verbose.
func Debug(format string, args ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stderr, "\033[90m[%s] DEBUG: \033[0m"+format+"\n", append([]interface{}{ts()}, args...)...)
	}
}

// Info prints informational message.
func Info(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// Warn prints warning.
func Warn(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "\033[33mWARN: \033[0m"+format+"\n", args...)
}

// Error prints error.
func Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "\033[31mERROR: \033[0m"+format+"\n", args...)
}

// Success prints success.
func Success(format string, args ...interface{}) {
	fmt.Printf("\033[32m✓ \033[0m"+format+"\n", args...)
}
