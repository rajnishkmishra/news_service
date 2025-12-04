package utils

import (
	"fmt"
	"net"
	"os"
	"runtime/debug"
	"syscall"

	"github.com/sirupsen/logrus"
)

type WrapperError interface {
	error
	HttpCode() int
}

type simpleWrapperError struct {
	err    error
	status int
}

func (s *simpleWrapperError) HttpCode() int {
	return s.status
}

func (s *simpleWrapperError) Error() string {
	return s.err.Error()
}

func NewWrapperError(status int, err error) *simpleWrapperError {
	return &simpleWrapperError{
		status: status,
		err:    err,
	}
}

func Recovery() {
	if r := recover(); r != nil {
		if ne, ok := r.(*net.OpError); ok {
			if se, ok := ne.Err.(*os.SyscallError); ok {
				if se.Err == syscall.EPIPE || se.Err == syscall.ECONNRESET {
					return
				}
			}
		}

		stackString := string(debug.Stack())
		fmt.Printf(stackString)
		err := fmt.Errorf("Recovered from following error:", r)
		logrus.WithError(err).Logln(logrus.ErrorLevel)
	}
}
