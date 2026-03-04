package common

import (
	"runtime"
	"strings"
)

const UnknownFn = "<fn unknown>"

func CallerName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return UnknownFn
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return UnknownFn
	}
	return fn.Name()
}

func CallerFuncOnly(skip int) string {
	name := CallerName(skip + 1)
	if name == UnknownFn {
		return name
	}
	if i := strings.LastIndex(name, "."); i != -1 {
		return name[i+1:]
	}
	return name
}

func CallerShort(skip int) string {
	name := CallerName(skip + 1)
	if name == "" {
		return "unknown"
	}

	if i := strings.LastIndex(name, "/"); i != -1 {
		return name[i+1:]
	}

	return name
}
