package databaseCache

import "time"

type PersistentTableCache struct {
	Version   int                         `json:"version"`
	UpdatedAt time.Time                   `json:"updatedAt"`
	Entries   []PersistentTableCacheEntry `json:"entries"`
}

type PersistentTableCacheEntry struct {
	ServerName   string    `json:"serverName"`
	DatabaseName string    `json:"databaseName"`
	Tables       []string  `json:"tables"`
	LastLoaded   time.Time `json:"lastLoaded"`
}

type tableCacheEntry struct {
	tables     []string
	lastLoaded time.Time
}

type serverDatabasePair struct {
	serverName string
	dbName     string
}
