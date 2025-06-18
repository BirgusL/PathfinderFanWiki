package main

import (
	"log"
	"net/http"
	"pathfinder-wiki/configs"
	"pathfinder-wiki/pkg/handler"
	"pathfinder-wiki/pkg/repository"
	"pathfinder-wiki/pkg/service"
)

func main() {
	cfg := configs.Load()

	db, err := repository.NewPostgresDB(cfg.DB)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	spellRepo := repository.NewSpellRepository(db)
	spellService := service.NewSpellService(spellRepo)

	h := handler.NewHandler(spellService)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: h,
	}

	log.Printf("Server running on http://localhost:%s", cfg.Port)
	log.Fatal(server.ListenAndServe())
}
