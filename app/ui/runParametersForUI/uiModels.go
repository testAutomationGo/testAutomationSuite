package runParametersForUI

type AppConfiguration struct {
	Envs                  []string `json:"environments"`
	Options               []string `json:"envOptions"`
	TestSectorNames       []string `json:"testSectorsNames"`
	TestSectorCamelCase   []string `json:"testSectorCamelCase"`
	EnvConfigurationFiles []string `json:"envConfigurationFiles"`
}

type AccountsRoot struct {
	AccountData []Accounts `json:"accountData"`
}

type Accounts struct {
	AccountInt string `json:"AccountInt"`
	FirstName  string `json:"FirstName"`
	LastName   string `json:"LastName"`
	Email      string `json:"Email"`
	Password   string `json:"Password"`
	Jwt        string `json:"Jwt"`
	ApiKeyID   string `json:"ApiKeyID"`
	AccountID  string `json:"AccountId"`
	Gateway    string `json:"Gateway"`
	GatewayID  string `json:"GatewayID"`
}

type UIArgsRoot struct {
	UIArgs UIArgs `json:"uiArgs"`
}

type UIArgs struct {
	GmailCreds                string          `json:"gmailCreds"`
	DatabaseConnectionStrings []DBConnections `json:"databaseConnectionStrings"`
}

type DBConnections struct {
	DatabaseName     string     `json:"dbName"`
	ConnectionString string     `json:"connectionString"`
	DBTables         []DBTables `json:"dbTables"`
}

type DBTables struct {
	TableName string      `json:"dbTableName"`
	Columns   []DBColumns `json:"dbColumns"`
}

type DBColumns struct {
	ColumnName string `json:"columnName"`
	ColumnType string `json:"columnType"`
}
