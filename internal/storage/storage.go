package storage

import (
	"encoding/json"
	"os"
	"sync"

	"servicehealthchecker/internal/models"
)

type Storage struct {
	mu           sync.RWMutex
	data         models.StorageData
	pending      models.PendingTasks
	dataPath     string
	pendingPath  string
}

func New(dataPath, pendingPath string) (*Storage, error) {
	s := &Storage{
		dataPath:    dataPath,
		pendingPath: pendingPath,
	}

	if err := s.loadData(); err != nil {
		return nil, err
	}

	if err := s.loadPending(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Storage) loadData() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.dataPath)
	if os.IsNotExist(err) {
		s.data = models.StorageData{
			LastID:   0,
			LinkSets: []models.LinkSet{},
		}
		return nil
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.data)
}

func (s *Storage) loadPending() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.pendingPath)
	if os.IsNotExist(err) {
		s.pending = models.PendingTasks{Tasks: []models.PendingTask{}}
		return nil
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.pending)
}

func (s *Storage) saveData() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.dataPath, data, 0644)
}

func (s *Storage) savePending() error {
	data, err := json.MarshalIndent(s.pending, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.pendingPath, data, 0644)
}

func (s *Storage) CreateLinkSet(links []string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.LastID++
	id := s.data.LastID

	linkStatuses := make([]models.LinkStatus, len(links))
	for i, link := range links {
		linkStatuses[i] = models.LinkStatus{
			URL:    link,
			Status: "pending",
		}
	}

	linkSet := models.LinkSet{
		ID:      id,
		Links:   linkStatuses,
		Checked: false,
	}

	s.data.LinkSets = append(s.data.LinkSets, linkSet)

	if err := s.saveData(); err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Storage) UpdateLinkSet(id int, statuses map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.data.LinkSets {
		if s.data.LinkSets[i].ID == id {
			for j := range s.data.LinkSets[i].Links {
				url := s.data.LinkSets[i].Links[j].URL
				if status, ok := statuses[url]; ok {
					s.data.LinkSets[i].Links[j].Status = status
				}
			}
			s.data.LinkSets[i].Checked = true
			break
		}
	}

	return s.saveData()
}

func (s *Storage) GetLinkSets(ids []int) []models.LinkSet {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idSet := make(map[int]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	var result []models.LinkSet
	for _, ls := range s.data.LinkSets {
		if idSet[ls.ID] && ls.Checked {
			result = append(result, ls)
		}
	}

	return result
}

func (s *Storage) AddPendingTask(id int, links []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pending.Tasks = append(s.pending.Tasks, models.PendingTask{
		ID:    id,
		Links: links,
	})

	return s.savePending()
}

func (s *Storage) GetPendingTasks() []models.PendingTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]models.PendingTask, len(s.pending.Tasks))
	copy(tasks, s.pending.Tasks)
	return tasks
}

func (s *Storage) RemovePendingTask(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, task := range s.pending.Tasks {
		if task.ID == id {
			s.pending.Tasks = append(s.pending.Tasks[:i], s.pending.Tasks[i+1:]...)
			break
		}
	}

	return s.savePending()
}

func (s *Storage) ClearPendingTasks() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pending.Tasks = []models.PendingTask{}
	return s.savePending()
}

