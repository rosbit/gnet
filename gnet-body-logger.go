package gnet

import (
	logr "github.com/rosbit/reader-logger"
	"io"
)

func bodyLogger(body io.Reader, logger io.Writer) (io.Reader, func()) {
	return logr.ReaderLogger(body, logger, "gnet logger body")
}
