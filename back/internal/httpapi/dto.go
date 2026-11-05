package httpapi

import "contractcheck/back/internal/domain"

// Response DTOs keep the public JSON contract separate from the domain model.
type paragraphDTO struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type documentDTO struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Kind       string         `json:"kind"`
	ContractID string         `json:"contract_id"`
	Status     string         `json:"status"`
	Paragraphs []paragraphDTO `json:"paragraphs"`
}

type evidenceDTO struct {
	DocumentID  string `json:"document_id"`
	FileName    string `json:"file_name"`
	ParagraphID string `json:"paragraph_id"`
	Quote       string `json:"quote"`
}

type fieldDTO struct {
	Value    string       `json:"value"`
	Evidence *evidenceDTO `json:"evidence,omitempty"`
	Changed  bool         `json:"changed"`
}

type noticeDTO struct {
	Kind     string      `json:"kind"`
	Message  string      `json:"message"`
	Evidence evidenceDTO `json:"evidence"`
}

type contractDTO struct {
	ID       string      `json:"id"`
	Supplier fieldDTO    `json:"supplier"`
	Amount   fieldDTO    `json:"amount"`
	Deadline fieldDTO    `json:"deadline"`
	Notices  []noticeDTO `json:"notices"`
	Review   bool        `json:"review"`
}

type analyzeResponse struct {
	AsOf      string        `json:"as_of"`
	Contracts []contractDTO `json:"contracts"`
	Documents []documentDTO `json:"documents"`
	Warnings  []string      `json:"warnings"`
}

func toResponse(result domain.Result) analyzeResponse {
	response := analyzeResponse{AsOf: result.AsOf, Contracts: make([]contractDTO, 0, len(result.Contracts)), Documents: make([]documentDTO, 0, len(result.Documents)), Warnings: result.Warnings}
	for _, c := range result.Contracts {
		item := contractDTO{ID: c.ID, Supplier: toField(c.Supplier), Amount: toField(c.Amount), Deadline: toField(c.Deadline), Notices: make([]noticeDTO, 0, len(c.Notices)), Review: c.Review}
		for _, n := range c.Notices {
			item.Notices = append(item.Notices, noticeDTO{Kind: n.Kind, Message: n.Message, Evidence: toEvidence(n.Evidence)})
		}
		response.Contracts = append(response.Contracts, item)
	}
	for _, d := range result.Documents {
		item := documentDTO{ID: d.ID, Name: d.Name, Kind: d.Kind, ContractID: d.ContractID, Status: d.Status, Paragraphs: make([]paragraphDTO, 0, len(d.Paragraphs))}
		for _, p := range d.Paragraphs {
			item.Paragraphs = append(item.Paragraphs, paragraphDTO{ID: p.ID, Text: p.Text})
		}
		response.Documents = append(response.Documents, item)
	}
	return response
}

func toField(f domain.Field) fieldDTO {
	item := fieldDTO{Value: f.Value, Changed: f.Changed}
	if f.Evidence != nil {
		e := toEvidence(*f.Evidence)
		item.Evidence = &e
	}
	return item
}

func toEvidence(e domain.Evidence) evidenceDTO {
	return evidenceDTO{DocumentID: e.DocumentID, FileName: e.FileName, ParagraphID: e.ParagraphID, Quote: e.Quote}
}
