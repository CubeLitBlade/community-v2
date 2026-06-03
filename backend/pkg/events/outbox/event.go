package outbox

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
)

// NewEvent creates a new CloudEvent with the given parameters and marshals it into JSON bytes.
func NewEvent(
	source string, eventType string, payload any, id int64, now time.Time,
) ([]byte, error) {
	event := ce.NewEvent()
	event.SetID(strconv.FormatInt(id, 10))
	event.SetSource(source)
	event.SetType(eventType)
	event.SetTime(now)

	event.SetDataContentType(ce.ApplicationJSON)

	if err := event.SetData(ce.ApplicationJSON, payload); err != nil {
		return nil, fmt.Errorf("set data: %w", err)
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("marshal event: %w", err)
	}

	return bytes, nil
}
