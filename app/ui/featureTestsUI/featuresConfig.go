package featureTestsUI

var featureTestsEnv string

func SetFeatureTestsEnv(env string) {
	if env == "DEV" {
		featureTestsEnv = "DEV"
	}
	if env == "PRD" {
		featureTestsEnv = "PRD"
	}
}

func GetFeatureTestsEnv() string {
	return featureTestsEnv
}

var entryBoxStrings []string

func SetEntryBoxStrings(strings []string) {
	entryBoxStrings = strings
}

func GetEntryBoxStrings() []string {
	return entryBoxStrings
}
