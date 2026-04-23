package task

import "testing"

func TestOrderTask_Name(t *testing.T) {
	task := &OrderTask{}
	if got, want := task.Name(), "order_sync"; got != want {
		t.Fatalf("OrderTask.Name() = %q, want %q", got, want)
	}
}

func TestOrderTask_DefaultSchedule(t *testing.T) {
	task := &OrderTask{}
	if got, want := task.DefaultSchedule(), "0/30 * * * *"; got != want {
		t.Fatalf("OrderTask.DefaultSchedule() = %q, want %q", got, want)
	}
}

func TestOrderTask_Schedule(t *testing.T) {
	t.Run("uses configured schedule", func(t *testing.T) {
		task := &OrderTask{CronSchedule: "15 * * * *"}
		if got, want := task.Schedule(), "15 * * * *"; got != want {
			t.Fatalf("OrderTask.Schedule() = %q, want %q", got, want)
		}
	})

	t.Run("uses default schedule when empty", func(t *testing.T) {
		task := &OrderTask{}
		if got, want := task.Schedule(), "0/30 * * * *"; got != want {
			t.Fatalf("OrderTask.Schedule() = %q, want %q", got, want)
		}
	})
}
