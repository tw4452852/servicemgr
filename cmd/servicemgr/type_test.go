package main

import (
	"testing"
)

func TestType(t *testing.T) {
	for i := TypeBegin; i < TypeEnd+2; i++ {
		s := i.String()
		if TypeBegin < i && i < TypeEnd && (s == "unknown" || !i.IsValid()) {
			t.Errorf("type %d should be valid, but not", int(i))
		}
		if !(TypeBegin < i && i < TypeEnd) && (s != "unknown" || i.IsValid()) {
			t.Errorf("type %d should be invalid, but not", int(i))
		}
	}

	for i := ErrorBegin; i < ErrorEnd+2; i++ {
		s := i.String()
		if ErrorBegin < i && i < ErrorEnd && (s == "unknown" || !i.IsValid()) {
			t.Errorf("type %#x should be valid, but not", int(i))
		}
		if !(ErrorBegin < i && i < ErrorEnd) && (s != "unknown" || i.IsValid()) {
			t.Errorf("type %#x should be invalid, but not", int(i))
		}
	}

	if TypeEnd >= ErrorBegin {
		t.Errorf("types have overlay with errors: types[%#x-%#x], errors[%#x-%#x]",
			TypeBegin, TypeEnd, ErrorBegin, ErrorEnd)
	}
}
