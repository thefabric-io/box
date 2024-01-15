package box

import (
	"encoding/json"
	"time"
)

func NewMessage(id string, _type string, createdAt time.Time, payload, metadata json.RawMessage) Message {
	return &message{
		id:        id,
		_type:     _type,
		createdAt: createdAt,
		payload:   payload,
		metadata:  metadata,
	}
}

type message struct {
	id           string
	_type        string
	createdAt    time.Time
	registeredAt time.Time
	payload      json.RawMessage
	metadata     json.RawMessage
}

func (m *message) Offset() int {
	return 0
}

func (m *message) ID() string {
	return m.id
}

func (m *message) Type() string {
	return m._type
}

func (m *message) CreatedAt() time.Time {
	return m.createdAt
}

func (m *message) RegisteredAt() time.Time {
	return m.createdAt
}

func (m *message) Payload() json.RawMessage {
	return m.payload
}

func (m *message) Metadata() json.RawMessage {
	return m.metadata
}
