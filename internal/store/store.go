package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

type VaultStore struct {
	db *sql.DB
}

type Manager struct {
	dataDir string
	vaults  map[int64]*VaultStore
	mu      sync.RWMutex
}

func NewManager(dataDir string) (*Manager, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	return &Manager{
		dataDir: dataDir,
		vaults:  make(map[int64]*VaultStore),
	}, nil
}

func (m *Manager) GetVault(userID int64) (*VaultStore, error) {
	m.mu.RLock()
	if v, ok := m.vaults[userID]; ok {
		m.mu.RUnlock()
		return v, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if v, ok := m.vaults[userID]; ok {
		return v, nil
	}

	vault, err := m.openVault(userID)
	if err != nil {
		return nil, err
	}
	m.vaults[userID] = vault
	return vault, nil
}

func (m *Manager) openVault(userID int64) (*VaultStore, error) {
	userDir := filepath.Join(m.dataDir, "users", fmt.Sprintf("%d", userID))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return nil, fmt.Errorf("create user dir: %w", err)
	}

	dbPath := filepath.Join(userDir, "vault.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := RunMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &VaultStore{db: db}, nil
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, v := range m.vaults {
		v.db.Close()
	}
	return nil
}

// DB returns the underlying database connection (for advanced use)
func (v *VaultStore) DB() *sql.DB {
	return v.db
}

// DataDir returns the base data directory path
func (m *Manager) DataDir() string {
	return m.dataDir
}

// UserDir returns the directory path for a specific user
func (m *Manager) UserDir(userID int64) string {
	return filepath.Join(m.dataDir, "users", fmt.Sprintf("%d", userID))
}

// GetSetting retrieves a setting value by key.
func (v *VaultStore) GetSetting(key string) (string, error) {
	var value string
	err := v.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

// SetSetting stores a setting value (upsert).
func (v *VaultStore) SetSetting(key, value string) error {
	_, err := v.db.Exec(
		"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		key, value,
	)
	return err
}
