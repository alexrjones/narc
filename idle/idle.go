package idle

import (
	"sync"
	"time"

	"github.com/alexrjones/narc"
)

const idleThresh = 300

func IdleChan() <-chan narc.IdleState {

	ch := make(chan narc.IdleState, 1)
	var awakeMu sync.Mutex
	systemAwake := true
	go func() {
		StartSleepWatcher(func(b bool) {
			awakeMu.Lock()
			systemAwake = b
			awakeMu.Unlock()
			if b {
				ch <- narc.IdleState{Active: true, ChangeReason: narc.ChangeReasonSystemAwake}
			} else {
				ch <- narc.IdleState{Active: false, ChangeReason: narc.ChangeReasonSystemSleep}
			}
		})
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
						return
					}
					awakeMu.Unlock()
					idleSeconds := getIdleSeconds()
					oldState := userActive
					userActive = idleSeconds < idleThresh
					if oldState != userActive {
						if userActive {
							ch <- narc.IdleState{Active: true, ChangeReason: narc.ChangeReasonUserActive}
						} else {
							ch <- narc.IdleState{Active: false, ChangeReason: narc.ChangeReasonUserIdle}
						}
					}
				}
			}
		}
	}()
	return ch
}
