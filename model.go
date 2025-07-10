package narc

import (
	"fmt"
	"time"
)

type (
	Activity struct {
		Name    string
		Periods []Period
	}
	Activities []Activity

	Period struct {
		Start     time.Time
		End       time.Time
		EndReason ChangeReason
	}

	DurationRow struct {
		Date     time.Time
		Name     string
		Duration time.Duration
	}

	IdleState struct {
		Active       bool
		ChangeReason ChangeReason
	}

	ChangeReason uint32
)

func (acts Activities) ToDurationRows() []DurationRow {

	rows := make([]DurationRow, 0, len(acts))
	for _, a := range acts {
		rows = append(rows, a.ToDurationRows()...)
	}
	//slices.SortFunc(rows, func(a, b DurationRow) int {
	//	if a.Date.Before(b.Date) {
	//		return -1
	//	} else if b.Date.Before(a.Date) {
	//		return 1
	//	}
	//	return 0
	//})
	return rows
}

func (a Activity) ToDurationRows() []DurationRow {

	dates := make(map[time.Time]time.Duration)
	for _, p := range a.Periods {
		t := p.Start
		date := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		dates[date] += p.End.Sub(p.Start)
	}
	ret := make([]DurationRow, 0, len(dates))
	for date, dur := range dates {
		ret = append(ret, DurationRow{
			Date:     date,
			Name:     a.Name,
			Duration: dur,
		})
	}
	return ret
}

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

func (cr *ChangeReason) UnmarshalText(text []byte) error {
	switch string(text) {
	case "System Awake":
		*cr = ChangeReasonSystemAwake
	case "System Sleep":
		*cr = ChangeReasonSystemSleep
	case "User Active":
		*cr = ChangeReasonUserActive
	case "User Idle":
		*cr = ChangeReasonUserIdle
	case "Activity Changed":
		*cr = ChangeReasonActivityChanged
	case "Explicit Stop":
		*cr = ChangeReasonExplicitStop
	case "Daemon Exit":
		*cr = ChangeReasonDaemonExit
	default:
		return fmt.Errorf("unrecognized ChangeReason: %q", text)
	}
	return nil
}
