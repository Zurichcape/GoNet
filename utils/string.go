package utils

import "bytes"

func JoinString(pieces ...string) string {
	var buffer bytes.Buffer
	for i := 0; i < len(pieces); i++ {
		buffer.WriteString(pieces[i])
	}
	return buffer.String()
}
