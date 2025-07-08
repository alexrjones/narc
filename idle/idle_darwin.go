//go:build darwin

package idle

import (
	"sync"
	"time"

	"github.com/alexrjones/narc"
)

const idleThresh = 300

func IdleChan() <-chan narc.IdleState {

	ch := make(chan narc.IdleState, 1)
	var activeMu sync.Mutex
	active := true
	StartSleepWatcher(func(b bool) {
		activeMu.Lock()
		active = b
		activeMu.Unlock()
		if b {
			ch <- narc.IdleState{Active: true, ChangeReason: narc.ChangeReasonSystemAwake}
		} else {
			ch <- narc.IdleState{Active: false, ChangeReason: narc.ChangeReasonSystemSleep}
		}
	})
	pollInterval := time.Second * 10
	go func() {
		for {
			select {
			case <-time.After(pollInterval):
				{
					idleSeconds := getIdleSeconds()
					var oldState bool
					activeMu.Lock()
					oldState = active
					active = idleSeconds < idleThresh
					activeMu.Unlock()
					if idleSeconds >= idleThresh {
						ch <- narc.IdleState{Active: false, ChangeReason: narc.ChangeReasonUserIdle}
					} else if oldState == false {
						ch <- narc.IdleState{Active: true, ChangeReason: narc.ChangeReasonUserActive}
					}
				}
			}
		}
	}()
	return ch
}
