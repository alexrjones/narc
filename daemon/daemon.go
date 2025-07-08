package daemon

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/alexrjones/narc"
	"github.com/alexrjones/narc/idle"
)

type ActivityKey int64

type Store interface {
	SaveActivity(ctx context.Context, name string) (ActivityKey, error)
	SavePeriod(ctx context.Context, key ActivityKey, start, end time.Time, startReason, endReason narc.ChangeReason) error
}

type Daemon struct {
	s       Store
	current current
	mu      sync.RWMutex
}

type current struct {
	activity          string
	activityKey       ActivityKey
	periodStartReason narc.ChangeReason
	periodStart       time.Time
}

func (c current) valid() bool {

	return c.activity != "" && c.activityKey != 0 && c.periodStart != time.Time{} && c.periodStartReason != 0
}

func New(s Store) (*Daemon, error) {
	d := &Daemon{
		s: s,
	}
	return d, nil
}

func (d *Daemon) SetActivity(ctx context.Context, name string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.current.valid() {
		err := d.endPeriod(ctx, narc.ChangeReasonActivityChanged)
		if err != nil {
			return err
		}
	}
	key, err := d.s.SaveActivity(ctx, name)
	if err != nil {
		return err
	}
	d.current = current{
		activity:          name,
		activityKey:       key,
		periodStartReason: narc.ChangeReasonActivityChanged,
		periodStart:       time.Now(),
	}
	return nil
}

func (d *Daemon) StopActivity(ctx context.Context, reason narc.ChangeReason) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.current.valid() {
		err := d.endPeriod(ctx, reason)
		if err != nil {
			return err
		}
	}
	d.current = current{}
	return nil
}

func (d *Daemon) Run(ctx context.Context) {

	idleCh := idle.IdleChan()
	go func() {
		for {
			select {
			case state := <-idleCh:
				{
					d.mu.Lock()
					if d.current.valid() {
						if state.Active {
							d.startPeriod(state.ChangeReason)
						} else {
							err := d.endPeriod(ctx, state.ChangeReason)
							if err != nil {
								log.Printf("Failed to end period: %s", err)
							}
						}
					}
					d.mu.Unlock()
				}
			case <-ctx.Done():
				{
					return
				}
			}
		}
	}()
}

func (d *Daemon) startPeriod(reason narc.ChangeReason) {
	d.current.periodStart = time.Now()
	d.current.periodStartReason = reason
}

func (d *Daemon) endPeriod(ctx context.Context, reason narc.ChangeReason) error {
	err := d.s.SavePeriod(ctx, d.current.activityKey, d.current.periodStart, time.Now(), d.current.periodStartReason, reason)
	if err != nil {
		return err
	}
	d.current.periodStart = time.Time{}
	d.current.periodStartReason = 0
	return nil
}
