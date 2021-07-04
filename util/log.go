package util

import "github.com/go-logr/logr"

type Logger logr.Logger

var NopLog Logger = (*nop)(nil)

type nop struct{}

func (n *nop) Enabled() bool                                             { return false }
func (n *nop) Error(err error, msg string, keysAndValues ...interface{}) { return }
func (n *nop) Info(msg string, keysAndValues ...interface{})             { return }
func (n *nop) V(level int) logr.Logger                                   { return n }
func (n *nop) WithValues(keysAndValues ...interface{}) logr.Logger       { return n }
func (n *nop) WithName(name string) logr.Logger                          { return n }
