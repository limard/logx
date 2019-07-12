package logxParse

import "time"

type LogxData struct {
	Date         time.Time
	FileName     string
	LineNo       int
	Level        string
	FunctionName string
	Content      string
}
