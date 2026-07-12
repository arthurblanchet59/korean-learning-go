package main

import "testing"

func TestSplitFields(t *testing.T) {
	fields := splitFields("deck-id | 안녕하세요 | bonjour")
	if len(fields) != 3 || fields[1] != "안녕하세요" {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}

func TestVisibleBoundsKeepsCursorVisible(t *testing.T) {
	start, end := visibleBounds(100, 52, 10)
	if start > 52 || end <= 52 || end-start != 10 {
		t.Fatalf("cursor not visible in [%d:%d]", start, end)
	}
}
