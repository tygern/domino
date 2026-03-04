package web

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
	"time"

	"github.com/tygern/domino/internal/coxeter"
	"github.com/tygern/domino/internal/svg"
	"github.com/tygern/domino/internal/tableau"
)

type templateData struct {
	Permutation  string
	Rank         int
	Length       int
	RightDescent string
	LeftDescent  string
	IsBad        bool
	Reduced      string
	RightTableau template.HTML
	LeftTableau  template.HTML
	HeapSVG      template.HTML
	Message      string
	Perm         string
	Display      string
	Count        int
	TimedOut     bool
	Rows         []templateData
}

func Handlers() *http.ServeMux {
	mux := http.NewServeMux()

	staticSub, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	mux.HandleFunc("GET /element", handleElement)
	mux.HandleFunc("GET /bad", handleBad)
	mux.HandleFunc("GET /bad/sample", handleBadSample)
	mux.HandleFunc("GET /bad/download", handleBadDownload)
	mux.HandleFunc("GET /", handleHome)

	return mux
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	render(w, homeTmpl, templateData{})
}

func handleElement(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	permStr := q.Get("perm")
	exprStr := q.Get("expr")
	rankStr := q.Get("rank")

	var elem coxeter.Element
	var err error

	switch {
	case permStr != "":
		elem, err = parsePermutation(permStr)
	case exprStr != "":
		rank, rankErr := strconv.Atoi(rankStr)
		if rankErr != nil || rank < 2 {
			render(w, errorTmpl, templateData{Message: "A valid rank is required with an expression."})
			return
		}
		elem, err = parseExpression(exprStr, rank)
	default:
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err != nil {
		render(w, errorTmpl, templateData{Message: err.Error()})
		return
	}

	rightTab := tableau.RightTableau(elem)
	leftTab := tableau.LeftTableau(elem)
	heap := tableau.NewHeap(elem)

	data := templateData{
		Permutation:  elem.String(),
		Rank:         elem.Rank(),
		Length:       elem.Length(),
		RightDescent: formatSet(elem.RightDescentSet()),
		LeftDescent:  formatSet(elem.LeftDescentSet()),
		IsBad:        elem.IsBad(),
		Reduced:      elem.ReducedExpression().String(),
		RightTableau: template.HTML(svg.RenderTableau(rightTab)),
		LeftTableau:  template.HTML(svg.RenderTableau(leftTab)),
		HeapSVG:      template.HTML(svg.RenderHeap(heap)),
	}

	render(w, elementTmpl, data)
}

const maxEnumerateRank = 14

func handleBad(w http.ResponseWriter, r *http.Request) {
	rankStr := r.URL.Query().Get("rank")
	if rankStr == "" {
		rows := make([]templateData, 0, 28)
		for n := 3; n <= 30; n++ {
			rows = append(rows, templateData{Rank: n, Count: badElementCount(n)})
		}
		render(w, badFormTmpl, templateData{Rows: rows})
		return
	}

	rank, err := strconv.Atoi(rankStr)
	if err != nil || rank < 2 {
		render(w, errorTmpl, templateData{Message: "Invalid rank. Must be an integer >= 2."})
		return
	}

	if rank > maxEnumerateRank {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
		defer cancel()
		samples := coxeter.BadElementsSample(ctx, rank, 10)
		rows := make([]templateData, len(samples))
		for i, elem := range samples {
			rows[i] = templateData{
				Perm:    formatPermForURL(elem),
				Display: elem.String(),
			}
		}
		render(w, badLargeTmpl, templateData{Rank: rank, Count: badElementCount(rank), Rows: rows})
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	renderStream(w, "bad_head", templateData{Rank: rank})
	flusher.Flush()

	count := 0
	coxeter.BadElementsStream(ctx, rank, func(elem coxeter.Element) {
		count++
		renderStream(w, "bad_item", templateData{
			Perm:    formatPermForURL(elem),
			Display: elem.String(),
		})
		flusher.Flush()
	})

	renderStream(w, "bad_done", templateData{
		Count:    count,
		Rank:     rank,
		TimedOut: ctx.Err() == context.DeadlineExceeded,
	})
	flusher.Flush()
}

func handleBadSample(w http.ResponseWriter, r *http.Request) {
	rankStr := r.URL.Query().Get("rank")
	rank, err := strconv.Atoi(rankStr)
	if err != nil || rank < 2 {
		http.Error(w, "invalid rank", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	samples := coxeter.BadElementsSample(ctx, rank, 10)

	type sampleItem struct {
		Perm    string `json:"perm"`
		Display string `json:"display"`
	}
	items := make([]sampleItem, len(samples))
	for i, elem := range samples {
		items[i] = sampleItem{
			Perm:    formatPermForURL(elem),
			Display: elem.String(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func handleBadDownload(w http.ResponseWriter, r *http.Request) {
	rankStr := r.URL.Query().Get("rank")
	rank, err := strconv.Atoi(rankStr)
	if err != nil || rank < 2 {
		render(w, errorTmpl, templateData{Message: "Invalid rank. Must be an integer >= 2."})
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="bad_elements_d%d.csv"`, rank))

	coxeter.BadElementsStream(ctx, rank, func(elem coxeter.Element) {
		fmt.Fprintln(w, formatPermForURL(elem))
		flusher.Flush()
	})
}

func badElementCount(rank int) int {
	if rank < 3 {
		return 0
	}

	prev2, prev1 := 0, 0
	for n := 3; n <= rank; n++ {
		next := prev1 + prev2
		if n%2 == 0 {
			next++
		}
		prev2 = prev1
		prev1 = next
	}
	return prev1
}
