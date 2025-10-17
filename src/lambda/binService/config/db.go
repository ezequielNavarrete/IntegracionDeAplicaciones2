package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB DBExecutor // variable global para la DB

func ConnectDatabase() {
	// load var from .env
	err := godotenv.Load()
	if err != nil {
		log.Println("No se encontró .env, usando variables de entorno existentes")
	}

	user := os.Getenv("MYSQL_USER")
	pass := os.Getenv("MYSQL_PASSWORD")
	host := os.Getenv("MYSQL_HOST")
	port := os.Getenv("MYSQL_PORT")
	dbname := os.Getenv("MYSQL_DATABASE")

	// Debug: mostrar variables cargadas (sin password)
	log.Printf("🔧 MySQL Config: user=%s, host=%s, port=%s, db=%s", user, host, port, dbname)

	if user == "" || host == "" || dbname == "" {
		log.Fatal("❌ Variables de MySQL no configuradas correctamente")
	}

	// build DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=skip-verify",
		user, pass, host, port, dbname)

	// Connect with GORM
	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ No se pudo conectar a la DB:", err)
	}

	fmt.Println("✅ Conexión exitosa a MySQL")

	// Guardamos el *gorm.DB como DBExecutor
	DB = &gormWrapper{db: gormDB}
}
