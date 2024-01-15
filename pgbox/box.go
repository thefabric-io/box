package pgbox

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/thefabric-io/box"
	"github.com/thefabric-io/sqlvalues"
	"github.com/thefabric-io/transactional"
)

// NewPostgresBox creates a new instance of postgresBox.
func NewPostgresBox(schema, topic, _type string) box.Box {
	return &postgresBox{
		schema: schema,
		topic:  topic,
		type_:  _type,
	}
}

// postgresBox implements the Box interface for PostgreSQL.
type postgresBox struct {
	schema string
	topic  string
	type_  string
}

func (pb *postgresBox) DefaultLogFields() map[string]any {
	_, file, _, _ := runtime.Caller(0)

	return map[string]any{
		"metadata": map[string]any{
			"file": file,
		},
		"subject": map[string]any{
			"system":            "postgres",
			"schema":            pb.schema,
			"topic":             pb.topic,
			"box_type":          pb.type_,
			"table_name":        pb.fullTableNameBox(),
			"table_name_offset": pb.fullTableNameOffset(),
		},
	}
}

func (pb *postgresBox) Exists(ctx context.Context, transaction transactional.Transaction, id string) (bool, error) {
	tx := transaction.(*sqlx.Tx)

	query := fmt.Sprintf(`
		select "offset", id, created_at, type, payload, metadata 
		from %s where id = $1;`,
		pb.fullTableNameBox())

	message := &PostgresMessage{}
	// Query the database
	if err := tx.GetContext(ctx, message, query, id); err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return false, nil
		}

		return false, fmt.Errorf("error querying messages: %w", err)
	}

	if message == nil {
		return false, nil
	}

	return true, nil
}

func (pb *postgresBox) fullTableNameBox() string {
	return pb.schema + "." + pb.type_ + "_" + pb.topic
}

func (pb *postgresBox) fullTableNameOffset() string {
	return pb.schema + "." + pb.type_ + "_" + pb.topic + "_offsets"
}

// Init initializes the PostgreSQL database.
func (pb *postgresBox) Init(ctx context.Context, transaction transactional.Transaction) error {
	tx := transaction.(*sqlx.Tx)

	strBuilder := strings.Builder{}

	strBuilder.WriteString(`CREATE SCHEMA IF NOT EXISTS ` + pb.schema + `;`)

	strBuilder.WriteString(`
        CREATE TABLE IF NOT EXISTS ` + pb.fullTableNameOffset() + ` (
		consumer_group VARCHAR(255) NOT NULL,
		offset_acked BIGINT,
		offset_consumed BIGINT NOT NULL,
		PRIMARY KEY(consumer_group));`)

	strBuilder.WriteString(`CREATE TABLE IF NOT EXISTS ` + pb.fullTableNameBox() + ` (
		"offset" SERIAL,
		id VARCHAR NOT NULL unique,
		type varchar NOT NULL,
		created_at TIMESTAMPTZ NOT NULL,
		registered_at TIMESTAMPTZ NOT NULL DEFAULT (current_timestamp AT TIME ZONE 'UTC'),
		payload JSONB DEFAULT NULL,
		metadata JSONB DEFAULT NULL);`)

	strBuilder.WriteString(`
		create index if not exists inbox_brands_offset_index on ` + pb.fullTableNameBox() + ` ("offset");
		`)

	if _, err := tx.ExecContext(ctx, strBuilder.String()); err != nil {
		return err
	}

	return nil
}

// Receive implements the Receive method of the Box interface.
func (pb *postgresBox) Receive(ctx context.Context, transaction transactional.Transaction, msg box.Message) error {
	tx := transaction.(*sqlx.Tx)

	query := `INSERT INTO ` + pb.fullTableNameBox() + ` (id, created_at, type, payload, metadata) VALUES ($1, $2, $3, $4, $5)`

	pgMsg := PostgresNewMessage{
		ID:        sqlvalues.SqlString(msg.ID()),
		CreatedAt: sqlvalues.SqlTime(msg.CreatedAt()),
		Type:      sqlvalues.SqlString(msg.Type()),
		Payload:   sqlvalues.SqlNullJSONB(msg.Payload()),
		Metadata:  sqlvalues.SqlNullJSONB(msg.Metadata()),
	}

	_, err := tx.ExecContext(ctx, query, pgMsg.ID, pgMsg.CreatedAt, pgMsg.Type, pgMsg.Payload, pgMsg.Metadata)
	if err != nil {
		return fmt.Errorf("error inserting message: %w", err)
	}

	return nil
}

// Retrieve fetches a batch of messages from the PostgreSQL database.
func (pb *postgresBox) Retrieve(ctx context.Context, transaction transactional.Transaction, consumer string, onlyTypes []string, batchSize int, waitTime time.Duration) ([]box.Message, error) {
	tx := transaction.(*sqlx.Tx)

	csm, err := pb.getConsumer(ctx, tx, consumer)
	if err != nil {
		return nil, err
	}

	return pb.getMessages(ctx, tx, onlyTypes, int(csm.OffsetAcked.Int64), batchSize, waitTime)
}

func (pb *postgresBox) getMessages(ctx context.Context, tx *sqlx.Tx, onlyTypes []string, offsetAcked, batchSize int, waitTime time.Duration) ([]box.Message, error) {
	query := fmt.Sprintf(`
		select "offset", id, created_at, registered_at, type, payload, metadata 
		from %s where "offset" > $1 
		and type = any($2)
		order by "offset" asc
		limit $3;`,
		pb.fullTableNameBox())

	messages := make([]box.Message, 0)

	rows, err := tx.QueryContext(ctx, query, offsetAcked, pq.Array(onlyTypes), batchSize)
	if err != nil {
		return nil, fmt.Errorf("error querying messages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var msg PostgresMessage
		if err := rows.Scan(&msg.Offset_, &msg.ID_, &msg.CreatedAt_, &msg.RegisteredAt_, &msg.Type_, &msg.Payload_, &msg.Metadata_); err != nil {
			return nil, fmt.Errorf("error scanning message: %w", err)
		}
		messages = append(messages, &msg)
	}

	// Check for errors after iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error in rows iteration: %w", err)
	}

	return messages, nil
}

func (pb *postgresBox) getConsumer(ctx context.Context, tx *sqlx.Tx, consumer string) (*pgConsumerModel, error) {
	query := fmt.Sprintf(`
		select consumer_group, offset_acked, offset_consumed 
		from %s where consumer_group = $1;`,
		pb.fullTableNameOffset())

	var result pgConsumerModel

	if err := tx.GetContext(ctx, &result, query, consumer); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := pb.insertConsumer(ctx, tx, consumer); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &result, nil
}

func (pb *postgresBox) UpdateConsumer(ctx context.Context, transaction transactional.Transaction, name string, offsetAcked, offsetConsumed int) error {
	tx := transaction.(*sqlx.Tx)

	query := fmt.Sprintf(`update %s set offset_acked = $1, offset_consumed = $2 where consumer_group = $3;`, pb.fullTableNameOffset())

	_, err := tx.ExecContext(ctx, query, offsetAcked, offsetConsumed, name)
	if err != nil {
		return err
	}

	return nil
}

func (pb *postgresBox) insertConsumer(ctx context.Context, tx *sqlx.Tx, consumer string) error {
	strBuilder := strings.Builder{}

	strBuilder.WriteString(`insert into ` + pb.fullTableNameOffset() + ` (consumer_group, offset_acked, offset_consumed) values ($1, 0, 0);`)

	if _, err := tx.ExecContext(ctx, strBuilder.String(), consumer); err != nil {
		return err
	}

	return nil
}
