package scheduler

import (
	"testing"
	"time"
)

func TestParseCronInvalidFieldCount(t *testing.T) {
	_, err := ParseCron("* * *")
	if err == nil {
		t.Fatal("expected error for 3 fields")
	}
}

func TestParseCronEveryMinute(t *testing.T) {
	expr, err := ParseCron("* * * * *")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	from := time.Date(2026, 1, 1, 12, 30, 0, 0, time.UTC)
	next := expr.Next(from)
	want := time.Date(2026, 1, 1, 12, 31, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("got %v, want %v", next, want)
	}
}

func TestParseCronEvery15Minutes(t *testing.T) {
	expr, err := ParseCron("*/15 * * * *")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	from := time.Date(2026, 1, 1, 12, 2, 0, 0, time.UTC)
	next := expr.Next(from)
	want := time.Date(2026, 1, 1, 12, 15, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("got %v, want %v", next, want)
	}
}

func TestParseCronSpecificTime(t *testing.T) {
	expr, err := ParseCron("30 9 * * 1")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// 2026-02-22 is a Sunday
	from := time.Date(2026, 2, 22, 10, 0, 0, 0, time.UTC)
	next := expr.Next(from)
	// Next Monday is 2026-02-23
	want := time.Date(2026, 2, 23, 9, 30, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("got %v, want %v", next, want)
	}
}

func TestParseCronRange(t *testing.T) {
	expr, err := ParseCron("0 9-17 * * *")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	from := time.Date(2026, 1, 1, 18, 0, 0, 0, time.UTC)
	next := expr.Next(from)
	want := time.Date(2026, 1, 2, 9, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("got %v, want %v", next, want)
	}
}

func TestParseCronCommaList(t *testing.T) {
	expr, err := ParseCron("0,30 * * * *")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	from := time.Date(2026, 1, 1, 12, 5, 0, 0, time.UTC)
	next := expr.Next(from)
	want := time.Date(2026, 1, 1, 12, 30, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("got %v, want %v", next, want)
	}
}

func TestParseCronRangeWithStep(t *testing.T) {
	expr, err := ParseCron("0-30/10 * * * *")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	from := time.Date(2026, 1, 1, 12, 5, 0, 0, time.UTC)
	next := expr.Next(from)
	want := time.Date(2026, 1, 1, 12, 10, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("got %v, want %v", next, want)
	}
}

func TestParseCronMonthBoundary(t *testing.T) {
	expr, err := ParseCron("0 0 1 * *")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	from := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	next := expr.Next(from)
	want := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("got %v, want %v", next, want)
	}
}

func TestParseCronInvalidExpression(t *testing.T) {
	cases := []string{
		"60 * * * *",
		"* 25 * * *",
		"* * 0 * *",
		"* * * 13 *",
		"* * * * 7",
		"abc * * * *",
	}
	for _, expr := range cases {
		if _, err := ParseCron(expr); err == nil {
			t.Errorf("expected error for %q", expr)
		}
	}
}
