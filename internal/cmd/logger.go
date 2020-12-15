package cmd

import (
	"fmt"
)

type Logger interface {
	Printf(string, ...interface{})
}

func (o *Options) Logf(format string, v ...interface{}) {
	if o.Logger != nil {
		o.Logger.Printf(format, v...)
	}
}

func (o *Options) Debugf(format string, v ...interface{}) {
	if o.Debug {
		o.Logf(format, v...)
	}
}

func (o *Options) Verbosef(format string, v ...interface{}) {
	if o.Verbose {
		fmt.Printf(format, v...)
		fmt.Print("\n")
	}
}
