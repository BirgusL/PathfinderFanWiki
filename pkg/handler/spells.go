package handler

import (
	"net/http"
	"pathfinder-wiki/pkg/models"
	"pathfinder-wiki/pkg/service"
	"strings"
	"text/template"
)

type Handler struct {
	spellService service.SpellService
	tmpl         *template.Template
	mux          *http.ServeMux
}

func NewHandler(spellService service.SpellService) *Handler {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"inFilter": inFilter,
	}).ParseGlob("templates/*.html"))

	h := &Handler{
		spellService: spellService,
		tmpl:         tmpl,
		mux:          http.NewServeMux(),
	}

	fs := http.FileServer(http.Dir("static"))
	h.mux.Handle("/static/", http.StripPrefix("/static/", fs))

	h.mux.HandleFunc("/", h.homeHandler)
	h.mux.HandleFunc("/kingmaker", h.kingmakerHandler)
	h.mux.HandleFunc("/kingmaker/spells", h.spellsHandler)

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func inFilter(filters []models.Filter, column, value string) bool {
	for _, f := range filters {
		if f.Column == column && f.Value == value {
			return true
		}
	}
	return false
}

func getLanguage(r *http.Request) string {
	if cookie, err := r.Cookie("lang"); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	if lang := r.URL.Query().Get("lang"); lang != "" {
		return lang
	}

	return "EN"
}

func ParseFilters(r *http.Request) []models.Filter {
	var filters []models.Filter

	for key, values := range r.URL.Query() {
		if strings.HasPrefix(key, "filter_") {
			column := strings.TrimPrefix(key, "filter_")
			for _, value := range values {
				if value != "" {
					for _, v := range strings.Split(value, ",") {
						if v != "" {
							filters = append(filters, models.Filter{
								Column: column,
								Value:  v,
							})
						}
					}
				}
			}
		}
	}

	return filters
}

func (h *Handler) spellsHandler(w http.ResponseWriter, r *http.Request) {
	lang := getLanguage(r)
	searchQuery := r.FormValue("search")
	sortColumn := r.FormValue("sort")
	sortOrder := r.FormValue("order")
	filters := ParseFilters(r)

	spells, err := h.spellService.GetSpells(r.Context(), lang, searchQuery, filters, sortColumn, sortOrder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filterOptions, err := h.spellService.GetFilterOptions(r.Context(), lang)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := models.SpellViewData{
		Spells:         spells,
		Language:       lang,
		AvailableLangs: []string{"EN", "RU", "DE", "FR"},
		SearchQuery:    searchQuery,
		SortColumn:     sortColumn,
		SortOrder:      sortOrder,
		Filters:        filters,
		FilterOptions:  filterOptions,
		Columns: []models.ColumnInfo{
			{Name: "Name", Filterable: false, Sortable: true},
			{Name: "Description", Filterable: false},
			{Name: "School", Filterable: true, Sortable: true},
			{Name: "Type", Filterable: true, Sortable: false},
			{Name: "Target", Filterable: true},
			{Name: "Descriptors", Filterable: true},
			{Name: "Duration", Filterable: true, Sortable: true},
			{Name: "Level", Filterable: true, Sortable: true},
			{Name: "Requirements", Filterable: true},
			{Name: "Saving_Throw", Filterable: true, Sortable: true},
			{Name: "Action_Type", Filterable: true, Sortable: true},
			{Name: "Spell_Resist", Filterable: true, Sortable: true},
			{Name: "Metamagic", Filterable: true},
			{Name: "Crafting", Filterable: true},
		},
	}

	h.tmpl.ExecuteTemplate(w, "spells.html", data)
}

func (h *Handler) homeHandler(w http.ResponseWriter, r *http.Request) {
	lang := getLanguage(r)

	data := map[string]interface{}{
		"LangData": map[string]interface{}{
			"Language":       lang,
			"ReturnTo":       r.URL.Path,
			"AvailableLangs": []string{"EN", "RU", "DE", "FR"},
		},
	}

	h.tmpl.ExecuteTemplate(w, "home.html", data)
}

func (h *Handler) kingmakerHandler(w http.ResponseWriter, r *http.Request) {
	lang := getLanguage(r)

	data := map[string]interface{}{
		"LangData": map[string]interface{}{
			"Language":       lang,
			"ReturnTo":       r.URL.Path,
			"AvailableLangs": []string{"EN", "RU", "DE", "FR"},
		},
	}
	h.tmpl.ExecuteTemplate(w, "kingmaker.html", data)
}

func (h *Handler) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/favicon.ico" || r.URL.Path == "/static/" {
			next.ServeHTTP(w, r)
			return
		}

		h.spellService.InsertLogs(r, getSessionID)

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
