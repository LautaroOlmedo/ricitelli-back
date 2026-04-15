package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Metadata describes a stored report file.
type Metadata struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Filename    string    `json:"filename"`
	Path        string    `json:"path"`
	GeneratedAt time.Time `json:"generated_at"`
	FileSize    int64     `json:"file_size"`
	FromDate    string    `json:"from_date,omitempty"`
	ToDate      string    `json:"to_date,omitempty"`
}

// Store persists PDF reports in the filesystem and maintains an in-memory index.
type Store struct {
	baseDir string
	mu      sync.RWMutex
	index   map[string]*Metadata // id -> metadata
}

// NewStore creates a Store rooted at baseDir, loading any existing index.json.
func NewStore(baseDir string) (*Store, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}
	s := &Store{baseDir: baseDir, index: make(map[string]*Metadata)}
	_ = s.loadIndex()
	return s, nil
}

func (s *Store) indexPath() string {
	return filepath.Join(s.baseDir, "index.json")
}

func (s *Store) loadIndex() error {
	data, err := os.ReadFile(s.indexPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var items []*Metadata
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	for _, m := range items {
		s.index[m.ID] = m
	}
	return nil
}

func (s *Store) saveIndexLocked() error {
	items := make([]*Metadata, 0, len(s.index))
	for _, m := range s.index {
		items = append(items, m)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].GeneratedAt.After(items[j].GeneratedAt)
	})
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.indexPath(), data, 0o644)
}

// Save writes the given PDF bytes to disk and returns its metadata.
// The file is placed at baseDir/YYYY-MM-DD/report-{type}-{uuid}.pdf
func (s *Store) Save(pdfBytes []byte, reportType string, fromDate, toDate string) (*Metadata, error) {
	now := time.Now().UTC()
	id := uuid.New().String()
	dayDir := filepath.Join(s.baseDir, now.Format("2006-01-02"))
	if err := os.MkdirAll(dayDir, 0o755); err != nil {
		return nil, fmt.Errorf("create day dir: %w", err)
	}
	filename := fmt.Sprintf("report-%s-%s.pdf", reportType, id[:8])
	fullPath := filepath.Join(dayDir, filename)
	if err := os.WriteFile(fullPath, pdfBytes, 0o644); err != nil {
		return nil, fmt.Errorf("write pdf: %w", err)
	}
	m := &Metadata{
		ID:          id,
		Type:        reportType,
		Filename:    filename,
		Path:        fullPath,
		GeneratedAt: now,
		FileSize:    int64(len(pdfBytes)),
		FromDate:    fromDate,
		ToDate:      toDate,
	}
	s.mu.Lock()
	s.index[id] = m
	err := s.saveIndexLocked()
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Get returns metadata by ID.
func (s *Store) Get(id string) (*Metadata, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.index[id]
	return m, ok
}

// List returns all metadata entries, optionally filtered by type, sorted by GeneratedAt desc.
func (s *Store) List(typeFilter string) []*Metadata {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Metadata, 0, len(s.index))
	for _, m := range s.index {
		if typeFilter != "" && m.Type != typeFilter {
			continue
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].GeneratedAt.After(out[j].GeneratedAt)
	})
	return out
}

// CleanupOlderThan removes reports older than the given duration.
func (s *Store) CleanupOlderThan(retention time.Duration) (int, error) {
	cutoff := time.Now().Add(-retention)
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := 0
	for id, m := range s.index {
		if m.GeneratedAt.Before(cutoff) {
			_ = os.Remove(m.Path)
			delete(s.index, id)
			removed++
		}
	}
	if removed > 0 {
		_ = s.saveIndexLocked()
	}
	return removed, nil
}
