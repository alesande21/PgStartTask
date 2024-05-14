package api

import "fmt"

type Attribute int

const (
	FgBlack Attribute = iota + 30
	FgRed
	FgGreen
	FgYellow
	FgBlue
	FgMagenta
	FgCyan
	FgWhite
)

func ColorString(color Attribute, str string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, str)
}
