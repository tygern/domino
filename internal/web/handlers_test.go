package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleHome(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handleHome(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Element Explorer")
	assert.Contains(t, w.Body.String(), `action="/element"`)
}

func TestHandleElement_Perm(t *testing.T) {
	req := httptest.NewRequest("GET", "/element?perm=1,-4,3,-2", nil)
	w := httptest.NewRecorder()
	handleElement(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "[1, -4, 3, -2]")
	assert.Contains(t, body, "<svg")
	assert.Contains(t, body, "Length")
}

func TestHandleElement_Expr(t *testing.T) {
	req := httptest.NewRequest("GET", "/element?expr=3&rank=4", nil)
	w := httptest.NewRecorder()
	handleElement(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "[1, 3, 2, 4]")
}

func TestHandleElement_InvalidPerm(t *testing.T) {
	req := httptest.NewRequest("GET", "/element?perm=1,-2,3", nil)
	w := httptest.NewRecorder()
	handleElement(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Error")
}

func TestHandleElement_NoParams(t *testing.T) {
	req := httptest.NewRequest("GET", "/element", nil)
	w := httptest.NewRecorder()
	handleElement(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
}

func TestHandleBad_NoRank(t *testing.T) {
	req := httptest.NewRequest("GET", "/bad", nil)
	w := httptest.NewRecorder()
	handleBad(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "<table>")
	assert.Contains(t, body, "View")
	assert.Contains(t, body, "Download CSV")
	assert.Contains(t, body, "Enumerate")
	assert.Contains(t, body, `/bad?rank=22">View`)
}

func TestHandleBad_Rank4(t *testing.T) {
	req := httptest.NewRequest("GET", "/bad?rank=4", nil)
	w := httptest.NewRecorder()
	handleBad(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "[1, -4, 3, -2]")
	assert.Contains(t, body, "1 bad elements")
}

func TestHandleBad_Rank3(t *testing.T) {
	req := httptest.NewRequest("GET", "/bad?rank=3", nil)
	w := httptest.NewRecorder()
	handleBad(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "0 bad elements")
}

func TestHandleBad_LargeRank(t *testing.T) {
	req := httptest.NewRequest("GET", "/bad?rank=22", nil)
	w := httptest.NewRecorder()
	handleBad(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "6765")
	assert.Contains(t, body, "/bad/download?rank=22")
	assert.True(t, strings.Contains(body, "<li>"), "should contain sample elements")
	assert.Contains(t, body, "Show 10 more")
}

func TestHandleBadSample(t *testing.T) {
	req := httptest.NewRequest("GET", "/bad/sample?rank=8", nil)
	w := httptest.NewRecorder()
	handleBadSample(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var items []struct {
		Perm    string `json:"perm"`
		Display string `json:"display"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &items)
	assert.NoError(t, err)
	assert.Len(t, items, 8, "D_8 has 8 bad elements, all should be returned")
	for _, item := range items {
		assert.NotEmpty(t, item.Perm)
		assert.NotEmpty(t, item.Display)
	}
}

func TestHandleBad_InvalidRank(t *testing.T) {
	req := httptest.NewRequest("GET", "/bad?rank=abc", nil)
	w := httptest.NewRecorder()
	handleBad(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Error")
}

func TestHandleBadDownload(t *testing.T) {
	req := httptest.NewRequest("GET", "/bad/download?rank=4", nil)
	w := httptest.NewRecorder()
	handleBadDownload(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "bad_elements_d4.csv")
	assert.Contains(t, w.Body.String(), "1,-4,3,-2")
}

func TestHandlers_StaticRoute(t *testing.T) {
	mux := Handlers()
	req := httptest.NewRequest("GET", "/static/style.css", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "font-family")
}

func TestHandleElement_BadElement(t *testing.T) {
	req := httptest.NewRequest("GET", "/element?perm=1,-4,3,-2", nil)
	w := httptest.NewRecorder()
	handleElement(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "bad")
	assert.True(t, strings.Contains(body, "Right tableau") || strings.Contains(body, "Right tableau"))
	assert.Contains(t, body, "Heap")
}

func TestBadElementCount(t *testing.T) {
	cases := []struct {
		rank int
		want int
	}{
		{2, 0},
		{3, 0},
		{4, 1},
		{5, 1},
		{6, 3},
		{7, 4},
		{8, 8},
		{10, 21},
		{14, 144},
		{22, 6765},
		{26, 46368},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, badElementCount(tc.rank), "rank %d", tc.rank)
	}
}
