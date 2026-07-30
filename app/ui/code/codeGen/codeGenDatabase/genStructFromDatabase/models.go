package genStructFromDatabase

import "database/sql"

type ColumnMetadataRow struct {
	COLUMN_NAME              string
	DATA_TYPE                string
	CHARACTER_MAXIMUM_LENGTH sql.NullInt64
	NUMERIC_PRECISION        sql.NullInt64
	NUMERIC_SCALE            sql.NullInt64
	IS_NULLABLE              sql.NullString
}

type TableMetadataRow struct {
	SchemaName string
	TableName  string
}
