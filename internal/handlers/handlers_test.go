package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/Numpkens/grip/internal/logic"
)

func TestHandleHome_JSON(t *testing.T) {
	h := &Handler{
		Engine: &logic.Engine{Sources: []logic.Source{}},
	}

	req, _ := http.NewRequest("GET", "/api/search?q=golang", nil)
	req.Header.Set("Accept", "application/json")

	rr := httptest.NewRecorder()
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts := h.Engine.Collect(r.Context(), "golang")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(posts)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var posts []logic.Post
	if err := json.NewDecoder(rr.Body).Decode(&posts); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if len(posts) != 0 {
		t.Errorf("Expected 0 posts for empty sources, got %d", len(posts))
	}
}