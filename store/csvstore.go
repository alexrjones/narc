package store

import (
	"context"
	"encoding/csv"
	"io"
	"sync"
	"time"

	"github.com/alexrjones/narc"
	"github.com/alexrjones/narc/daemon"
)

type CSV struct {
	handle       io.ReadWriteSeeker
	seq          int64
	wroteHeaders bool
	activities   map[int64]string
	mu           sync.Mutex
}

func (c *CSV) SaveActivity(ctx context.Context, name string) (daemon.ActivityKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seq++
	key := c.seq
	c.activities[key] = name
	return daemon.ActivityKey(key), nil
}

func (c *CSV) SavePeriod(ctx context.Context, key daemon.ActivityKey, start, end time.Time, startReason, endReason narc.ChangeReason) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	//if !c.wroteHeaders {
	//	c.wroteHeaders = true
	//	c.csvw.Write([]string{"activity", "start", "end", "startReason", "endReason"})
	//}
	csvw := csv.NewWriter(c.handle)
	csvw.Write([]string{c.activities[int64(key)], start.Format(time.RFC3339), end.Format(time.RFC3339), endReason.String()})
	csvw.Flush()
	return nil
}

type activityRow struct {
	name string
	narc.Period
}

func unmarshalActivityRow(record []string) (activityRow, error) {

	name, startStr, endStr, endReasonStr := record[0], record[1], record[2], record[3]
	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		return activityRow{}, err
	}
	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		return activityRow{}, err
	}
	cr := new(narc.ChangeReason)
	err = cr.UnmarshalText([]byte(endReasonStr))
	if err != nil {
		return activityRow{}, err
	}
	return activityRow{
		name: name,
		Period: narc.Period{
			Start:     start,
			End:       end,
			EndReason: *cr,
		},
	}, nil
}

func (c *CSV) GetActivities(ctx context.Context, start, end time.Time) ([]narc.Activity, error) {
	c.mu.Lock()
	c.mu.Unlock()
	c.handle.Seek(0, io.SeekStart)
	defer c.handle.Seek(0, io.SeekEnd)
	records, err := csv.NewReader(c.handle).ReadAll()
	if err != nil {
		return nil, err
	}
	ret := make([]narc.Activity, 0, len(records)/2)
	var lastActivity narc.Activity
	for _, record := range records {
		row, err := unmarshalActivityRow(record)
		if err != nil {
			return nil, err
		}
		if (!start.IsZero() && row.Start.Before(start)) || (!end.IsZero() && row.Start.After(end)) {
			continue
		}
		if row.name != lastActivity.Name {
			if lastActivity.Name != "" {
				ret = append(ret, lastActivity)
			}
			lastActivity = narc.Activity{Name: row.name}
		}
		lastActivity.Periods = append(lastActivity.Periods, row.Period)
	}
	if lastActivity.Name != "" {
		ret = append(ret, lastActivity)
	}
	return ret, nil
}

func (c *CSV) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if wc, ok := c.handle.(io.Closer); ok {
		wc.Close()
	}
}

func NewCSVStore(h io.ReadWriteSeeker) *CSV {

	return &CSV{h, 0, false, map[int64]string{}, sync.Mutex{}}
}
