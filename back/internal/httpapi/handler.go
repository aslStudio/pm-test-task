package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"contractcheck/back/internal/analysis"
	"contractcheck/back/internal/domain"
)

type Analyzer interface {
	Analyze(files []analysis.InputFile, asOf time.Time) domain.Result
}

type Handler struct{ analyzer Analyzer }

func NewHandler(analyzer Analyzer) *Handler { return &Handler{analyzer: analyzer} }

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/analyze", h.analyze)
	return mux
}

func (h *Handler) analyze(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, "Не удалось прочитать файлы (общий лимит 10 МБ)")
		return
	}
	asOf := r.FormValue("as_of")
	if asOf == "" {
		writeError(w, "Укажите дату проверки в формате YYYY-MM-DD")
		return
	}
	date, err := time.Parse("2006-01-02", asOf)
	if err != nil {
		writeError(w, "Дата проверки должна иметь формат YYYY-MM-DD")
		return
	}
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		writeError(w, "Загрузите хотя бы один TXT-файл")
		return
	}
	if len(files) > 100 {
		writeError(w, "Допускается не более 100 файлов")
		return
	}
	input := make([]analysis.InputFile, 0, len(files))
	for _, fh := range files {
		if !strings.HasSuffix(strings.ToLower(fh.Filename), ".txt") {
			writeError(w, "Поддерживаются только TXT-файлы")
			return
		}
		f, err := fh.Open()
		if err != nil {
			writeError(w, "Не удалось открыть файл")
			return
		}
		body, readErr := io.ReadAll(io.LimitReader(f, (1<<20)+1))
		_ = f.Close()
		if readErr != nil || len(body) > 1<<20 {
			writeError(w, "Файл превышает 1 МБ")
			return
		}
		if !utf8.Valid(body) {
			writeError(w, "TXT должен быть в UTF-8")
			return
		}
		input = append(input, analysis.InputFile{Name: fh.Filename, Body: string(body)})
	}
	writeJSON(w, http.StatusOK, toResponse(h.analyzer.Analyze(input, date)))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": message})
}
