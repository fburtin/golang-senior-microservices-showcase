package customerdebt

import (
	"context"
	"fmt"

	bcra "github.com/fburtin/golang-senior-microservices-showcase/internal/bcra"
)

type BCRADebtProvider interface {
	GetDebts(
		ctx context.Context,
		cuit string,
	) (*bcra.DebtResponse, error)
}

type Service struct {
	bcraClient BCRADebtProvider
	repository Repository
}

func NewService(
	bcraClient BCRADebtProvider,
	repository Repository,
) *Service {
	return &Service{
		bcraClient: bcraClient,
		repository: repository,
	}
}

func (s *Service) GetAndSave(
	ctx context.Context,
	cuit string,
) (*bcra.DebtResponse, error) {
	response, err := s.bcraClient.GetDebts(ctx, cuit)
	if err != nil {
		return nil, err
	}

	if err := s.repository.Upsert(
		ctx,
		cuit,
		response.Results,
	); err != nil {
		return nil, fmt.Errorf("save BCRA debt response: %w", err)
	}

	return response, nil
}
