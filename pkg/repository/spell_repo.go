package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"pathfinder-wiki/configs"
	"pathfinder-wiki/pkg/models"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type SpellRepository interface {
	GetSpells(ctx context.Context, lang, search string, filters []models.Filter, sortColumn, sortOrder string) ([]models.Spell, error)
	GetFilterOptions(ctx context.Context, lang string) (map[string][]string, error)
	InsertLogs(r *http.Request, ses models.SessionID)
}

type spellRepository struct {
	db *sql.DB
}

func NewSpellRepository(db *sql.DB) SpellRepository {
	return &spellRepository{db: db}
}

func NewPostgresDB(cfg configs.DBConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func (r *spellRepository) buildQuery(lang, search string, filters []models.Filter, sortColumn, sortOrder string) (string, []interface{}) {
	var where []string
	var args []interface{}
	argCounter := 1

	// Search
	if search != "" {
		searchColumns := []string{"s.\"Name\"", "s.\"Description\""}
		var searchConditions []string
		for _, col := range searchColumns {
			searchConditions = append(searchConditions, fmt.Sprintf("%s LIKE $%d", col, argCounter))
			args = append(args, "%"+search+"%")
			argCounter++
		}
		where = append(where, "("+strings.Join(searchConditions, " OR ")+")")
	}

	// Group by columns
	filterGroups := make(map[string][]string)
	for _, filter := range filters {
		filterGroups[filter.Column] = append(filterGroups[filter.Column], filter.Value)
	}

	// Generate conditions
	for column, values := range filterGroups {
		if len(values) == 0 {
			continue
		}

		isJSON := column == "TypeJSON" || column == "DescriptorsJSON" ||
			column == "RequirementsJSON" || column == "MetamagicJSON" ||
			column == "CraftingJSON"

		var conditions []string

		for _, value := range values {
			if isJSON {
				conditions = append(conditions,
					fmt.Sprintf("i.\"%s\"::jsonb @> $%d::jsonb", column, argCounter))
				args = append(args, fmt.Sprintf(`["%s"]`, value))
			} else {
				conditions = append(conditions,
					fmt.Sprintf("i.\"%s\" = $%d", column, argCounter))
				args = append(args, value)
			}
			argCounter++
		}

		if len(conditions) > 0 {
			where = append(where, "("+strings.Join(conditions, " OR ")+")")
		}
	}

	query := fmt.Sprintf(`
        SELECT 
            s."id", s."Name", s."Description", i."School", i."TypeJSON", i."Target", 
            i."DescriptorsJSON", i."Duration", i."Level", i."RequirementsJSON", 
            i."Saving_Throw", i."Action_Type", i."Spell_Resist", 
            i."MetamagicJSON", i."CraftingJSON", i."Image"  
        FROM "Spells_%s" s 
        LEFT JOIN "Spells_Info" i ON s."id" = i."spell_id"
    `, lang)

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	if sortColumn != "" {
		query += " ORDER BY " + "\"" + sortColumn + "\""
		if sortOrder == "desc" {
			query += " DESC"
		}
	}

	return query, args
}

func (r *spellRepository) GetSpells(ctx context.Context, lang, search string, filters []models.Filter, sortColumn, sortOrder string) ([]models.Spell, error) {
	query, args := r.buildQuery(lang, search, filters, sortColumn, sortOrder)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query spells: %w", err)
	}
	defer rows.Close()

	var spells []models.Spell
	for rows.Next() {
		var s models.Spell
		err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.School, &s.Type, &s.Target,
			&s.Descriptors, &s.Duration, &s.Level, &s.Requirements, &s.SavingThrow,
			&s.ActionType, &s.SpellResist, &s.Metamagic, &s.Crafting, &s.Image)
		if err != nil {
			return nil, fmt.Errorf("failed to scan spell: %w", err)
		}
		s.Image, err = url.QueryUnescape("/static/images/" + s.Image + ".webp")
		if err != nil {
			return nil, fmt.Errorf("invalid image path %w", err)
		}
		spells = append(spells, s)
	}

	return spells, nil
}

func uniqueStrings(input []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range input {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func (r *spellRepository) GetFilterOptions(ctx context.Context, lang string) (map[string][]string, error) {
	columns := map[string]string{
		"school":       "i.\"School\"",
		"type":         "i.\"TypeJSON\"",
		"target":       "i.\"Target\"",
		"descriptors":  "i.\"DescriptorsJSON\"",
		"duration":     "i.\"Duration\"",
		"level":        "i.\"Level\"::text",
		"requirements": "i.\"RequirementsJSON\"",
		"savingThrow":  "i.\"Saving_Throw\"",
		"action_type":  "i.\"Action_Type\"",
		"spellResist":  "i.\"Spell_Resist\"",
		"metamagic":    "i.\"MetamagicJSON\"",
		"crafting":     "i.\"CraftingJSON\"",
	}
	options := make(map[string][]string)

	for colName, col := range columns {
		query := fmt.Sprintf(`
            SELECT %s
            FROM "Spells_%s" s 
            LEFT JOIN "Spells_Info" i ON s."id" = i."spell_id" 
            WHERE %s IS NOT NULL AND %s != ''
            ORDER BY %s
        `, col, lang, col, col, col)

		rows, err := r.db.Query(query)
		if err != nil {
			fmt.Printf("Error getting filter options for %s: %v", colName, err)
			continue
		}
		defer rows.Close()

		var values []string
		for rows.Next() {
			var val string
			if err := rows.Scan(&val); err == nil && val != "" {
				if col == "i.\"TypeJSON\"" || col == "i.\"DescriptorsJSON\"" ||
					col == "i.\"RequirementsJSON\"" || col == "i.\"MetamagicJSON\"" ||
					col == "i.\"CraftingJSON\"" {
					var jsonArr []string
					if err := json.Unmarshal([]byte(val), &jsonArr); err == nil {
						values = append(values, jsonArr...)
					}
				} else {
					values = append(values, val)
				}
			}
		}
		values = uniqueStrings(values)
		options[colName] = values
	}

	return options, nil
}

func (repo *spellRepository) InsertLogs(r *http.Request, ses models.SessionID) {
	_, err := repo.db.Exec(`
    		INSERT INTO "Visit_Info" 
    		("Ip", "User_Agent", "Referer", "Path", "Visit_Time", "Session_id") 
    		VALUES ($1, $2, $3, $4, $5, $6)`,
		r.RemoteAddr,
		r.UserAgent(),
		r.Referer(),
		r.URL.Path,
		time.Now(),
		ses(r),
	)
	if err != nil {
		fmt.Printf("Ошибка записи посещения: %v", err)
	}
}
