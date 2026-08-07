package alarm

import (
	"sync"
)

type AlarmRepository interface {
	Save(alarm Alarm) error
	GetAll() ([]Alarm, error)
}

type MemoryAlarmRepository struct {
	mu     sync.RWMutex
	alarms []Alarm
}

func NewMemoryAlarmRepository() *MemoryAlarmRepository {
	return &MemoryAlarmRepository{
		alarms: make([]Alarm, 0),
	}
}

func (r *MemoryAlarmRepository) Save(alarm Alarm) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.alarms = append(r.alarms, alarm)
	return nil
}

func (r *MemoryAlarmRepository) GetAll() ([]Alarm, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	copied := make([]Alarm, len(r.alarms))
	copy(copied, r.alarms)
	return copied, nil
}
