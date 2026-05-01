package common

import "bytes"

type LineBuffer struct {
	buf []byte
}

func (lb *LineBuffer) Append(data []byte) []string {
	lb.buf = append(lb.buf, data...)

	var lines []string

	for {
		i := bytes.IndexByte(lb.buf, '\n')
		if i == -1 {
			break
		}

		line := lb.buf[:i]

		// Handle Windows-style CRLF line endings.
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}

		lines = append(lines, string(line))

		// Remove the emitted line plus the newline byte.
		lb.buf = lb.buf[i+1:]
	}

	return lines
}
