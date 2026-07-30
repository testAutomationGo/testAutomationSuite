package databaseCache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testAutomationSuiteGO/app/ui/runParametersForUI"
	"time"
)

var servers []string
var databaseNames [][]string

var tableCache map[string]tableCacheEntry
var cacheMutex sync.RWMutex

const tableCacheJSONName = "databaseCache.json"

func SetServers() {
	baseServers := runParametersForUI.GetDatabaseServers()
	servers = append([]string(nil), baseServers...)
}

func SetDatabases() {
	baseDatabases := runParametersForUI.GetDatabaseNames()
	databaseNames = make([][]string, len(databaseNames))
	for i := range baseDatabases {
		databaseNames[i] = append([]string(nil), baseDatabases[i]...)
	}
}

func SetTables() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	tableCache = make(map[string]tableCacheEntry, len(servers))
}

func Initialize() {
	SetServers()
	SetDatabases()
	SetTables()
	loadTableCacheFromDisk()
}

func GetServers() []string {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	return append([]string(nil), servers...)
}

func GetDatabases() [][]string {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	copied := make([][]string, len(databaseNames))
	for i := range databaseNames {
		copied[i] = append([]string(nil), databaseNames[i]...)
	}
	return copied
}

func InvalidateTables(serverName, dbName string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	delete(tableCache, buildTableCacheKey(serverName, dbName))
	persistTableCacheToDiskLocked()
}

func InvalidateAllTables() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	tableCache = make(map[string]tableCacheEntry, len(servers))
	persistTableCacheToDiskLocked()
}

func GetOrLoadTables(serverName, dbName string, forceRefresh bool, loadTables func() []string) []string {
	if loadTables == nil || !isKnownServerDatabase(serverName, dbName) {
		return []string{}
	}
	cacheKey := buildTableCacheKey(serverName, dbName)
	now := time.Now()
	cacheMutex.RLock()
	entry, found := tableCache[cacheKey]
	cacheMutex.RUnlock()
	if found && !forceRefresh {
		return append([]string(nil), entry.tables...)
	}
	tables := loadTables()
	if tables == nil {
		tables = []string{}
	}
	cacheMutex.Lock()
	tableCache[cacheKey] = tableCacheEntry{
		tables:     append([]string(nil), tables...),
		lastLoaded: now,
	}
	persistTableCacheToDiskLocked()
	cacheMutex.Unlock()
	return append([]string(nil), tables...)
}

func PrewarmKnownTables(forceRefresh bool, loadTables func(serverName, dbName string) []string) int {
	if loadTables == nil {
		return 0
	}
	pairs := getKnownServerDatabasePairs()
	if len(pairs) == 0 {
		return 0
	}
	now := time.Now()
	updatedCount := 0

	for _, pair := range pairs {
		cachedKey := buildTableCacheKey(pair.serverName, pair.dbName)
		if !forceRefresh {
			cacheMutex.RLock()
			_, found := tableCache[cachedKey]
			cacheMutex.RUnlock()
			if found {
				continue
			}
		}
		tables := loadTables(pair.serverName, pair.dbName)
		if tables == nil {
			tables = []string{}
		}
		cacheMutex.Lock()
		tableCache[cachedKey] = tableCacheEntry{
			tables:     append([]string(nil), tables...),
			lastLoaded: now,
		}
		cacheMutex.Unlock()
		updatedCount++
	}
	if updatedCount > 0 {
		cacheMutex.Lock()
		persistTableCacheToDiskLocked()
		cacheMutex.Unlock()
	}
	return updatedCount
}

func ClearTableCacheFile() error {
	cacheMutex.Lock()
	tableCache = make(map[string]tableCacheEntry, len(servers))
	cacheMutex.Unlock()
	cacheFilePath := getTableCacheFilePath()
	if err := os.Remove(cacheFilePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func buildTableCacheKey(serverName, dbName string) string {
	return strings.ToLower(strings.TrimSpace(serverName)) + "|" + strings.ToLower(strings.TrimSpace(dbName))
}

func isKnownServerDatabase(serverName string, dbName string) bool {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	for i := range servers {
		if !strings.EqualFold(strings.TrimSpace(servers[i]), strings.TrimSpace(serverName)) {
			continue
		}
		if i >= len(databaseNames) {
			return false
		}
		for _, knownDB := range databaseNames[i] {
			if strings.EqualFold(strings.TrimSpace(knownDB), strings.TrimSpace(dbName)) {
				return true
			}
		}
		return false
	}
	return false
}

func getTableCacheFilePath() string {
	_, currentFilePath, _, ok := runtime.Caller(0)
	if ok {
		return filepath.Join(filepath.Dir(currentFilePath), tableCacheJSONName)
	}
	return filepath.Join("app", "ui", "code", "codeGen", "codeGenDatabase", "databaseCache", tableCacheJSONName)
}
func loadTableCacheFromDisk() {
	cacheFilePath := getTableCacheFilePath()
	cacheBytes, err := os.ReadFile(cacheFilePath)
	if err != nil {
		return
	}
	var persistent PersistentTableCache
	if err := json.Unmarshal(cacheBytes, &persistent); err != nil {
		return
	}
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	for _, entry := range persistent.Entries {
		if !isKnownServerDatabaseNoLock(entry.ServerName, entry.DatabaseName) {
			continue
		}
		tableCache[buildTableCacheKey(entry.ServerName, entry.DatabaseName)] = tableCacheEntry{
			tables:     append([]string(nil), entry.Tables...),
			lastLoaded: entry.LastLoaded,
		}
	}
}

func persistTableCacheToDiskLocked() {
	cacheFilePath := getTableCacheFilePath()
	cacheDir := filepath.Dir(cacheFilePath)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return
	}
	persistent := PersistentTableCache{
		Version:   1,
		UpdatedAt: time.Now(),
		Entries:   make([]PersistentTableCacheEntry, 0, len(tableCache)),
	}
	for key, entry := range tableCache {
		parts := strings.SplitN(key, "|", 2)
		if len(parts) != 2 {
			continue
		}
		persistent.Entries = append(persistent.Entries, PersistentTableCacheEntry{
			ServerName:   parts[0],
			DatabaseName: parts[1],
			Tables:       append([]string(nil), entry.tables...),
			LastLoaded:   entry.lastLoaded,
		})
	}
	cacheBytes, err := json.MarshalIndent(persistent, "", "  ")
	if err != nil {
		return
	}
	tempPath := cacheFilePath + ".tmp"
	if err := os.WriteFile(tempPath, cacheBytes, 0o644); err != nil {
		return
	}
	_ = os.Rename(tempPath, cacheFilePath)
}

func isKnownServerDatabaseNoLock(serverName, dbName string) bool {
	for i := range servers {
		if !strings.EqualFold(strings.TrimSpace(servers[i]), strings.TrimSpace(serverName)) {
			continue
		}
		if i >= len(databaseNames) {
			return false
		}
		for _, knownDB := range databaseNames[i] {
			if strings.EqualFold(strings.TrimSpace(knownDB), strings.TrimSpace(dbName)) {
				return true
			}
		}
		return false
	}
	return false
}

func getKnownServerDatabasePairs() []serverDatabasePair {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	var pairs []serverDatabasePair
	for i, serverName := range servers {
		if i >= len(databaseNames) {
			continue
		}
		for _, dbName := range databaseNames[i] {
			if strings.TrimSpace(serverName) == "" || strings.TrimSpace(dbName) == "" {
				continue
			}
			pairs = append(pairs, serverDatabasePair{serverName: serverName, dbName: dbName})
		}
	}
	return pairs
}
