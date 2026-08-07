package analysis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"goAlarmMonitoring/internal/alarm"
	"goAlarmMonitoring/internal/bus"
	"goAlarmMonitoring/pkg/types"

	"github.com/google/uuid"
)

type DeviceState struct {
	motionCount   int
	motionWindow  time.Time
	smokeCount    int
	smokeWindow   time.Time
	doorOpen      bool
	lastAlarmTime time.Time
}

type AnalysisService struct {
	subscriber  bus.EventSubscriber
	alarmSvc    *alarm.AlarmService
	cancel      context.CancelFunc
	unsubscribe func()
	wg          sync.WaitGroup

	mu           sync.Mutex
	deviceStates map[uuid.UUID]*DeviceState

	eventThreshold int
	windowDuration time.Duration
}

func NewAnalysisService(sub bus.EventSubscriber, alarmSvc *alarm.AlarmService) *AnalysisService {
	return &AnalysisService{
		subscriber:     sub,
		alarmSvc:       alarmSvc,
		deviceStates:   make(map[uuid.UUID]*DeviceState),
		eventThreshold: 3,
		windowDuration: 5 * time.Second,
	}
}

func (as *AnalysisService) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	as.cancel = cancel

	ch, unsub := as.subscriber.Subscribe()
	as.unsubscribe = unsub

	as.wg.Add(1)
	go func() {
		defer as.wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("AnalysisService stopped.")
				return
			case event, ok := <-ch:
				if !ok {
					return
				}
				if sensorEvent, ok := event.(types.SensorEvent); ok {
					as.analyze(sensorEvent)
				}
			}
		}
	}()
}

func (as *AnalysisService) Stop() {
	if as.cancel != nil {
		as.cancel()
	}
	as.wg.Wait()
	if as.unsubscribe != nil {
		as.unsubscribe()
	}
}

func (as *AnalysisService) getState(deviceID uuid.UUID) *DeviceState {
	state, exists := as.deviceStates[deviceID]
	if !exists {
		state = &DeviceState{
			motionWindow: time.Now(),
			smokeWindow:  time.Now(),
		}
		as.deviceStates[deviceID] = state
	}
	return state
}

func (as *AnalysisService) analyze(e types.SensorEvent) {
	as.mu.Lock()
	defer as.mu.Unlock()

	state := as.getState(e.DeviceID)
	const alarmCooldown = 60 * time.Second

	switch e.Type {
	case types.Motion:
		if time.Since(state.motionWindow) > as.windowDuration {
			state.motionCount = 1
			state.motionWindow = e.Time
		} else {
			state.motionCount++
		}

		if state.motionCount >= as.eventThreshold && !state.doorOpen {
			if time.Since(state.lastAlarmTime) > alarmCooldown {
				_ = as.alarmSvc.TriggerAlarm(e.DeviceID, e.Type, e.Time)
				state.lastAlarmTime = e.Time
			}
			state.motionCount = 0
		}

	case types.Door:
		state.doorOpen = !state.doorOpen
		state.motionCount = 0

	case types.Smoke:
		if time.Since(state.smokeWindow) > as.windowDuration {
			state.smokeCount = 1
			state.smokeWindow = e.Time
		} else {
			state.smokeCount++
		}

		if state.smokeCount >= as.eventThreshold {
			if time.Since(state.lastAlarmTime) > alarmCooldown {
				_ = as.alarmSvc.TriggerAlarm(e.DeviceID, e.Type, e.Time)
				state.lastAlarmTime = e.Time
			}
			state.smokeCount = 0
		}
	}
}
