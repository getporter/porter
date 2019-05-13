package printer

import (
	"math"
	"time"

	humanize "github.com/dustin/go-humanize"
)

const defaultDateFormat = "2006-01-02"

// DateTimePrinter uses go-humanize to print just the time portion of recent
// timestamps (within the last day) and for anything past a day, only prints the
// date portion of the timestamp.
type DateTimePrinter struct {
	// DateFormat to use for printing timestamps beyond a day old.
	DateFormat string

	// Now allows you use a fixed point in time when printing.
	// This is useful when printing tables of data, so all the relative times
	// are in relation to the same "now", and when unit testing.
	Now func() time.Time
}

var prettyFormats = []humanize.RelTimeMagnitude{
	{time.Second, "now", time.Second},
	{2 * time.Second, "1 second %s", 1},
	{time.Minute, "%d seconds %s", time.Second},
	{2 * time.Minute, "1 minute %s", 1},
	{time.Hour, "%d minutes %s", time.Minute},
	{2 * time.Hour, "1 hour %s", 1},
	{humanize.Day, "%d hours %s", time.Hour},
	{math.MaxInt64, "", 1},
}

// DateFormatOrDefault gets the format to apply to dates.
func (t DateTimePrinter) DateFormatOrDefault() string {
	if t.DateFormat != "" {
		return t.DateFormat
	}

	return defaultDateFormat
}

// NowOrDefault gets the current time, using the overridden Now(), or time.Now().
func (t DateTimePrinter) NowOrDefault() time.Time {
	if t.Now != nil {
		return t.Now()
	}

	return time.Now()
}

// Format the specified timestamp relative to now.
func (t DateTimePrinter) Format(value time.Time) string {
	relativeResult := humanize.CustomRelTime(value, t.NowOrDefault(), "ago", "from now", prettyFormats)
	if relativeResult != "" {
		return relativeResult
	}

	return value.Format(t.DateFormatOrDefault())
}
