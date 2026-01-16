package sources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestDevTo_Search_Robustness(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		w.Write([]byte(`[{"title": "Broken", "url": ]`)) 
	}))
	defer ts.Close()

	d := &DevTo{
		Client:  ts.Client(),
		BaseURL: ts.URL,
	}

	posts, err := d.Search(context.Background(), "golang")

	assert.Error(t, err, "Should return an error for malformed JSON")
	assert.Nil(t, posts)
}