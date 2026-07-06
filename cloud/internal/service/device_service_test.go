package service

import (
	"testing"
	"time"
)

func TestParseBindCodeExpirationTreatsNaiveTimestampAsUTC(t *testing.T) {
	got, err := parseBindCodeExpiration("2026-07-06T10:42:01")
	if err != nil {
		t.Fatalf("parseBindCodeExpiration returned error: %v", err)
	}

	want := time.Date(2026, 7, 6, 10, 42, 1, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("parsed time = %s, want %s", got.Format(time.RFC3339), want.Format(time.RFC3339))
	}
}

func TestParseBindCodeExpirationAcceptsRFC3339Timestamp(t *testing.T) {
	got, err := parseBindCodeExpiration("2026-07-06T10:42:01Z")
	if err != nil {
		t.Fatalf("parseBindCodeExpiration returned error: %v", err)
	}

	want := time.Date(2026, 7, 6, 10, 42, 1, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("parsed time = %s, want %s", got.Format(time.RFC3339), want.Format(time.RFC3339))
	}
}
