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
	handle       io.Writer
	csvw         *csv.Writer
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
	if !c.wroteHeaders {
		c.wroteHeaders = true
		c.csvw.Write([]string{"activity", "start", "end", "startReason", "endReason"})
	}
	c.csvw.Write([]string{c.activities[int64(key)], start.Format(time.RFC3339), end.Format(time.RFC3339), startReason.String(), endReason.String()})
	c.csvw.Flush()
	return nil
}

func (c *CSV) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.csvw.Flush()
	if wc, ok := c.handle.(io.WriteCloser); ok {
		wc.Close()
	}
}

func NewCSVStore(h io.Writer) *CSV {

	csvw := csv.NewWriter(h)
	return &CSV{h, csvw, 0, false, map[int64]string{}, sync.Mutex{}}
}
