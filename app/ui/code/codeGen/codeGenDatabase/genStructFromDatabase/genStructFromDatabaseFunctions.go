package genStructFromDatabase

import (
	"database/sql"
	"embed"
	"log"
	"strings"
	"testAutomationSuiteGO/app/shared"
)

//go:embed sql/*.sql
var Queries embed.FS

func BuildTableRowStruct(serverName, dbName, tableName string, deps shared.AppDependencies) string {
	if strings.Contains(tableName, ".") {
		parts := strings.SplitN(tableName, ".", 2)
		if len(parts) != 2 {
			log.Println("Invalid table name format. Expected 'schema.table', got: " + tableName)
			return "Invalid table name format."
		}
		tableName = parts[1]
	}
	tableColumnMetadata := GetDatabaseTableTypeMapping(serverName, dbName, tableName, deps)
	//log.Println("Length of tableColumnMetadata: " + strconv.Itoa(len(tableColumnMetadata)))
	var structCode strings.Builder
	structCode.WriteString("type ")
	structCode.WriteString(tableName)
	structCode.WriteString(" struct {\n")
	for _, col := range tableColumnMetadata {
		goType := mapSQLTypeToGoType(col.DATA_TYPE, col.IS_NULLABLE)
		structCode.WriteString("\t")
		structCode.WriteString(col.COLUMN_NAME)
		structCode.WriteString(" ")
		structCode.WriteString(goType)
		structCode.WriteString("\n")
	}
	structCode.WriteString("}\n")
	return structCode.String()
}

func mapSQLTypeToGoType(sqlType string, isNullable sql.NullString) string {
	typeMapping := map[string]string{
		"bigint":           "int64",
		"binary":           "[]byte",
		"bit":              "bool",
		"char":             "string",
		"date":             "time.Time",
		"datetime":         "time.Time",
		"datetime2":        "time.Time",
		"datetimeoffset":   "time.Time",
		"decimal":          "float64",
		"float":            "float64",
		"geography":        "string",
		"geometry":         "string",
		"hierarchyid":      "string",
		"image":            "[]byte",
		"int":              "int",
		"intList":          "[]int",
		"money":            "float64",
		"nchar":            "string",
		"ntext":            "string",
		"numeric":          "float64",
		"nvarchar":         "string",
		"real":             "float32",
		"smalldatetime":    "time.Time",
		"smallint":         "int16",
		"smallmoney":       "float32",
		"sql_variant":      "interface{}",
		"sysname":          "string",
		"text":             "string",
		"time":             "time.Time",
		"timestamp":        "[]byte",
		"tinyint":          "uint8",
		"uniqueidentifier": "string",
		"varbinary":        "[]byte",
		"varchar":          "string",
		"xml":              "string",
	}
	goType, ok := typeMapping[sqlType]
	if !ok {
		goType = "interface{}"
	}
	if isNullable.Valid && isNullable.String == "YES" && goType != "interface{}" {
		goType = ConvertGoTypeToSqlNullType(goType)
	}
	return goType
}

func ConvertGoTypeToSqlNullType(goType string) string {
	switch goType {
	case "string":
		return "sql.NullString"
	case "int":
		return "sql.NullInt64"
	case "int64":
		return "sql.NullInt64"
	case "float64":
		return "sql.NullFloat64"
	case "bool":
		return "sql.NullBool"
	case "time.Time":
		return "sql.NullString"
	case "[]byte":
		return "sql.NullByte"
	default:
		return goType
	}
}

func GetDatabaseTableTypeMapping(dbServer, dbName, tableName string, deps shared.AppDependencies) []ColumnMetadataRow {
	rows, err := ConductQuery(dbServer, dbName, []any{sql.Named("tableName", tableName)}, "getColumnTypes.sql", deps)
	if err != nil {
		log.Println("Error fetching table column metadata: " + err.Error())
		return nil
	}
	defer rows.Close()
	var columns []ColumnMetadataRow
	for rows.Next() {
		var col ColumnMetadataRow
		if err := rows.Scan(
			&col.COLUMN_NAME,
			&col.DATA_TYPE,
			&col.CHARACTER_MAXIMUM_LENGTH,
			&col.NUMERIC_PRECISION,
			&col.NUMERIC_SCALE,
			&col.IS_NULLABLE,
		); err != nil {
			log.Println("Error scanning rows: " + err.Error())
			return nil
		}
		columns = append(columns, col)
	}
	return columns
}

func GetConductQueryFunctionFromStruct(dbServer, dbName string, queryArgs string, sqlFile string, structToUse string) {
	firstLine := "row, err := ConductQuery(" + dbServer + ", " + dbName + ", " + queryArgs + ", " + sqlFile + ")"
	nextSection := "if err != nil {\n\tlog.Println(\"Error executing query: \" + err.Error())\n\treturn nil, err\n}\ndefer rows.Close()"
	log.Println("First line of generated fuction: " + firstLine + "\nNext Section: " + nextSection)
}

func FormatStructToUse(structToUse string) string {
	/*type Deal struct {
		ID int
		DealNumber string
		Reserver float64
		VehicleYear int
		VehicleMake string
	}*/
	//pull struct name from structToUse
	structName := strings.Split(structToUse, " ")[1]
	formatedStruct := structName
	return formatedStruct
}

func GetDatabaseTables(dbServer, dbName string, deps shared.AppDependencies) []string {
	rows, err := ConductQuery(dbServer, dbName, []any{}, "getTables.sql", deps)
	if err != nil {
		log.Println("Error fetching database tables: " + err.Error())
		return nil
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var table TableMetadataRow
		if err := rows.Scan(&table.SchemaName, &table.TableName); err != nil {
			log.Println("Error scanning row: " + err.Error())
			return nil
		}
		fullTableName := table.SchemaName + "." + table.TableName
		tables = append(tables, fullTableName)
	}
	return tables
}

func ConductQuery(dbServer, dbName string, queryArgs []any, splFile string, deps shared.AppDependencies) (*sql.Rows, error) {
	connString := `Server=` + dbServer + `,1433;Database=` + dbName + `;User ID=` + `usea\` + deps.App.Preferences().String("userName") + `;Password=` + deps.App.Preferences().String("password") + `;Encrypt=disable;TrustServerCertificate=true;`
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		log.Println("Could not connect to the DB: " + err.Error())
		return nil, err
	}
	defer db.Close()
	qb, err := Queries.ReadFile("sql/" + splFile)
	if err != nil {
		log.Println("Could not read SQL file: " + err.Error())
		return nil, err
	}
	stmt, err := db.Prepare(string(qb))
	if err != nil {
		log.Println("Could not prepare the SQL statement: " + err.Error())
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(queryArgs...)
	if err != nil {
		log.Println("Error executing query: " + err.Error())
		return nil, err
	}
	return rows, nil
}
