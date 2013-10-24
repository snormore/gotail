package tail

import (
	"encoding/json"
	"fmt"
	"github.com/snormore/gologger"
	"io/ioutil"
	"launchpad.net/tomb"
	"math"
	"os"
	"path"
	"strings"
)

const (
	PreviousEventsDir = "/tmp"
	EmptyAttribute    = "-"
)

type Event struct {
	Id string `json:"event_id"`
}

var (
	previousEventFilePath = path.Join(PreviousEventsDir, fmt.Sprintf("local.event"))
)

func getPreviousEventJson() string {
	if _, err := os.Stat(previousEventFilePath); os.IsNotExist(err) {
		logger.Info("No such file or directory: %s", previousEventFilePath)
		return ""
	}
	eventBytes, err := ioutil.ReadFile(previousEventFilePath)
	if err != nil {
		logger.Panic("Failed to open previous event file: %+v", err)
	}
	event := string(eventBytes)
	event = strings.Trim(string(event), " \n\r\t")
	if string(event) == EmptyAttribute {
		event = ""
	}
	return event
}

func getEventId(eventJson string) string {
	var event Event
	err := json.Unmarshal([]byte(eventJson), &event)
	if err != nil {
		logger.Error("Error parsing event JSON: %s", err)
	}
	return event.Id
}

func setPreviousEvent(event string) {
	err := ioutil.WriteFile(previousEventFilePath, []byte(event), 0644)
	if err != nil {
		logger.Panic("Failed to open previous-event file for writing: %+v", err)
	}
}

func eventsListener(events chan string, t *tomb.Tomb) {
	eventCounter := 0
	lastEvent := ""
	for {
		select {
		case event := <-events:
			if math.Mod(float64(eventCounter), float64(SavePreviousEventMod)) == 0.0 && strings.Trim(event, " \n\r\t") != "" {
				logger.VerboseDebug("Saving previous event ID %s...", event)
				setPreviousEvent(event)
			}
			if event != "" {
				lastEvent = event
			}
			eventCounter++
		case <-t.Dying():
			for {
				select {
				case e := <-events:
					if e != "" {
						lastEvent = e
					}
				default:
					if lastEvent != "" {
						logger.Info("Saving previous event ID %s before exit...", lastEvent)
						setPreviousEvent(lastEvent)
					}
					t.Done()
					return
				}
			}
		}
	}
}
