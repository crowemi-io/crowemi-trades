package task

import "testing"

func TestPositionTask_Name(t *testing.T) {
	task := &PositionTask{}
	if got, want := task.Name(), "position_sync"; got != want {
		t.Fatalf("PositionTask.Name() = %q, want %q", got, want)
	}
}

func TestPositionTask_Schedule(t *testing.T) {
	t.Run("uses configured schedule", func(t *testing.T) {
		task := &PositionTask{CronSchedule: "30 * * * *"}
		if got, want := task.Schedule(), "30 * * * *"; got != want {
			t.Fatalf("PositionTask.Schedule() = %q, want %q", got, want)
		}
	})

	t.Run("uses default schedule when empty", func(t *testing.T) {
		task := &PositionTask{}
		if got, want := task.Schedule(), "0/30 * * * *"; got != want {
			t.Fatalf("PositionTask.Schedule() = %q, want %q", got, want)
		}
	})
}
