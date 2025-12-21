package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/prashantsinghb/workflow-engine/pkg/execution"
	"github.com/prashantsinghb/workflow-engine/pkg/execution/timeline"
)

type Server struct {
	timelineBuilder *timeline.TimelineBuilder
}

func NewExecutionTimelineServer(store execution.Store) *Server {
	return &Server{
		timelineBuilder: timeline.NewTimelineBuilder(store),
	}
}

func (s *Server) GetExecutionTimeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectID := chi.URLParam(r, "projectId")
	execIDStr := chi.URLParam(r, "executionId")

	execID, err := uuid.Parse(execIDStr)
	if err != nil {
		http.Error(w, "invalid execution id", http.StatusBadRequest)
		return
	}

	timeline, err := s.timelineBuilder.Build(ctx, projectID, execID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timeline)
}
