package store

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	// obviously use sqlite dialect for SQLite store
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/matoous/mailback/internal/models"
)

var ErrNotFound = errors.New("record not found")

// SQLiteStore is SQLite backed storage.
type SQLiteStore struct {
	db *gorm.DB
}

// NewSQLiteStore creates new SQLite store using file with given filename as the persistent storage.
func NewSQLiteStore(filename string) (*SQLiteStore, error) {
	db, err := gorm.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	return &SQLiteStore{db}, nil
}

func (s *SQLiteStore) Migrate() error {
	return s.db.AutoMigrate(&models.Entry{}).Error
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) Save(e *models.Entry) error {
	return s.db.Save(e).Error
}

func (s *SQLiteStore) Update(e *models.Entry) error {
	return s.db.Save(e).Error
}

func (s *SQLiteStore) Delete(e *models.Entry) error {
	res := s.db.Delete(e)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLiteStore) PendingEntries() ([]models.Entry, error) {
	var entries []models.Entry
	err := s.db.Where("scheduled_for < ?", time.Now()).Find(&entries).Error
	if err != nil {
		return nil, err
	}
	return entries, nil
}
