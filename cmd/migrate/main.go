package main

import (
	"log"
	"vue-golang/internal/config"
	"vue-golang/internal/migration"
	"vue-golang/internal/storage/mysql"
)

func main() {
	cfg := config.MustConfig()

	log.Println("start migration")

	//err := migration.Up("./migrations", mysql.MigrateDSN(*cfg))
	err := migration.Up(cfg.MigrationPath, mysql.MigrateDSN(*cfg))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("migration completed")
}
