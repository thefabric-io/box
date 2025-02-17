package pgbox

import (
	"database/sql"
	"encoding/json"
	"testing"
)

func TestPostgresMessage_Metadata(t *testing.T) {
	tests := []struct {
		name           string
		metadataField  sql.NullString
		expectedOutput json.RawMessage
	}{
		{
			name: "invalid metadata returns nil",
			metadataField: sql.NullString{
				Valid: false,
			},
			expectedOutput: nil,
		},
		{
			name: "valid metadata returns json.RawMessage",
			metadataField: sql.NullString{
				String: `{"key": "value"}`,
				Valid:  true,
			},
			expectedOutput: json.RawMessage(`{"key": "value"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &PostgresMessage{
				Metadata_: tt.metadataField,
			}

			result := msg.Metadata()
			if tt.expectedOutput == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
			} else {
				if string(result) != string(tt.expectedOutput) {
					t.Errorf("expected %s, got %s", tt.expectedOutput, result)
				}
			}
		})
	}
}
