package gnet

import (
	"io"
)

func bodyLogger(body io.Reader, logger io.Writer) (io.Reader, func()) {
	if logger == nil {
		return body, func(){}
	}

	io.WriteString(logger, "--- gnet logger body begin ---\n")
	r := io.TeeReader(body, logger)
	return r, func() {
		io.WriteString(logger, "\n--- gnet logger body end ---\n")
	}
}
