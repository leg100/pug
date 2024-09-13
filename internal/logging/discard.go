package logging

var Discard Interface = &noop{}

type noop struct{}

func (noop) Debug(msg string, args ...any) {}

func (noop) Info(msg string, args ...any) {}

func (noop) Warn(msg string, args ...any) {}

func (noop) Error(msg string, args ...any) {}

func (noop) AddArgsUpdater(updater ArgsUpdater) {}
