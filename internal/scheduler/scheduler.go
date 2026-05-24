package scheduler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/sunshow/siphongear/internal/events"
	"github.com/sunshow/siphongear/internal/runner"
	"github.com/sunshow/siphongear/internal/store/models"
)

type Scheduler struct {
	db      *gorm.DB
	runner  *runner.Runner
	bus     *events.Bus
	cron    *cron.Cron
	mu      sync.Mutex
	entries map[uint]cron.EntryID
	subs    map[string]map[uint]bool // event topic -> set of collector ids
}

func New(db *gorm.DB, r *runner.Runner, bus *events.Bus) *Scheduler {
	c := cron.New(cron.WithSeconds())
	return &Scheduler{
		db:      db,
		runner:  r,
		bus:     bus,
		cron:    c,
		entries: map[uint]cron.EntryID{},
		subs:    map[string]map[uint]bool{},
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	if err := s.loadAll(); err != nil {
		return err
	}
	s.cron.Start()
	go s.listenEvents(ctx)
	return nil
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
}

// Reload re-syncs scheduler state with DB; safe to call after CRUD on collectors.
func (s *Scheduler) Reload() error { return s.loadAll() }

func (s *Scheduler) loadAll() error {
	var collectors []models.Collector
	if err := s.db.Where("enabled = ?", true).Find(&collectors).Error; err != nil {
		return err
	}
	s.mu.Lock()
	for id, eid := range s.entries {
		s.cron.Remove(eid)
		delete(s.entries, id)
	}
	s.subs = map[string]map[uint]bool{}
	s.mu.Unlock()

	for _, c := range collectors {
		if err := s.add(c); err != nil {
			log.Warn().Err(err).Uint("collector", c.ID).Msg("schedule add failed")
		}
	}
	return nil
}

func (s *Scheduler) add(c models.Collector) error {
	switch c.ScheduleType {
	case "interval":
		spec, err := intervalToCron(c.ScheduleSpec)
		if err != nil {
			return err
		}
		return s.addCron(c.ID, spec)
	case "cron":
		spec := normalizeCron(c.ScheduleSpec)
		return s.addCron(c.ID, spec)
	case "event":
		topic := c.ScheduleSpec
		if topic == "" {
			return fmt.Errorf("event topic empty")
		}
		s.mu.Lock()
		if _, ok := s.subs[topic]; !ok {
			s.subs[topic] = map[uint]bool{}
		}
		s.subs[topic][c.ID] = true
		s.mu.Unlock()
		return nil
	}
	return nil
}

func (s *Scheduler) addCron(id uint, spec string) error {
	eid, err := s.cron.AddFunc(spec, func() {
		_, err := s.runner.Trigger(context.Background(), id, "schedule", nil, false)
		if err != nil {
			log.Warn().Err(err).Uint("collector", id).Msg("scheduled run failed")
		}
	})
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.entries[id] = eid
	s.mu.Unlock()
	return nil
}

func (s *Scheduler) listenEvents(ctx context.Context) {
	ch := s.bus.Subscribe("run.completed", 64)
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			cid, _ := ev.Payload["collector_id"].(uint)
			topic := fmt.Sprintf("collector.%d.completed", cid)
			s.mu.Lock()
			subs := append([]uint(nil), keysOf(s.subs[topic])...)
			s.mu.Unlock()
			for _, target := range subs {
				go func(id uint) {
					_, err := s.runner.Trigger(context.Background(), id, "event", nil, false)
					if err != nil {
						log.Warn().Err(err).Uint("collector", id).Msg("event-triggered run failed")
					}
				}(target)
			}
		}
	}
}

func keysOf(m map[uint]bool) []uint {
	out := make([]uint, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// intervalToCron converts "5m" / "30s" / "1h" / "10" to a cron-with-seconds spec.
func intervalToCron(spec string) (string, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", fmt.Errorf("empty interval")
	}
	d, err := time.ParseDuration(spec)
	if err != nil {
		return "", fmt.Errorf("parse interval %q: %w", spec, err)
	}
	if d < time.Second {
		return "", fmt.Errorf("interval too short")
	}
	return fmt.Sprintf("@every %s", d), nil
}

// normalizeCron accepts either 5-field (no seconds) or 6-field cron and normalizes to 6-field.
func normalizeCron(spec string) string {
	spec = strings.TrimSpace(spec)
	if strings.HasPrefix(spec, "@") {
		return spec
	}
	parts := strings.Fields(spec)
	if len(parts) == 5 {
		return "0 " + spec
	}
	return spec
}
