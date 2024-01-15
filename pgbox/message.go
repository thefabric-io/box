package pgbox

import (
	"database/sql"
	"encoding/json"
	"time"
)

type PostgresNewMessage struct {
	ID        sql.NullString `db:"id"`
	CreatedAt sql.NullTime   `db:"created_at"`
	Type      sql.NullString `db:"type"`
	Payload   sql.NullString `db:"payload"`
	Metadata  sql.NullString `db:"metadata"`
}

// PostgresMessage represents a message stored in PostgreSQL, implementing the Message interface.
type PostgresMessage struct {
	Offset_       sql.NullInt64  `db:"offset"`
	ID_           sql.NullString `db:"id"`
	CreatedAt_    sql.NullTime   `db:"created_at"`
	RegisteredAt_ sql.NullTime   `db:"registered_at"`
	Type_         sql.NullString `db:"type"`
	Payload_      sql.NullString `db:"payload"`
	Metadata_     sql.NullString `db:"metadata"`
}

func (m *PostgresMessage) RegisteredAt() time.Time {
	return m.RegisteredAt_.Time
}

func (m *PostgresMessage) Offset() int {
	return int(m.Offset_.Int64)
}

func (m *PostgresMessage) ID() string {
	return m.ID_.String
}

func (m *PostgresMessage) Type() string {
	return m.Type_.String
}

func (m *PostgresMessage) CreatedAt() time.Time {
	return m.CreatedAt_.Time
}

func (m *PostgresMessage) Payload() json.RawMessage {
	return json.RawMessage(m.Payload_.String)
}

func (m *PostgresMessage) Metadata() json.RawMessage {
	return json.RawMessage(m.Metadata_.String)
}
