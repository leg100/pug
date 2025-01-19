package machine

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// Unmarshal a stream of json objects into machine readable UI messages
func Unmarshal(r io.Reader) ([]Message, error) {
	var messages []Message
	for scanner := bufio.NewScanner(r); scanner.Scan(); {
		msg, err := UnmarshalMessage(scanner.Bytes())
		if err != nil {
			return nil, err
		}
		if msg == nil {
			continue
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func UnmarshalMessage(b []byte) (Message, error) {
	var (
		msg    Message
		common Common
	)
	if err := json.Unmarshal(b, &common); err != nil {
		return nil, fmt.Errorf("unmarshaling common fields: %w", err)
	}
	switch common.Type {
	case MessageVersion:
		msg = new(VersionMsg)
	case MessagePlannedChange:
		msg = new(PlannedChangeMsg)
	case MessageChangeSummary:
		msg = new(ChangeSummaryMsg)
	case MessageOutputs:
		msg = new(OutputMsg)
	default:
		return nil, nil
	}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, fmt.Errorf("unmarshaling message of type %T: %w", msg, err)
	}
	return msg, nil
}
