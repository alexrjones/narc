package narc

import "fmt"

type (
	Activity struct {
		Name    string
		Periods []Period
	}

	Period struct {
		Start       int64
		End         int64
		StartReason ChangeReason
		EndReason   ChangeReason
	}

	IdleState struct {
		Active       bool
		ChangeReason ChangeReason
	}

	ChangeReason uint32
)

func (i IdleState) String() string {
	return fmt.Sprintf("Active: %t, ChangeReason: %s", i.Active, i.ChangeReason)
}

const (
	ChangeReasonSystemAwake     ChangeReason = 1
	ChangeReasonSystemSleep     ChangeReason = 2
	ChangeReasonUserActive      ChangeReason = 3
	ChangeReasonUserIdle        ChangeReason = 4
	ChangeReasonActivityChanged ChangeReason = 5
	ChangeReasonExplicitStop    ChangeReason = 6
	ChangeReasonDaemonExit      ChangeReason = 7
)

func (cr ChangeReason) String() string {
	switch cr {
	case ChangeReasonSystemAwake:
		return "System Awake"
	case ChangeReasonSystemSleep:
		return "System Sleep"
	case ChangeReasonUserActive:
		return "User Active"
	case ChangeReasonUserIdle:
		return "User Idle"
	case ChangeReasonActivityChanged:
		return "Activity Changed"
	case ChangeReasonExplicitStop:
		return "Explicit Stop"
	case ChangeReasonDaemonExit:
		return "Daemon Exit"
	}
	return fmt.Sprintf("(unknown change reason %d)", cr)
}
