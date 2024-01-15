package pgbox

import "database/sql"

type pgConsumerModel struct {
	Name           sql.NullString `db:"consumer_group"`
	OffsetAcked    sql.NullInt64  `db:"offset_acked"`
	OffsetConsumed sql.NullInt64  `db:"offset_consumed"`
}
