package tail

import (
	"strings"
)

type Tailer struct {
	lines chan string
}

var (
	LinesChannelSize = 1024
)

func NewTailer() *Tailer {
	t := Tailer{}
	t.lines = make(chan string, LinesChannelSize)
	return &t
}

func (t *Tailer) Listen() chan string {
	return t.lines
}

func (t *Tailer) Push(log string) {
	lines := strings.Split(strings.Trim(log, " \n\r\t"), "\n")
	for _, line := range lines {
		t.lines <- strings.Trim(line, " \t\n\r")
	}
}

func (t *Tailer) Read(filePath string, startEvent string) error {
	return nil
}
