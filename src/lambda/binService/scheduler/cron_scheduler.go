package scheduler

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/robfig/cron/v3"
)

// StartCronScheduler inicia el scheduler de tareas programadas
func StartCronScheduler() {
	// Solo ejecutar en producción si está habilitado
	if os.Getenv("ENABLE_INTERNAL_CRON") != "true" {
		log.Println("⏰ Cron interno deshabilitado (usa ENABLE_INTERNAL_CRON=true para habilitar)")
		return
	}

	c := cron.New()

	// Programar actualización de rutas diaria a las 4 AM
	// Cron format: segundo minuto hora díaMes mes díaSemana
	_, err := c.AddFunc("0 4 * * *", func() {
		log.Println("🕐 [CRON SCHEDULER] Ejecutando actualización automática de rutas...")
		executeRouteUpdate()
	})

	if err != nil {
		log.Printf("❌ Error programando cron: %v", err)
		return
	}

	c.Start()
	log.Println("✅ Cron scheduler iniciado - Actualización diaria a las 4:00 AM")
}

// executeRouteUpdate ejecuta la actualización de rutas
func executeRouteUpdate() {
	apiKey := os.Getenv("CRON_API_KEY")
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	url := baseURL + "/cron/update-routes"

	client := &http.Client{
		Timeout: 10 * time.Minute, // Timeout de 10 minutos para el cron
	}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Printf("❌ Error creando request: %v", err)
		return
	}

	req.Header.Set("Authorization", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Error ejecutando cron: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("✅ Cron ejecutado exitosamente - Status: %d", resp.StatusCode)
	} else {
		log.Printf("⚠️  Cron completado con advertencias - Status: %d", resp.StatusCode)
	}
}
