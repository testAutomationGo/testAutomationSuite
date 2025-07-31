package testRunParameters

type ConfigRunnerRoot struct {
	Env Env `json:"env"`
}

type Env struct {
	Name              string `json:"name"`
	EnvInt            int    `json:"envInt"`
	EnvString         string `json:"envString"`
	EnvLowers         string `json:"envLowers"`
	EnvUppers         string `json:"envUppers"`
	ApiEndpoint       string `json:"apiEndpoint"`
	AppUrl            string `json:"appUrl"`
	MobileAppID       string `json:"mobileAppID"`
	LocalVariableFile string `json:"localVariableFile"`
}

type AccountDataRoot struct {
	AccountData []Account `json:"accountData"`
}

type Account struct {
	AccountInt string `json:"AccountInt"`
	FirstName  string `json:"FirstName"`
	LastName   string `json:"LastName"`
	Email      string `json:"Email"`
	Password   string `json:"Password"`
	Jwt        string `json:"Jwt"`
	ApiKeyID   string `json:"ApiKeyID"`
	AccountId  string `json:"AccountId"`
	Gateway    string `json:"Gateway"`
	GatewayID  string `json:"GatewayID"`
}
