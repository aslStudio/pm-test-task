package analysis

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fixtureFiles(t *testing.T) []InputFile {
	t.Helper()
	names := []string{"01_contract.txt", "02_amendment_signed.txt", "03_contract.txt", "04_contract.txt", "05_supplier_email.txt", "06_amendment_draft.txt"}
	files := make([]InputFile, 0, len(names))
	for _, name := range names {
		body, err := os.ReadFile(filepath.Join("..", "..", "..", "documents", name))
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, InputFile{Name: name, Body: string(body)})
	}
	return files
}

func asOf() time.Time { return time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC) }

func TestFixtureResultsAndEvidence(t *testing.T) {
	result := NewService().Analyze(fixtureFiles(t), asOf())
	if len(result.Contracts) != 3 {
		t.Fatalf("contracts = %d", len(result.Contracts))
	}
	byID := map[string]int{}
	for i, c := range result.Contracts {
		byID[c.ID] = i
	}
	c101 := result.Contracts[byID["STN-101"]]
	if c101.Amount.Value != "1380000" || c101.Deadline.Value != "2026-04-30" || !c101.Amount.Changed || len(c101.Notices) != 2 {
		t.Fatalf("STN-101: %+v", c101)
	}
	if c101.Amount.Evidence == nil || c101.Amount.Evidence.FileName != "02_amendment_signed.txt" || c101.Deadline.Evidence == nil || c101.Deadline.Evidence.FileName != "02_amendment_signed.txt" {
		t.Fatalf("signed amendment is not cited: %+v", c101)
	}
	c102 := result.Contracts[byID["STN-102"]]
	if c102.Amount.Value != "840000" || c102.Deadline.Value != "2026-03-20" || len(c102.Notices) != 1 {
		t.Fatalf("STN-102: %+v", c102)
	}
	if c102.Notices[0].Kind != "proposal" || c102.Notices[0].Evidence.FileName != "05_supplier_email.txt" {
		t.Fatalf("supplier proposal is not cited: %+v", c102.Notices)
	}
	c103 := result.Contracts[byID["STN-103"]]
	if c103.Amount.Value != "" || c103.Deadline.Value != "2026-04-30" {
		t.Fatalf("STN-103: %+v", c103)
	}
	for _, c := range result.Contracts {
		for _, field := range []Field{c.Amount, c.Deadline} {
			if field.Value == "" {
				continue
			}
			if field.Evidence == nil {
				t.Fatalf("missing evidence: %s", c.ID)
			}
			found := false
			for _, d := range result.Documents {
				if d.ID == field.Evidence.DocumentID {
					for _, p := range d.Paragraphs {
						if p.ID == field.Evidence.ParagraphID && p.Text == field.Evidence.Quote {
							found = true
						}
					}
				}
			}
			if !found {
				t.Fatalf("invalid evidence: %s", c.ID)
			}
		}
	}
}

func TestChangedValueAndNewContract(t *testing.T) {
	files := fixtureFiles(t)
	for i := range files {
		if files[i].Name == "03_contract.txt" {
			files[i].Body = strings.ReplaceAll(files[i].Body, "840 000", "910 000")
		}
	}
	files = append(files, InputFile{Name: "new.txt", Body: "ДОГОВОР ПОСТАВКИ № STN-104\nСтатус: подписан обеими сторонами 01.03.2026.\n1.1. Покупатель: ООО «Тест». Поставщик: ООО «Новый».\n2.1. Общая стоимость поставки составляет 25 000 рублей.\n3.1. Поставщик обязуется поставить оборудование не позднее 15.05.2026."})
	result := NewService().Analyze(files, asOf())
	if len(result.Contracts) != 4 {
		t.Fatalf("contracts = %d", len(result.Contracts))
	}
	for _, c := range result.Contracts {
		if c.ID == "STN-102" && c.Amount.Value != "910000" {
			t.Fatalf("mutated amount: %s", c.Amount.Value)
		}
		if c.ID == "STN-104" && (c.Amount.Value != "25000" || c.Deadline.Value != "2026-05-15") {
			t.Fatalf("new contract: %+v", c)
		}
	}
}

func TestUploadOrderDoesNotOverrideEffectiveDate(t *testing.T) {
	files := fixtureFiles(t)
	for left, right := 0, len(files)-1; left < right; left, right = left+1, right-1 {
		files[left], files[right] = files[right], files[left]
	}
	result := NewService().Analyze(files, asOf())
	for _, contract := range result.Contracts {
		if contract.ID == "STN-101" && (contract.Amount.Value != "1380000" || contract.Deadline.Value != "2026-04-30") {
			t.Fatalf("wrong effective terms after reorder: %+v", contract)
		}
	}
}

func TestAmbiguousSignedAmendmentsRequireReview(t *testing.T) {
	files := []InputFile{
		{Name: "contract.txt", Body: "ДОГОВОР ПОСТАВКИ № STN-200\nСтатус: подписан обеими сторонами 01.02.2026.\n2.1. Общая стоимость поставки составляет 100 000 рублей.\n3.1. Поставщик обязуется поставить оборудование не позднее 30.04.2026."},
		{Name: "first.txt", Body: "ДОПОЛНИТЕЛЬНОЕ СОГЛАШЕНИЕ К ДОГОВОРУ STN-200\nСтатус: подписано обеими сторонами 01.03.2026.\n1. Пункт 2.1 договора изложить в следующей редакции: «Общая стоимость поставки составляет 120 000 рублей».\n2. Настоящее соглашение вступает в силу 01.03.2026."},
		{Name: "second.txt", Body: "ДОПОЛНИТЕЛЬНОЕ СОГЛАШЕНИЕ К ДОГОВОРУ STN-200\nСтатус: подписано обеими сторонами 01.03.2026.\n1. Пункт 2.1 договора изложить в следующей редакции: «Общая стоимость поставки составляет 130 000 рублей».\n2. Настоящее соглашение вступает в силу 01.03.2026."},
	}
	result := NewService().Analyze(files, asOf())
	contract := result.Contracts[0]
	if !contract.Review || contract.Amount.Value != "" || len(contract.Notices) == 0 {
		t.Fatalf("ambiguous amount should require review: %+v", contract)
	}
}

func TestUnparsedSignedChangeRequiresReview(t *testing.T) {
	files := []InputFile{
		{Name: "contract.txt", Body: "ДОГОВОР ПОСТАВКИ № STN-201\nСтатус: подписан обеими сторонами 01.02.2026.\n2.1. Общая стоимость поставки составляет 100 000 рублей.\n3.1. Поставщик обязуется поставить оборудование не позднее 30.04.2026."},
		{Name: "amendment.txt", Body: "ДОПОЛНИТЕЛЬНОЕ СОГЛАШЕНИЕ К ДОГОВОРУ STN-201\nСтатус: подписано обеими сторонами 01.03.2026.\n1. Пункт 2.1 договора изложить в новой редакции согласно приложению.\n2. Настоящее соглашение вступает в силу 01.03.2026."},
	}
	contract := NewService().Analyze(files, asOf()).Contracts[0]
	if !contract.Review || contract.Amount.Value != "" || len(contract.Notices) == 0 {
		t.Fatalf("missing review: %+v", contract)
	}
}

func TestSignedAfterCheckDateDoesNotApply(t *testing.T) {
	files := []InputFile{
		{Name: "contract.txt", Body: "ДОГОВОР ПОСТАВКИ № STN-202\nСтатус: подписан обеими сторонами 01.02.2026.\n2.1. Общая стоимость поставки составляет 100 000 рублей.\n3.1. Поставщик обязуется поставить оборудование не позднее 30.04.2026."},
		{Name: "amendment.txt", Body: "ДОПОЛНИТЕЛЬНОЕ СОГЛАШЕНИЕ К ДОГОВОРУ STN-202\nСтатус: подписано обеими сторонами 15.03.2026.\n1. Пункт 2.1 договора изложить в следующей редакции: «Общая стоимость поставки составляет 120 000 рублей».\n2. Настоящее соглашение вступает в силу 01.03.2026."},
	}
	contract := NewService().Analyze(files, asOf()).Contracts[0]
	if contract.Amount.Value != "100000" || contract.Amount.Changed || len(contract.Notices) != 1 || contract.Notices[0].Kind != "pending" {
		t.Fatalf("future signature changed current terms: %+v", contract)
	}
}

func TestNonRubleAmountIsNotPresentedAsRubles(t *testing.T) {
	files := []InputFile{{Name: "contract.txt", Body: "ДОГОВОР ПОСТАВКИ № STN-203\nСтатус: подписан обеими сторонами 01.02.2026.\n2.1. Общая стоимость поставки составляет 100 000 долларов.\n3.1. Поставщик обязуется поставить оборудование не позднее 30.04.2026."}}
	contract := NewService().Analyze(files, asOf()).Contracts[0]
	if contract.Amount.Value != "" {
		t.Fatalf("non-ruble amount: %+v", contract.Amount)
	}
}
