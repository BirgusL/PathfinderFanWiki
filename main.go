package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var tmpl *template.Template

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

type Filter struct {
	Column string
	Value  string
}

type SpellViewData struct {
	Spells         []Spell
	Filters        map[string][]Filter
	Language       string
	AvailableLangs []string
	SearchQuery    string
	SortColumn     string
	SortOrder      string
	ReturnTo       string
}

type JSONStringSlice []string

// Scan реализует интерфейс sql.Scanner
func (j *JSONStringSlice) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	// Получаем данные из БД (может быть []byte или string)
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("неподдерживаемый тип для JSONStringSlice")
	}

	// Десериализуем JSON
	return json.Unmarshal(bytes, j)
}

func (j JSONStringSlice) String() string {
	if len(j) == 0 {
		return "" // или "none", если нужно специальное значение для пустого списка
	}
	return strings.Join(j, ", ")
}

func main() {
	var err error
	// Initialize database
	db, err = sql.Open("sqlite3", "./database/spells.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	// Initialize templates
	tmpl = template.Must(template.New("").Funcs(template.FuncMap{
		"translate": translate, // Добавляем функцию translate
		"inFilter":  inFilter,  // Добавляем новую функцию
	}).ParseGlob("templates/*.html"))

	// Create static file server
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/kingmaker", kingmakerHandler)
	mux.HandleFunc("/kingmaker/spells", spellsHandler)
	mux.HandleFunc("/set-language", languageHandler)

	var handler http.Handler = mux
	handler = loggingMiddleware(db, handler)

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	// Start server
	fmt.Println("Server running on http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, handler))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	//log.Println("Home handler executed")

	lang := getLanguage(r)

	if lang == "" {
		lang = "en" // гарантированное значение
		http.SetCookie(w, &http.Cookie{
			Name:  "lang",
			Value: lang,
			Path:  "/",
		})
	}

	data := map[string]interface{}{
		"Language":       lang,
		"ReturnTo":       r.URL.Path,
		"AvailableLangs": []string{"EN", "RU", "DE", "FR"},
	}

	//log.Printf("Template data: %+v", data)

	if err := tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		log.Println("Render error:", err) // Выведет ошибку в консоль
		http.Error(w, "Server error", 500)
	}
}

func kingmakerHandler(w http.ResponseWriter, r *http.Request) {
	lang := getLanguage(r)

	if lang == "" {
		lang = "en" // гарантированное значение
		http.SetCookie(w, &http.Cookie{
			Name:  "lang",
			Value: lang,
			Path:  "/",
		})
	}

	data := map[string]interface{}{
		"Language":       lang,
		"ReturnTo":       r.URL.Path,
		"AvailableLangs": []string{"EN", "RU", "DE", "FR"},
	}
	tmpl.ExecuteTemplate(w, "kingmaker.html", data)
}

func spellsHandler(w http.ResponseWriter, r *http.Request) {
	lang := getLanguage(r)
	if lang == "" {
		lang = "en"
		http.SetCookie(w, &http.Cookie{
			Name:  "lang",
			Value: lang,
			Path:  "/",
		})
	}

	// Получаем параметры
	searchQuery := r.FormValue("search")
	sortColumn := r.FormValue("sort")
	sortOrder := r.FormValue("order")
	filters := parseFilters(r)

	// Строим запрос
	query, args := buildQuery(lang, searchQuery, filters, sortColumn, sortOrder)

	// Выполняем запрос
	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var spells []Spell
	for rows.Next() {
		var s Spell
		err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.School, &s.Type, &s.Target,
			&s.Descriptors, &s.Duration, &s.Level, &s.Requirements, &s.SavingThrow,
			&s.ActionType, &s.SpellResist, &s.Metamagic, &s.Crafting, &s.Image)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.Image, err = url.QueryUnescape("/static/images/" + s.Image + ".webp")
		if err != nil {
			http.Error(w, "Invalid image path", http.StatusBadRequest)
			return
		}

		spells = append(spells, s)
	}

	// Получаем уникальные значения для фильтров
	filterOptions := getFilterOptions(lang)

	data := map[string]interface{}{
		"Spells":         spells,
		"Language":       lang,
		"AvailableLangs": []string{"EN", "RU", "DE", "FR"},
		"SearchQuery":    searchQuery,
		"SortColumn":     sortColumn,
		"SortOrder":      sortOrder,
		"Filters":        filters,
		"FilterOptions":  filterOptions,
		"Columns": []ColumnInfo{
			{Name: "Name", Filterable: false, Sortable: true},
			{Name: "Description", Filterable: false},
			{Name: "School", Filterable: true, Sortable: true},
			{Name: "Type", Filterable: true, Sortable: false},
			{Name: "Target", Filterable: true},
			{Name: "Descriptors", Filterable: true},
			{Name: "Duration", Filterable: true, Sortable: true},
			{Name: "Level", Filterable: true, Sortable: true},
			{Name: "Requirements", Filterable: true},
			{Name: "SavingThrow", Filterable: true, Sortable: true},
			{Name: "ActionType", Filterable: true, Sortable: true},
			{Name: "SpellResist", Filterable: true, Sortable: true},
			{Name: "Metamagic", Filterable: true},
			{Name: "Crafting", Filterable: true},
		},
	}

	tmpl.ExecuteTemplate(w, "spells.html", data)
}

type ColumnInfo struct {
	Name       string
	Filterable bool
	Sortable   bool
}

func parseFilters(r *http.Request) []Filter {
	var filters []Filter

	for key, values := range r.URL.Query() {
		if strings.HasPrefix(key, "filter_") {
			column := strings.TrimPrefix(key, "filter_")
			// Обрабатываем несколько значений через запятую
			if len(values) > 0 {
				for _, value := range strings.Split(values[0], ",") {
					if value != "" {
						filters = append(filters, Filter{
							Column: column,
							Value:  value,
						})
					}
				}
			}
		}
	}

	return filters
}

func languageHandler(w http.ResponseWriter, r *http.Request) {
	lang := r.FormValue("lang")
	returnTo := r.FormValue("return_to")
	if returnTo == "" {
		returnTo = "/"
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "lang",
		Value: lang,
		Path:  "/",
	})

	http.Redirect(w, r, returnTo, http.StatusSeeOther)
}

func getLanguage(r *http.Request) string {
	// 1. Проверяем куки
	if cookie, err := r.Cookie("lang"); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// 2. Проверяем URL-параметр
	if lang := r.URL.Query().Get("lang"); lang != "" {
		return lang
	}

	// 3. Значение по умолчанию
	return "EN" // обязательно возвращаем значение по умолчанию
}

func translate(lang, key string) string {
	// Простая реализация - можно заменить на более сложную логику
	translations := map[string]map[string]string{
		"Kingmaker": {
			"EN": "Kingmaker",
			"RU": "Король-изгнанник",
			"DE": "Königsmacher",
			"FR": "Faiseur de Rois",
		},
		"Wrath of the Righteous": {
			"EN": "Wrath of the Righteous",
			"RU": "Гнев праведников",
			"DE": "Zorn der Gerechten",
			"FR": "Colère des Justes",
		},
		"Conjuration": {
			"EN": "Conjuration",
			"RU": "Заклинание",
			"DE": "Zorn der Gerechten",
			"FR": "Colère des Justes",
		},
	}

	if tr, ok := translations[key]; ok {
		if val, ok := tr[lang]; ok {
			return val
		}
	}
	return key // Возвращаем ключ, если перевод не найден
}

func buildQuery(lang, search string, filters []Filter, sortColumn, sortOrder string) (string, []interface{}) {
	var where []string
	var args []interface{}
	argCounter := 1

	// Поиск
	if search != "" {
		searchColumns := []string{"s.Name", "s.Description"}
		var searchConditions []string
		for _, col := range searchColumns {
			searchConditions = append(searchConditions, fmt.Sprintf("%s LIKE $%d", col, argCounter))
			args = append(args, "%"+search+"%")
			argCounter++
		}
		where = append(where, "("+strings.Join(searchConditions, " OR ")+")")
	}

	// Группируем фильтры по колонкам
	filterGroups := make(map[string][]string)
	for _, filter := range filters {
		filterGroups[filter.Column] = append(filterGroups[filter.Column], filter.Value)
	}

	// Для каждой колонки создаем условие (AND между значениями)
	for column, values := range filterGroups {
		var conditions []string

		for _, value := range values {
			if column == "type" || column == "descriptors" ||
				column == "requirements" || column == "metamagic" ||
				column == "crafting" {
				// Для JSON-полей
				conditions = append(conditions, fmt.Sprintf("i.%sJSON LIKE $%d", strings.Title(column), argCounter))
				args = append(args, `%"`+value+`"%`)
			} else {
				// Для обычных полей
				conditions = append(conditions, fmt.Sprintf("i.%s = $%d", strings.Title(column), argCounter))
				args = append(args, value)
			}
			argCounter++
		}

		// Объединяем условия для одной колонки через AND
		if len(conditions) > 0 {
			where = append(where, "("+strings.Join(conditions, " AND ")+")")
		}
	}

	query := fmt.Sprintf(`
        SELECT 
            s.id, s.Name, s.Description, i.School, i.TypeJSON, i.Target, 
            i.DescriptorsJSON, i.Duration, i.Level, i.RequirementsJSON, 
            i.Saving_Throw, i.Action_Type, i.Spell_Resist, 
            i.MetamagicJSON, i.CraftingJSON, i.Image  
        FROM Spells_%s s 
        LEFT JOIN Spells_Info i ON s.id = i.spell_id
    `, lang)

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	if sortColumn != "" {
		query += " ORDER BY " + sortColumn
		if sortOrder == "desc" {
			query += " DESC"
		}
	}

	return query, args
}

func getFilterOptions(lang string) map[string][]string {
	columns := map[string]string{
		"school":       "i.School",
		"type":         "i.TypeJSON", // Исправлено с type на TypeJSON
		"target":       "i.Target",
		"descriptors":  "i.DescriptorsJSON",
		"duration":     "i.Duration",
		"level":        "i.Level",
		"requirements": "i.RequirementsJSON",
		"savingThrow":  "i.Saving_Throw",
		"action_type":  "i.Action_Type",
		"spellResist":  "i.Spell_Resist",
		"metamagic":    "i.MetamagicJSON",
		"crafting":     "i.CraftingJSON",
	}
	options := make(map[string][]string)

	for colName, col := range columns {
		query := fmt.Sprintf(`
            SELECT DISTINCT %s 
            FROM Spells_%s s 
            LEFT JOIN Spells_Info i ON s.id = i.spell_id 
            WHERE %s IS NOT NULL AND %s != ''
            ORDER BY %s
        `, col, lang, col, col, col)

		rows, err := db.Query(query)
		if err != nil {
			log.Printf("Error getting filter options for %s: %v", colName, err)
			continue
		}
		defer rows.Close()

		var values []string
		for rows.Next() {
			var val string
			if err := rows.Scan(&val); err == nil && val != "" {
				// Для JSON-полей нужно извлечь значения из массива
				if col == "i.TypeJSON" || col == "i.DescriptorsJSON" ||
					col == "i.RequirementsJSON" || col == "i.MetamagicJSON" ||
					col == "i.CraftingJSON" {
					var jsonArr []string
					if err := json.Unmarshal([]byte(val), &jsonArr); err == nil {
						values = append(values, jsonArr...)
					}
				} else {
					values = append(values, val)
				}
			}
		}
		// Удаляем дубликаты
		values = uniqueStrings(values)
		options[colName] = values
	}

	return options
}

// Вспомогательная функция для удаления дубликатов
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

func inFilter(filters []Filter, column, value string) bool {
	for _, f := range filters {
		if f.Column == column && f.Value == value {
			return true
		}
	}
	return false
}

func loggingMiddleware(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Пропускаем статические файлы
		if r.URL.Path == "/favicon.ico" || r.URL.Path == "/static/" {
			next.ServeHTTP(w, r)
			return
		}

		// Записываем посещение
		_, err := db.Exec(`
			INSERT INTO Visit_Info 
			(Ip, User_Agent, Referer, Path, Visit_time, Session_id) 
			VALUES (?, ?, ?, ?, ?, ?)`,
			r.RemoteAddr,
			r.UserAgent(),
			r.Referer(),
			r.URL.Path,
			time.Now(),
			getSessionID(r),
		)
		if err != nil {
			log.Printf("Ошибка записи посещения: %v", err)
		}

		next.ServeHTTP(w, r)
	})
}

func getSessionID(r *http.Request) string {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return ""
	}
	return cookie.Value
}
