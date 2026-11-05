package analysis

import (
	"contractcheck/back/internal/domain"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Document = domain.Document
type Paragraph = domain.Paragraph
type Evidence = domain.Evidence
type Field = domain.Field

var (
	contractIDRE = regexp.MustCompile(`(?i)\bSTN-\d+\b`)
	amountRE     = regexp.MustCompile(`(?i)(?:составляет|стоимость поставки на)\s+([0-9][0-9\s\x{00a0}]*)`)
	dateRE       = regexp.MustCompile(`\b([0-3]?\d)\.([01]?\d)\.(\d{4})\b`)
)

func ParseDocument(id, name, body string) Document {
	d := Document{ID: id, Name: name, Kind: "unknown", Status: "unknown", Paragraphs: []Paragraph{}}
	body = strings.ReplaceAll(strings.TrimPrefix(body, "\ufeff"), "\r\n", "\n")
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			d.Paragraphs = append(d.Paragraphs, Paragraph{ID: "p" + strconv.Itoa(len(d.Paragraphs)+1), Text: line})
		}
	}
	if len(d.Paragraphs) == 0 {
		return d
	}
	header := strings.ToLower(d.Paragraphs[0].Text)
	switch {
	case strings.Contains(header, "проект") && strings.Contains(header, "соглашения"):
		d.Kind = "amendment"
		d.Status = "draft"
	case strings.Contains(header, "дополнительное соглашение"):
		d.Kind = "amendment"
	case strings.Contains(header, "договор поставки"):
		d.Kind = "contract"
	case strings.Contains(header, "письмо"):
		d.Kind = "email"
		d.Status = "proposal"
	}
	if m := contractIDRE.FindString(body); m != "" {
		d.ContractID = strings.ToUpper(m)
	}
	for _, p := range d.Paragraphs {
		low := strings.ToLower(p.Text)
		if strings.HasPrefix(low, "статус:") {
			if strings.Contains(low, "не подписан") || strings.Contains(low, "проект") {
				d.Status = "draft"
			} else if strings.Contains(low, "подписан") && strings.Contains(low, "обеими сторонами") {
				d.Status = "signed"
			}
		}
	}
	return d
}

func evidence(d Document, p Paragraph) *Evidence {
	return &Evidence{DocumentID: d.ID, FileName: d.Name, ParagraphID: p.ID, Quote: p.Text}
}
func firstEvidence(d Document) Evidence {
	if len(d.Paragraphs) == 0 {
		return Evidence{DocumentID: d.ID, FileName: d.Name}
	}
	return *evidence(d, d.Paragraphs[0])
}

func extractSupplier(d Document) Field {
	for _, p := range d.Paragraphs {
		idx := strings.Index(strings.ToLower(p.Text), "поставщик:")
		if idx < 0 {
			continue
		}
		value := strings.TrimSpace(p.Text[idx+len("поставщик:"):])
		if end := strings.Index(value, "."); end >= 0 {
			value = value[:end]
		}
		if value != "" {
			return Field{Value: value, Evidence: evidence(d, p)}
		}
	}
	return Field{}
}

func extractAmount(d Document) Field {
	for _, p := range d.Paragraphs {
		low := strings.ToLower(p.Text)
		eligible := false
		switch d.Kind {
		case "contract":
			eligible = strings.HasPrefix(low, "2.1.") && strings.Contains(low, "общая стоимость")
		case "amendment":
			eligible = (strings.Contains(low, "пункт 2.1") || strings.Contains(low, "общую стоимость")) && strings.Contains(low, "стоимост")
		}
		if !eligible {
			continue
		}
		if !strings.Contains(low, "руб") {
			continue
		}
		m := amountRE.FindStringSubmatch(p.Text)
		if len(m) < 2 {
			continue
		}
		value := strings.ReplaceAll(strings.ReplaceAll(m[1], " ", ""), "\u00a0", "")
		if _, err := strconv.ParseInt(value, 10, 64); err == nil {
			return Field{Value: value, Evidence: evidence(d, p)}
		}
	}
	return Field{}
}

func extractDeadline(d Document) Field {
	for _, p := range d.Paragraphs {
		low := strings.ToLower(p.Text)
		eligible := false
		switch d.Kind {
		case "contract":
			eligible = strings.HasPrefix(low, "3.1.") && strings.Contains(low, "постав")
		case "amendment":
			eligible = (strings.Contains(low, "пункт 3.1") || strings.Contains(low, "срок поставки")) && strings.Contains(low, "постав")
		}
		if !eligible {
			continue
		}
		if date := lastDate(p.Text); date != "" {
			return Field{Value: date, Evidence: evidence(d, p)}
		}
	}
	return Field{}
}

func extractProposedDeadline(d Document) Field {
	for _, p := range d.Paragraphs {
		low := strings.ToLower(p.Text)
		if !strings.Contains(low, "перенести срок поставки") {
			continue
		}
		if date := lastDate(p.Text); date != "" {
			return Field{Value: date, Evidence: evidence(d, p)}
		}
	}
	return Field{}
}

func changeClause(d Document, kind string) *Evidence {
	for _, p := range d.Paragraphs {
		low := strings.ToLower(p.Text)
		if kind == "amount" && (strings.Contains(low, "пункт 2.1") || strings.Contains(low, "изменить общую стоимость")) {
			return evidence(d, p)
		}
		if kind == "deadline" && (strings.Contains(low, "пункт 3.1") || strings.Contains(low, "изменить срок поставки")) {
			return evidence(d, p)
		}
	}
	return nil
}

func effectiveDate(d Document) time.Time {
	for _, p := range d.Paragraphs {
		low := strings.ToLower(p.Text)
		if !strings.Contains(low, "вступает в силу") {
			continue
		}
		if date := lastDate(p.Text); date != "" {
			t, _ := time.Parse("2006-01-02", date)
			return t
		}
	}
	return time.Time{}
}

func signedAt(d Document) time.Time {
	for _, p := range d.Paragraphs {
		low := strings.ToLower(p.Text)
		if !strings.HasPrefix(low, "статус:") || !strings.Contains(low, "подписан") {
			continue
		}
		if date := lastDate(p.Text); date != "" {
			t, _ := time.Parse("2006-01-02", date)
			return t
		}
	}
	return time.Time{}
}

func lastDate(s string) string {
	matches := dateRE.FindAllString(s, -1)
	for i := len(matches) - 1; i >= 0; i-- {
		if d, err := parseRuDate(matches[i]); err == nil {
			return d.Format("2006-01-02")
		}
	}
	return ""
}

var errInvalidDate = errors.New("invalid date")

func parseRuDate(s string) (time.Time, error) {
	p := strings.Split(s, ".")
	if len(p) != 3 {
		return time.Time{}, errInvalidDate
	}
	d, err := strconv.Atoi(p[0])
	if err != nil {
		return time.Time{}, err
	}
	m, err := strconv.Atoi(p[1])
	if err != nil {
		return time.Time{}, err
	}
	y, err := strconv.Atoi(p[2])
	if err != nil {
		return time.Time{}, err
	}
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	if t.Day() != d || int(t.Month()) != m || t.Year() != y {
		return time.Time{}, errInvalidDate
	}
	return t, nil
}
