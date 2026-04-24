package store

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// JobRecord holds the last known execution info for a cron job.
type JobRecord struct {
	Name      string    `json:"name"`
	LastRun   time.Time `json:"last_run"`
	LastExit  int       `json:"last_exit"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Store is a simple file-backed key/value store for job records.
type Store struct {
	mu      sync.RWMutex
	path    string
	records map[string]JobRecord
}

// New opens (or creates) the store at the given file path.
func New(path string) (*Store, error) {
	s := &Store{
		path:    path,
		records: make(map[string]JobRecord),
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

// Get returns the record for a job by name.
func (s *Store) Get(name string) (JobRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.records[name]
	return r, ok
}

// Set writes or updates a job record and persists the store.
func (s *Store) Set(r JobRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r.UpdatedAt = time.Now().UTC()
	s.records[r.Name] = r
	return s.save()
}

// All returns a copy of all stored records.
func (s *Store) All() []JobRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]JobRecord, 0, len(s.records))
	for _, r := range s.records {
		out = append(out, r)
	}
	return out
}

func (s *Store) load() error {
	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&s.records)
}

func (s *Store) save() error {
	tmp := s.path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(s.records); err != nil {
		f.Close()
		return err
	}
	f.Close()
	return os.Rename(tmp, s.path)
}
