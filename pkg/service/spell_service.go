package service

import (
	"context"
	"pathfinder-wiki/pkg/models"
	"pathfinder-wiki/pkg/repository"
)

type SpellService interface {
	GetSpells(ctx context.Context, lang, search string, filters []models.Filter, sortColumn, sortOrder string) ([]models.Spell, error)
	GetFilterOptions(ctx context.Context, lang string) (map[string][]string, error)
}

type spellService struct {
	repo repository.SpellRepository
}

func NewSpellService(repo repository.SpellRepository) SpellService {
	return &spellService{repo: repo}
}

func (s *spellService) GetSpells(ctx context.Context, lang, search string, filters []models.Filter, sortColumn, sortOrder string) ([]models.Spell, error) {
	return s.repo.GetSpells(ctx, lang, search, filters, sortColumn, sortOrder)
}

func (s *spellService) GetFilterOptions(ctx context.Context, lang string) (map[string][]string, error) {
	return s.repo.GetFilterOptions(ctx, lang)
}
