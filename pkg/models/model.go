package models

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Spell struct {
	ID           int
	Name         string
	Description  string
	School       string
	Type         JSONStringSlice
	Target       string
	Descriptors  JSONStringSlice
	Duration     string
	Level        string
	Requirements JSONStringSlice
	SavingThrow  string
	ActionType   string
	SpellResist  string
	Metamagic    JSONStringSlice
	Crafting     JSONStringSlice
	Image        string
}

type SessionID func(*http.Request) string

type Filter struct {
	Column string
	Value  string
}

type ColumnInfo struct {
	Name       string
	Filterable bool
	Sortable   bool
}

type SpellViewData struct {
	Spells         []Spell
	Filters        []Filter
	Language       string
	AvailableLangs []string
	SearchQuery    string
	SortColumn     string
	SortOrder      string
	ReturnTo       string
	FilterOptions  map[string][]string
	LangData       struct {
		Language       string
		ReturnTo       string
		AvailableLangs []string
	}
	Columns []ColumnInfo
}

type JSONStringSlice []string

func (j *JSONStringSlice) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("неподдерживаемый тип для JSONStringSlice")
	}

	return json.Unmarshal(bytes, j)
}
