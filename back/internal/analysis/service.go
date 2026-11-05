package analysis

import (
	"strconv"
	"time"

	"contractcheck/back/internal/domain"
)

type InputFile struct {
	Name string
	Body string
}

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) Analyze(files []InputFile, asOf time.Time) domain.Result {
	docs := make([]domain.Document, 0, len(files))
	for i, file := range files {
		docs = append(docs, ParseDocument("d"+strconv.Itoa(i+1), file.Name, file.Body))
	}
	return Evaluate(docs, asOf)
}
