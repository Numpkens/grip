package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/Numpkens/grip/internal/logic"
)

func TestHandleHome_JSON(t *testing.T) {
	h := &Handler{
		Engine: &logic.Engine{Sources: []logic.Source{}},
	}

	req, _ := http.NewRequest("GET", "/?q=golang", nil)
	req.Header.Set("Accept", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.HandleHome)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedContentType := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, expectedContentType)
	}
}
