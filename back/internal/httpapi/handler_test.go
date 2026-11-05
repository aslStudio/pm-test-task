package httpapi

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"contractcheck/back/internal/analysis"
)

func uploadRequest(t *testing.T, date string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("as_of", date)
	file, err := writer.CreateFormFile("files", "contract.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = file.Write([]byte("ДОГОВОР ПОСТАВКИ № STN-105\nСтатус: подписан обеими сторонами 01.03.2026.\n2.1. Общая стоимость поставки составляет 25 000 рублей.\n3.1. Поставщик обязуется поставить оборудование не позднее 15.05.2026."))
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, "/api/analyze", &body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	return r
}

func TestAnalyzeRoute(t *testing.T) {
	w := httptest.NewRecorder()
	NewHandler(analysis.NewService()).Routes().ServeHTTP(w, uploadRequest(t, "2026-03-10"))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	var body analyzeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Contracts) != 1 || body.Contracts[0].Amount.Value != "25000" || body.Contracts[0].Amount.Evidence.Quote == "" {
		t.Fatalf("response: %+v", body)
	}
}

func TestInvalidDate(t *testing.T) {
	w := httptest.NewRecorder()
	NewHandler(analysis.NewService()).Routes().ServeHTTP(w, uploadRequest(t, "invalid"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestMissingDate(t *testing.T) {
	w := httptest.NewRecorder()
	NewHandler(analysis.NewService()).Routes().ServeHTTP(w, uploadRequest(t, ""))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", w.Code)
	}
}
