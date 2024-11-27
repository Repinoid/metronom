package main

import (
	"net/http"
	"testing"
)


func TestgetMetrix(t *testing.T) {
	result1 := getMetrix(memStor)
	if result1 != nil {
		t.Errorf("Result was incorrect, got: %d, want: %s.", result1, "nil")
	}
	result2 := postMetric("gaug", "Alloc", "55.55")
	if result2 != http.StatusBadRequest {
		t.Errorf("Result was incorrect, got: %d, want: %s.", result2, "http.StatusBadRequest")
	}
	result3 := postMetric("gauge", "Alloc", "a55.55")
	if result3 != http.StatusBadRequest {
		t.Errorf("Result was incorrect, got: %d, want: %s.", result3, "http.StatusBadRequest")
	}
}
