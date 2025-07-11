package idle

import (
	"context"
	"sync"
	"time"

	"github.com/alexrjones/narc"
)

type Monitor struct {
	idleTimeout time.Duration
}

func NewMonitor(timeout time.Duration) *Monitor {
	return &Monitor{
		idleTimeout: timeout,
	}
}

var once sync.Once

func sleepStateChanged(b bool) {
	sleepStateChangedCb(b)
}

var sleepStateChangedCb = func(b bool) {}

func (m *Monitor) Start(ctx context.Context) <-chan narc.IdleState {
	ch := make(chan narc.IdleState, 1)
	var awakeMu sync.Mutex
	systemAwake := true
	go func() {
		once.Do(func() {
			StartSleepWatcher(sleepStateChanged)
		})
		sleepStateChangedCb = func(b bool) {
			awakeMu.Lock()
			systemAwake = b
			awakeMu.Unlock()
			if b {
				ch <- narc.IdleState{Active: true, ChangeReason: narc.ChangeReasonSystemAwake}
			} else {
				ch <- narc.IdleState{Active: false, ChangeReason: narc.ChangeReasonSystemSleep}
			}
		}
	}()
	userActive := true
	pollInterval := time.Second * 10
	go func() {
		for {
			select {
			case <-time.After(pollInterval):
				{
					awakeMu.Lock()
					if !systemAwake {
						awakeMu.Unlock()
						continue
					}
					awakeMu.Unlock()
					idleSeconds := getIdleSeconds()
					oldState := userActive
					userActive = float64(idleSeconds) < m.idleTimeout.Seconds()
					if oldState != userActive {
						if userActive {
							ch <- narc.IdleState{Active: true, ChangeReason: narc.ChangeReasonUserActive}
						} else {
							ch <- narc.IdleState{Active: false, ChangeReason: narc.ChangeReasonUserIdle}
						}
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
}
