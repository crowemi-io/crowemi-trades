package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type CronExpr struct {
	minute []bool // 0-59
	hour   []bool // 0-23
	dom    []bool // 1-31
	month  []bool // 1-12
	dow    []bool // 0-6 (Sunday=0)
}

func ParseCron(expr string) (CronExpr, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return CronExpr{}, fmt.Errorf("cron: expected 5 fields, got %d", len(fields))
	}

	minute, err := parseField(fields[0], 0, 59)
	if err != nil {
		return CronExpr{}, fmt.Errorf("cron: minute: %w", err)
	}
	hour, err := parseField(fields[1], 0, 23)
	if err != nil {
		return CronExpr{}, fmt.Errorf("cron: hour: %w", err)
	}
	dom, err := parseField(fields[2], 1, 31)
	if err != nil {
		return CronExpr{}, fmt.Errorf("cron: day-of-month: %w", err)
	}
	month, err := parseField(fields[3], 1, 12)
	if err != nil {
		return CronExpr{}, fmt.Errorf("cron: month: %w", err)
	}
	dow, err := parseField(fields[4], 0, 6)
	if err != nil {
		return CronExpr{}, fmt.Errorf("cron: day-of-week: %w", err)
	}

	return CronExpr{
		minute: minute,
		hour:   hour,
		dom:    dom,
		month:  month,
		dow:    dow,
	}, nil
}

// Next returns the next time after `from` that matches the cron expression.
// It searches up to 4 years ahead to handle leap-year edge cases.
func (c CronExpr) Next(from time.Time) time.Time {
	t := from.Truncate(time.Minute).Add(time.Minute)

	limit := t.AddDate(4, 0, 0)
	for t.Before(limit) {
		if !c.month[t.Month()] {
			t = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
			continue
		}
		if !c.dom[t.Day()] || !c.dow[t.Weekday()] {
			t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())
			continue
		}
		if !c.hour[t.Hour()] {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, t.Location())
			continue
		}
		if !c.minute[t.Minute()] {
			t = t.Add(time.Minute)
			continue
		}
		return t
	}
	return time.Time{}
}

// parseField parses a single cron field into a boolean slice where index i is
// true if value i is included. The slice length is max+1.
func parseField(field string, min, max int) ([]bool, error) {
	set := make([]bool, max+1)
	for _, part := range strings.Split(field, ",") {
		if err := parsePart(part, min, max, set); err != nil {
			return nil, err
		}
	}
	return set, nil
}

func parsePart(part string, min, max int, set []bool) error {
	var rangeStr, stepStr string
	if idx := strings.Index(part, "/"); idx != -1 {
		rangeStr = part[:idx]
		stepStr = part[idx+1:]
	} else {
		rangeStr = part
	}

	var lo, hi int
	switch {
	case rangeStr == "*":
		lo, hi = min, max
	case strings.Contains(rangeStr, "-"):
		parts := strings.SplitN(rangeStr, "-", 2)
		var err error
		lo, err = strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("invalid range start %q", parts[0])
		}
		hi, err = strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("invalid range end %q", parts[1])
		}
	default:
		v, err := strconv.Atoi(rangeStr)
		if err != nil {
			return fmt.Errorf("invalid value %q", rangeStr)
		}
		lo, hi = v, v
	}

	if lo < min || hi > max || lo > hi {
		return fmt.Errorf("value %d-%d out of range %d-%d", lo, hi, min, max)
	}

	step := 1
	if stepStr != "" {
		var err error
		step, err = strconv.Atoi(stepStr)
		if err != nil || step <= 0 {
			return fmt.Errorf("invalid step %q", stepStr)
		}
	}

	for i := lo; i <= hi; i += step {
		set[i] = true
	}
	return nil
}
