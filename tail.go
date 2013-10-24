package tail

import (
	"bufio"
	"github.com/snormore/gologger"
	"io"
	"launchpad.net/tomb"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Tailer struct {
	lines  chan string
	events chan string
}

const (
	CommandFindEvent     = "find_event.sh"
	CommandTailFromEvent = "tail_from_event.sh"
	CommandTailFromStart = "tail_from_start.sh"
	CommandTailFromEnd   = "tail_from_end.sh"
)

var (
	ScriptPath           = filepath.Join("./", "sbin")
	LinesChannelSize     = 1024
	FilePath             = "logs/development.json.log"
	SavePreviousEventMod = 1000
)

func NewTailer() *Tailer {
	t := Tailer{}
	t.lines = make(chan string, LinesChannelSize)
	return &t
}

func (t *Tailer) Listen() chan string {
	return t.lines
}

func (t *Tailer) Read(filePath string, startEvent string, tm *tomb.Tomb) error {
	return t.findAndRead(filePath, true, startEvent, tm)
}

func logStderr(stderrPipe io.ReadCloser) {
	reader := bufio.NewReader(stderrPipe)
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				logger.Error("Error reading stderr: %s", err)
				break
			}
			logger.Error("stderr: %s", line)
		}
	}()
}

func (t *Tailer) findAndRead(filePath string, follow bool, startEventId string, tm *tomb.Tomb) error {

	filePath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		return err
	}

	logger.Info("Searching %s for event %s...", filePath, startEventId)
	cmd := exec.Command(filepath.Join(ScriptPath, CommandFindEvent), filePath, startEventId)
	cmdStderr, _ := cmd.StderrPipe()
	logStderr(cmdStderr)
	eventLocation, err := cmd.Output()
	if err != nil {
		return err
	}

	if string(eventLocation) == "-1" {
		logger.Info("Event %s found to be the latest, reading from end of file...", startEventId)
		cmd = exec.Command(filepath.Join(ScriptPath, CommandTailFromEnd), strconv.FormatBool(follow), filePath)
	} else if string(eventLocation) == "0" {
		logger.Info("Event %s not found, reading from beginning of file...", startEventId)
		// airbrake.Notify(errors.New(fmt.Sprintf("Node %s failed to find event: %s", NodeId, startEventId)))
		cmd = exec.Command(filepath.Join(ScriptPath, CommandTailFromStart), strconv.FormatBool(follow), filePath)
	} else {
		logger.Info("Event %s found at line -%s, reading from here...", startEventId, eventLocation)
		cmd = exec.Command(filepath.Join(ScriptPath, CommandTailFromEvent), strconv.FormatBool(follow), filePath, startEventId)
	}
	cmdStderr, _ = cmd.StderrPipe()
	logStderr(cmdStderr)
	cmdStdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Panic("Error reading file with tail: %s", err)
	}
	cmdReader := bufio.NewReader(cmdStdout)
	cmd.Start()

	go func() {
		for {
			line, err := cmdReader.ReadString('\n')
			if err != nil {
				logger.Error("Error reading line from file: %s", err)
				tm.Killf("Error reading line from file: %s", err)
				break
			}
			logger.VerboseDebug("Read line: %s", line)
			select {
			case <-tm.Dying():
				return
			default:
			}
			t.lines <- strings.Trim(line, " \n")
		}
	}()

	<-tm.Dying()
	cmd.Process.Kill()
	return nil
}
