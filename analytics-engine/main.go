package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Event struct {
	ID         string  `json:"id"`
	EventType  string  `json:"event"`
	WorkflowID string  `json:"workflow_id,omitempty"`
	Timestamp  float64 `json:"timestamp"`
}

type EventStore struct {
	mu     sync.RWMutex
	events []Event
}

func NewEventStore() *EventStore {
	return &EventStore{events: make([]Event, 0)}
}

func (s *EventStore) Add(e Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, e)
}

func (s *EventStore) GetAll() []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Event, len(s.events))
	copy(result, s.events)
	return result
}

func (s *EventStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.events)
}

var store = NewEventStore()
var logger = log.New(os.Stdout, "[analytics-engine] ", log.LstdFlags)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"service":   "analytics-engine",
		"timestamp": time.Now().Unix(),
	})
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		EventType  string `json:"event"`
		WorkflowID string `json:"workflow_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Printf("Invalid request body: %v", err)
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.EventType == "" {
		logger.Println("Missing event type")
		http.Error(w, `{"error":"event type is required"}`, http.StatusBadRequest)
		return
	}

	event := Event{
		ID:         fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		EventType:  req.EventType,
		WorkflowID: req.WorkflowID,
		Timestamp:  float64(time.Now().Unix()),
	}
	store.Add(event)
	logger.Printf("Event recorded: %s (type=%s)", event.ID, event.EventType)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}

func eventsListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	events := store.GetAll()
	logger.Printf("Listing %d events", len(events))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"total":  len(events),
	})
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	events := store.GetAll()
	typeCounts := make(map[string]int)
	for _, e := range events {
		typeCounts[e.EventType]++
	}
	logger.Printf("Stats requested: %d total events", len(events))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_events":    len(events),
		"events_by_type":  typeCounts,
		"uptime_seconds":  time.Since(startTime).Seconds(),
	})
}

var startTime = time.Now()

func main() {
	port := os.Getenv("ANALYTICS_PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/analytics/event", eventHandler)
	mux.HandleFunc("/analytics/events", eventsListHandler)
	mux.HandleFunc("/analytics/stats", statsHandler)

	logger.Printf("Starting Analytics Engine on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Fatalf("Server failed: %v", err)
	}
}
