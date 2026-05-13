package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "healthy" {
		t.Errorf("expected healthy status")
	}
	if resp["service"] != "analytics-engine" {
		t.Errorf("expected analytics-engine service")
	}
}

func TestEventHandler_Success(t *testing.T) {
	body, _ := json.Marshal(map[string]string{
		"event":       "workflow_created",
		"workflow_id": "wf-123",
	})
	req := httptest.NewRequest(http.MethodPost, "/analytics/event", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	eventHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	var resp Event
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.EventType != "workflow_created" {
		t.Errorf("expected workflow_created event type, got %s", resp.EventType)
	}
}

func TestEventHandler_MissingEventType(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"workflow_id": "wf-123"})
	req := httptest.NewRequest(http.MethodPost, "/analytics/event", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	eventHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEventHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/analytics/event", nil)
	w := httptest.NewRecorder()
	eventHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestEventsListHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/analytics/events", nil)
	w := httptest.NewRecorder()
	eventsListHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if _, ok := resp["events"]; !ok {
		t.Errorf("expected events field in response")
	}
}

func TestStatsHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/analytics/stats", nil)
	w := httptest.NewRecorder()
	statsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if _, ok := resp["total_events"]; !ok {
		t.Errorf("expected total_events field")
	}
	if _, ok := resp["uptime_seconds"]; !ok {
		t.Errorf("expected uptime_seconds field")
	}
}

func TestStatsHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/analytics/stats", nil)
	w := httptest.NewRecorder()
	statsHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
