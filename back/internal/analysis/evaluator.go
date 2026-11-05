package analysis

import (
	"contractcheck/back/internal/domain"
	"sort"
	"time"
)

type Contract = domain.Contract
type Notice = domain.Notice
type Result = domain.Result

type appliedField struct {
	date       time.Time
	value      string
	conflicted bool
}

func Evaluate(docs []Document, asOf time.Time) Result {
	res := Result{AsOf: asOf.Format("2006-01-02"), Documents: docs, Contracts: []Contract{}, Warnings: []string{}}
	byID := map[string]*Contract{}
	for _, d := range docs {
		if d.Kind == "unknown" {
			res.Warnings = append(res.Warnings, "Не распознан тип документа: "+d.Name)
		}
		if d.Kind != "contract" {
			continue
		}
		if d.ContractID == "" {
			res.Warnings = append(res.Warnings, "Не найден номер договора: "+d.Name)
			continue
		}
		if _, ok := byID[d.ContractID]; ok {
			res.Warnings = append(res.Warnings, "Повторный основной договор "+d.ContractID+": требуется проверка")
			byID[d.ContractID].Review = true
			continue
		}
		c := &Contract{ID: d.ContractID, Supplier: extractSupplier(d), Amount: extractAmount(d), Deadline: extractDeadline(d), Notices: []Notice{}}
		if d.Status != "signed" {
			c.Review = true
			c.Notices = append(c.Notices, Notice{Kind: "review", Message: "Подписание основного договора не подтверждено", Evidence: firstEvidence(d)})
		}
		byID[d.ContractID] = c
	}
	// Signed amendments are applied by their effective date, independent of upload order and file name.
	amendments := []Document{}
	for _, d := range docs {
		if d.Kind == "amendment" {
			amendments = append(amendments, d)
		}
	}
	sort.SliceStable(amendments, func(i, j int) bool { return effectiveDate(amendments[i]).Before(effectiveDate(amendments[j])) })
	applied := map[string]appliedField{}
	for _, d := range amendments {
		c := byID[d.ContractID]
		if c == nil {
			res.Warnings = append(res.Warnings, "Документ "+d.Name+" не связан с загруженным договором")
			continue
		}
		a, deadline := extractAmount(d), extractDeadline(d)
		if d.Status != "signed" {
			label := "Неподписанный проект изменения"
			if d.Status != "draft" {
				label = "Статус соглашения не подтвержден"
				c.Review = true
			}
			appendChangeNotice(c, a, deadline, label, "proposal")
			continue
		}
		signed := signedAt(d)
		if signed.IsZero() {
			c.Review = true
			c.Notices = append(c.Notices, Notice{Kind: "review", Message: "Не найдена дата подписания соглашения", Evidence: firstEvidence(d)})
			continue
		}
		if signed.After(asOf) {
			appendChangeNotice(c, a, deadline, "Соглашение подписано после даты проверки", "pending")
			continue
		}
		eff := effectiveDate(d)
		if eff.IsZero() {
			c.Review = true
			c.Notices = append(c.Notices, Notice{Kind: "review", Message: "У подписанного соглашения не найдена дата вступления в силу", Evidence: firstEvidence(d)})
			continue
		}
		if eff.After(asOf) {
			appendChangeNotice(c, a, deadline, "Подписанное изменение еще не вступило в силу", "pending")
			continue
		}
		if a.Value == "" {
			if clause := changeClause(d, "amount"); clause != nil {
				c.Review = true
				c.Amount = Field{}
				c.Notices = append(c.Notices, Notice{Kind: "review", Message: "Не удалось извлечь новую сумму из подписанного соглашения", Evidence: *clause})
			}
		}
		if deadline.Value == "" {
			if clause := changeClause(d, "deadline"); clause != nil {
				c.Review = true
				c.Deadline = Field{}
				c.Notices = append(c.Notices, Notice{Kind: "review", Message: "Не удалось извлечь новый срок из подписанного соглашения", Evidence: *clause})
			}
		}
		if a.Value != "" {
			applyField(c, &c.Amount, a, eff, "amount", applied)
		}
		if deadline.Value != "" {
			applyField(c, &c.Deadline, deadline, eff, "deadline", applied)
		}
	}
	for _, d := range docs {
		if d.Kind != "email" {
			continue
		}
		c := byID[d.ContractID]
		if c == nil {
			res.Warnings = append(res.Warnings, "Документ "+d.Name+" не связан с загруженным договором")
			continue
		}
		appendChangeNotice(c, Field{}, extractProposedDeadline(d), "Предложение поставщика, не согласовано", "proposal")
	}
	for _, c := range byID {
		res.Contracts = append(res.Contracts, *c)
	}
	sort.Slice(res.Contracts, func(i, j int) bool { return res.Contracts[i].ID < res.Contracts[j].ID })
	return res
}

func applyField(c *Contract, target *Field, next Field, effective time.Time, kind string, applied map[string]appliedField) {
	key := c.ID + ":" + kind
	previous, exists := applied[key]
	if exists && previous.date.Equal(effective) {
		if previous.conflicted {
			return
		}
		if previous.value == next.Value {
			return
		}
		c.Review = true
		*target = Field{}
		message := "Конфликт подписанных соглашений с одинаковой датой вступления в силу: сумма"
		if kind == "deadline" {
			message = "Конфликт подписанных соглашений с одинаковой датой вступления в силу: срок поставки"
		}
		c.Notices = append(c.Notices, Notice{Kind: "review", Message: message, Evidence: *next.Evidence})
		applied[key] = appliedField{date: effective, conflicted: true}
		return
	}
	next.Changed = true
	*target = next
	applied[key] = appliedField{date: effective, value: next.Value}
}

func appendChangeNotice(c *Contract, amount, deadline Field, prefix, kind string) {
	if amount.Value != "" {
		c.Notices = append(c.Notices, Notice{Kind: kind, Message: prefix + ": сумма " + amount.Value + " ₽", Evidence: *amount.Evidence})
	}
	if deadline.Value != "" {
		c.Notices = append(c.Notices, Notice{Kind: kind, Message: prefix + ": срок поставки " + deadline.Value, Evidence: *deadline.Evidence})
	}
}
