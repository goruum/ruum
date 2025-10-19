// Package scheduler provides cron-like task scheduling functionality.
package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Task represents a scheduled task
type Task struct {
	ID       string
	Name     string
	Schedule Schedule
	Handler  func(ctx context.Context) error
	Enabled  bool
	LastRun  time.Time
	NextRun  time.Time
	RunCount int64
	Errors   []error
}

// Schedule defines when a task should run
type Schedule interface {
	Next(t time.Time) time.Time
	String() string
}

// CronSchedule represents a cron-like schedule
type CronSchedule struct {
	expression string
	interval   time.Duration
}

// Next returns the next run time
func (cs *CronSchedule) Next(t time.Time) time.Time {
	return t.Add(cs.interval)
}

// String returns the schedule expression
func (cs *CronSchedule) String() string {
	return cs.expression
}

// IntervalSchedule runs at fixed intervals
type IntervalSchedule struct {
	interval time.Duration
}

// Next returns the next run time
func (is *IntervalSchedule) Next(t time.Time) time.Time {
	return t.Add(is.interval)
}

// String returns the schedule description
func (is *IntervalSchedule) String() string {
	return fmt.Sprintf("every %s", is.interval)
}

// DailySchedule runs at a specific time each day
type DailySchedule struct {
	hour   int
	minute int
}

// Next returns the next run time
func (ds *DailySchedule) Next(t time.Time) time.Time {
	next := time.Date(t.Year(), t.Month(), t.Day(), ds.hour, ds.minute, 0, 0, t.Location())
	if next.Before(t) || next.Equal(t) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

// String returns the schedule description
func (ds *DailySchedule) String() string {
	return fmt.Sprintf("daily at %02d:%02d", ds.hour, ds.minute)
}

// WeeklySchedule runs on specific days of the week
type WeeklySchedule struct {
	weekday time.Weekday
	hour    int
	minute  int
}

// Next returns the next run time
func (ws *WeeklySchedule) Next(t time.Time) time.Time {
	daysUntil := int(ws.weekday - t.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}

	next := time.Date(t.Year(), t.Month(), t.Day()+daysUntil, ws.hour, ws.minute, 0, 0, t.Location())

	if next.Before(t) || next.Equal(t) {
		next = next.Add(7 * 24 * time.Hour)
	}

	return next
}

// String returns the schedule description
func (ws *WeeklySchedule) String() string {
	return fmt.Sprintf("weekly on %s at %02d:%02d", ws.weekday, ws.hour, ws.minute)
}

// Scheduler manages scheduled tasks
type Scheduler struct {
	tasks   map[string]*Task
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	running bool
	logger  Logger
}

// Logger interface for scheduler logging
type Logger interface {
	Info(message string, fields map[string]interface{})
	Error(message string, fields map[string]interface{})
	Debug(message string, fields map[string]interface{})
}

// NewScheduler creates a new scheduler
func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		tasks:   make(map[string]*Task),
		ctx:     ctx,
		cancel:  cancel,
		running: false,
	}
}

// SetLogger sets the logger
func (s *Scheduler) SetLogger(logger Logger) {
	s.logger = logger
}

// AddTask adds a new task to the scheduler
func (s *Scheduler) AddTask(id, name string, schedule Schedule, handler func(ctx context.Context) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; exists {
		return fmt.Errorf("task with id %s already exists", id)
	}

	task := &Task{
		ID:       id,
		Name:     name,
		Schedule: schedule,
		Handler:  handler,
		Enabled:  true,
		NextRun:  schedule.Next(time.Now()),
		Errors:   make([]error, 0),
	}

	s.tasks[id] = task

	if s.logger != nil {
		s.logger.Info("Task registered", map[string]interface{}{
			"id":       id,
			"name":     name,
			"schedule": schedule.String(),
			"nextRun":  task.NextRun,
		})
	}

	return nil
}

// RemoveTask removes a task from the scheduler
func (s *Scheduler) RemoveTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return fmt.Errorf("task with id %s not found", id)
	}

	delete(s.tasks, id)

	if s.logger != nil {
		s.logger.Info("Task removed", map[string]interface{}{
			"id": id,
		})
	}

	return nil
}

// EnableTask enables a task
func (s *Scheduler) EnableTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return fmt.Errorf("task with id %s not found", id)
	}

	task.Enabled = true
	return nil
}

// DisableTask disables a task
func (s *Scheduler) DisableTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return fmt.Errorf("task with id %s not found", id)
	}

	task.Enabled = false
	return nil
}

// GetTask returns a task by id
func (s *Scheduler) GetTask(id string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task with id %s not found", id)
	}

	return task, nil
}

// GetTasks returns all tasks
func (s *Scheduler) GetTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.running = true
	s.mu.Unlock()

	if s.logger != nil {
		s.logger.Info("Scheduler started", nil)
	}

	go s.run()
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	s.cancel()

	if s.logger != nil {
		s.logger.Info("Scheduler stopped", nil)
	}
}

func (s *Scheduler) run() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case now := <-ticker.C:
			s.checkAndRunTasks(now)
		}
	}
}

func (s *Scheduler) checkAndRunTasks(now time.Time) {
	s.mu.RLock()
	tasks := make([]*Task, 0)
	for _, task := range s.tasks {
		if task.Enabled && (now.After(task.NextRun) || now.Equal(task.NextRun)) {
			tasks = append(tasks, task)
		}
	}
	s.mu.RUnlock()

	for _, task := range tasks {
		go s.runTask(task)
	}
}

func (s *Scheduler) runTask(task *Task) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in task %s: %v", task.ID, r)
			s.handleTaskError(task, err)
		}
	}()

	if s.logger != nil {
		s.logger.Debug("Running task", map[string]interface{}{
			"id":   task.ID,
			"name": task.Name,
		})
	}

	task.LastRun = time.Now()
	task.RunCount++

	err := task.Handler(s.ctx)

	if err != nil {
		s.handleTaskError(task, err)
	} else if s.logger != nil {
		s.logger.Debug("Task completed successfully", map[string]interface{}{
			"id":   task.ID,
			"name": task.Name,
		})
	}

	// Schedule next run
	s.mu.Lock()
	task.NextRun = task.Schedule.Next(time.Now())
	s.mu.Unlock()
}

func (s *Scheduler) handleTaskError(task *Task, err error) {
	task.Errors = append(task.Errors, err)

	// Keep only last 10 errors
	if len(task.Errors) > 10 {
		task.Errors = task.Errors[len(task.Errors)-10:]
	}

	if s.logger != nil {
		s.logger.Error("Task error", map[string]interface{}{
			"id":    task.ID,
			"name":  task.Name,
			"error": err.Error(),
		})
	}
}

// Schedule builders

// Every creates an interval schedule
func Every(duration time.Duration) Schedule {
	return &IntervalSchedule{interval: duration}
}

// Daily creates a daily schedule
func Daily(hour, minute int) Schedule {
	return &DailySchedule{hour: hour, minute: minute}
}

// Weekly creates a weekly schedule
func Weekly(weekday time.Weekday, hour, minute int) Schedule {
	return &WeeklySchedule{weekday: weekday, hour: hour, minute: minute}
}

// Cron creates a cron schedule (simplified version supporting only intervals)
func Cron(interval time.Duration) Schedule {
	return &CronSchedule{
		expression: fmt.Sprintf("*/%s", interval),
		interval:   interval,
	}
}

// DefaultScheduler is the default global scheduler
var DefaultScheduler = NewScheduler()

// Global convenience functions

// AddTask adds a task to the default scheduler
func AddTask(id, name string, schedule Schedule, handler func(ctx context.Context) error) error {
	return DefaultScheduler.AddTask(id, name, schedule, handler)
}

// Start starts the default scheduler
func Start() error {
	return DefaultScheduler.Start()
}

// Stop stops the default scheduler
func Stop() {
	DefaultScheduler.Stop()
}
