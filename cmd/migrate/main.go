package migrate

import (
	"log"
	"vue-golang/internal/config"
	"vue-golang/internal/migrate"
	"vue-golang/internal/storage/mysql"
)

func main() {
	cfg := config.MustConfig()

	err := migrate.Up("migrations", mysql.MigrateDSN(*cfg))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("migration completed")
}
