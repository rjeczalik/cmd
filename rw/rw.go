package rw

import "bytes"

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func indexnl(p []byte) (i, n int) {
	if i = bytes.IndexAny(p, "\r\n"); i != -1 {
		n = 1
		if i+1 < len(p) && p[i] == '\r' && p[i+1] == '\n' {
			n = 2
		}
	}
	return
}
