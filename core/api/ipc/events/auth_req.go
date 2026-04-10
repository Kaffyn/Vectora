package events

import (
	"encoding/json"
)

type PermissionPayload struct {
	Action    string `json:"action"` // e.g: "git_snapshot", "write_file", "run_terminal"
	FilePath  string `json:"file_path"`
	RequestID string `json:"request_id"`
}

// Connection interface abstracts the Named Pipe transport mechanism
type Connection interface {
	Write(b []byte) (n int, err error)
}

// SendPermissionRequest pushes the auth gate event to the local vectory systray or ide plugin
func SendPermissionRequest(conn Connection, payload PermissionPayload) error {
	msg := map[string]interface{}{
		"event_type": "permission_request",
		"payload":    payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Appends EOL to ensure atomic reading on the UI end
	data = append(data, '\n')
	_, err = conn.Write(data)
	return err
}
