package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewScheduler(t *testing.T) {
	s := NewScheduler()
	if s == nil {
		t.Fatal("NewScheduler() returned nil")
	}
}

func TestScheduler_AddTask(t *testing.T) {
	s := NewScheduler()

	err := s.AddTask("test", "Test Task",
		Every(time.Second),
		func(ctx context.Context) error {
			return nil
		},
	)

	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}

	task, err := s.GetTask("test")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if task.Name != "Test Task" {
		t.Errorf("Task name = %v, want %v", task.Name, "Test Task")
	}
}

func TestScheduler_AddTask_Duplicate(t *testing.T) {
	s := NewScheduler()

	_ = s.AddTask("test", "Test", Every(time.Second), func(ctx context.Context) error { return nil })
	err := s.AddTask("test", "Test2", Every(time.Second), func(ctx context.Context) error { return nil })

	if err == nil {
		t.Error("Expected error for duplicate task ID")
	}
}

func TestScheduler_RemoveTask(t *testing.T) {
	s := NewScheduler()
	_ = s.AddTask("test", "Test", Every(time.Second), func(ctx context.Context) error { return nil })

	err := s.RemoveTask("test")
	if err != nil {
		t.Fatalf("RemoveTask() error = %v", err)
	}

	_, err = s.GetTask("test")
	if err == nil {
		t.Error("Task should not exist after removal")
	}
}

func TestScheduler_EnableDisable(t *testing.T) {
	s := NewScheduler()
	_ = s.AddTask("test", "Test", Every(time.Second), func(ctx context.Context) error { return nil })

	err := s.DisableTask("test")
	if err != nil {
		t.Fatalf("DisableTask() error = %v", err)
	}

	task, _ := s.GetTask("test")
	if task.Enabled {
		t.Error("Task should be disabled")
	}

	err = s.EnableTask("test")
	if err != nil {
		t.Fatalf("EnableTask() error = %v", err)
	}

	task, _ = s.GetTask("test")
	if !task.Enabled {
		t.Error("Task should be enabled")
	}
}

func TestScheduler_GetTasks(t *testing.T) {
	s := NewScheduler()
	_ = s.AddTask("test1", "Test 1", Every(time.Second), func(ctx context.Context) error { return nil })
	_ = s.AddTask("test2", "Test 2", Every(time.Second), func(ctx context.Context) error { return nil })

	tasks := s.GetTasks()
	if len(tasks) != 2 {
		t.Errorf("GetTasks() returned %d tasks, want 2", len(tasks))
	}
}

func TestScheduler_RunTask(t *testing.T) {
	s := NewScheduler()
	var wg sync.WaitGroup
	called := false

	wg.Add(1)
	_ = s.AddTask("test", "Test",
		Every(100*time.Millisecond),
		func(ctx context.Context) error {
			called = true
			wg.Done()
			return nil
		},
	)

	_ = s.Start()
	defer s.Stop()

	wg.Wait()

	if !called {
		t.Error("Task was not executed")
	}
}

func TestScheduler_Stop(t *testing.T) {
	s := NewScheduler()
	_ = s.Start()
	s.Stop()
	// Should not panic
}

func TestEverySchedule(t *testing.T) {
	schedule := Every(time.Hour)
	now := time.Now()
	next := schedule.Next(now)

	if next.Before(now) || next.Equal(now) {
		t.Error("Next time should be in the future")
	}

	if next.Sub(now) != time.Hour {
		t.Errorf("Next time should be 1 hour from now, got %v", next.Sub(now))
	}
}

func TestDailySchedule(t *testing.T) {
	schedule := Daily(10, 30) // 10:30 AM
	now := time.Date(2024, 1, 1, 9, 0, 0, 0, time.Local)
	next := schedule.Next(now)

	if next.Hour() != 10 || next.Minute() != 30 {
		t.Errorf("Next time should be at 10:30, got %02d:%02d", next.Hour(), next.Minute())
	}

	// Test when time has passed today
	now = time.Date(2024, 1, 1, 11, 0, 0, 0, time.Local)
	next = schedule.Next(now)

	if next.Day() != 2 {
		t.Error("Next time should be tomorrow")
	}
}

func TestWeeklySchedule(t *testing.T) {
	schedule := Weekly(time.Monday, 9, 0) // Monday 9:00 AM
	now := time.Date(2024, 1, 1, 8, 0, 0, 0, time.Local) // Assuming this is not a Monday

	next := schedule.Next(now)

	if next.Weekday() != time.Monday {
		t.Errorf("Next time should be on Monday, got %v", next.Weekday())
	}

	if next.Hour() != 9 {
		t.Errorf("Next time should be at 9:00, got %02d:00", next.Hour())
	}
}

func TestCronSchedule(t *testing.T) {
	schedule := Cron(time.Minute * 5)
	now := time.Now()
	next := schedule.Next(now)

	diff := next.Sub(now)
	if diff != time.Minute*5 {
		t.Errorf("Next time should be 5 minutes from now, got %v", diff)
	}
}

func TestScheduler_TaskError(t *testing.T) {
	s := NewScheduler()
	var wg sync.WaitGroup

	wg.Add(1)
	_ = s.AddTask("test", "Test",
		Every(100*time.Millisecond),
		func(ctx context.Context) error {
			wg.Done()
			return context.DeadlineExceeded
		},
	)

	_ = s.Start()
	defer s.Stop()

	wg.Wait()

	task, _ := s.GetTask("test")
	if len(task.Errors) == 0 {
		t.Error("Task errors should be recorded")
	}
}

func TestDefaultScheduler(t *testing.T) {
	t.Run("AddTask", func(t *testing.T) {
		err := AddTask("global-test", "Global Test",
			Every(time.Hour),
			func(ctx context.Context) error { return nil },
		)
		if err != nil {
			t.Fatalf("AddTask() error = %v", err)
		}
		_ = DefaultScheduler.RemoveTask("global-test")
	})
}

func TestScheduler_SetLogger(t *testing.T) {
	s := NewScheduler()
	logger := &mockLogger{}
	s.SetLogger(logger)
	// Logger should be set without error
}

type mockLogger struct{}

func (m *mockLogger) Info(message string, fields map[string]interface{})  {}
func (m *mockLogger) Error(message string, fields map[string]interface{}) {}
func (m *mockLogger) Debug(message string, fields map[string]interface{}) {}

