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
		var common Common
		if err := json.Unmarshal(scanner.Bytes(), &common); err != nil {
			return nil, fmt.Errorf("unmarshaling common fields: %w", err)
		}
		var dst Message
		switch common.Type {
		case MessageVersion:
			dst = new(VersionMsg)
		case MessagePlannedChange:
			dst = new(PlannedChangeMsg)
		case MessageChangeSummary:
			dst = new(ChangeSummaryMsg)
		case MessageOutputs:
			dst = new(OutputMsg)
		default:
			continue
		}
		if err := json.Unmarshal(scanner.Bytes(), dst); err != nil {
			return nil, fmt.Errorf("unmarshaling message of type %T: %w", dst, err)
		}
		messages = append(messages, dst)
	}
	return messages, nil
}
