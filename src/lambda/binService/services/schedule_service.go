package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
)

// ScheduleConfig representa la configuración de horarios de recolección
type ScheduleConfig struct {
	CronSchedule     string    `json:"cron_schedule"`      // Formato cron: "0 16 * * *"
	NextCollection   time.Time `json:"next_collection"`    // Próxima recolección planificada
	LastUpdated      time.Time `json:"last_updated"`       // Última actualización
	Description      string    `json:"description"`        // Descripción legible del cron
}

// UpdateScheduleRequest representa los datos para actualizar el horario
type UpdateScheduleRequest struct {
	CronSchedule string `json:"cron_schedule" binding:"required"` // Ej: "0 16 * * *" = 4PM diario
	Description  string `json:"description"`                      // Opcional: descripción personalizada (si no se provee, se genera automáticamente)
}

const scheduleRedisKey = "schedule_config"

// GetScheduleConfig obtiene la configuración actual de horarios desde Redis
func GetScheduleConfig() (*ScheduleConfig, error) {
	if config.RedisClient == nil {
		return nil, fmt.Errorf("redis connection not available")
	}

	ctx := context.Background()
	data, err := config.RedisClient.Get(ctx, scheduleRedisKey).Result()
	
	// Si no existe, devolver configuración por defecto
	if err != nil {
		defaultConfig := &ScheduleConfig{
			CronSchedule:   "0 16 * * *",
			NextCollection: calculateNextCronExecution("0 16 * * *"),
			LastUpdated:    time.Now(),
			Description:    "Diario a las 16:00 (4 PM)",
		}
		
		// Guardar la configuración por defecto
		if err := saveScheduleConfig(defaultConfig); err != nil {
			return nil, fmt.Errorf("error saving default config: %v", err)
		}
		
		return defaultConfig, nil
	}

	var config ScheduleConfig
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, fmt.Errorf("error parsing schedule config: %v", err)
	}

	return &config, nil
}

// UpdateScheduleConfig actualiza la configuración de horarios
func UpdateScheduleConfig(req UpdateScheduleRequest) (*ScheduleConfig, error) {
	// Validar formato cron (básico)
	if !isValidCronFormat(req.CronSchedule) {
		return nil, fmt.Errorf("formato cron inválido. Ejemplo válido: '0 16 * * *'")
	}

	// Si no se provee descripción personalizada, generarla automáticamente
	description := req.Description
	if description == "" {
		description = describeCronSchedule(req.CronSchedule)
	}

	newConfig := &ScheduleConfig{
		CronSchedule:   req.CronSchedule,
		NextCollection: calculateNextCronExecution(req.CronSchedule),
		LastUpdated:    time.Now(),
		Description:    description,
	}

	if err := saveScheduleConfig(newConfig); err != nil {
		return nil, fmt.Errorf("error saving schedule config: %v", err)
	}

	return newConfig, nil
}

// UpdateLastCollectionTime actualiza la última vez que se ejecutó la recolección
// Esta función se debe llamar desde el cron handler cuando se ejecuta exitosamente
func UpdateLastCollectionTime() error {
	currentConfig, err := GetScheduleConfig()
	if err != nil {
		return err
	}

	currentConfig.LastUpdated = time.Now()
	currentConfig.NextCollection = calculateNextCronExecution(currentConfig.CronSchedule)

	return saveScheduleConfig(currentConfig)
}

// saveScheduleConfig guarda la configuración en Redis
func saveScheduleConfig(scheduleConfig *ScheduleConfig) error {
	if config.RedisClient == nil {
		return fmt.Errorf("redis connection not available")
	}

	data, err := json.Marshal(scheduleConfig)
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	ctx := context.Background()
	if err := config.RedisClient.Set(ctx, scheduleRedisKey, data, 0).Err(); err != nil {
		return fmt.Errorf("error saving to redis: %v", err)
	}

	return nil
}

// isValidCronFormat valida básicamente el formato cron (5 campos)
func isValidCronFormat(cron string) bool {
	// Validación simple: debe tener 5 campos separados por espacios
	// Formato: "minuto hora día mes día_semana"
	// Ejemplo: "0 16 * * *" = 4 PM todos los días
	fields := len(splitCronFields(cron))
	return fields == 5
}

// splitCronFields divide el cron en campos
func splitCronFields(cron string) []string {
	fields := []string{}
	current := ""
	
	for _, char := range cron {
		if char == ' ' {
			if current != "" {
				fields = append(fields, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		fields = append(fields, current)
	}
	
	return fields
}

// calculateNextCronExecution calcula la próxima ejecución basada en el cron
// Versión simplificada - solo maneja algunos casos comunes
func calculateNextCronExecution(cronSchedule string) time.Time {
	now := time.Now()
	fields := splitCronFields(cronSchedule)
	
	if len(fields) != 5 {
		// Si no es válido, asumir 24 horas desde ahora
		return now.Add(24 * time.Hour)
	}

	minute := fields[0]
	hour := fields[1]
	
	// Caso simple: horario fijo diario (ej: "0 16 * * *")
	if minute != "*" && hour != "*" {
		var targetHour, targetMinute int
		fmt.Sscanf(hour, "%d", &targetHour)
		fmt.Sscanf(minute, "%d", &targetMinute)
		
		next := time.Date(now.Year(), now.Month(), now.Day(), targetHour, targetMinute, 0, 0, now.Location())
		
		// Si ya pasó hoy, programar para mañana
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
		
		return next
	}

	// Caso: cada hora (ej: "0 * * * *")
	if hour == "*" && minute != "*" {
		var targetMinute int
		fmt.Sscanf(minute, "%d", &targetMinute)
		
		next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), targetMinute, 0, 0, now.Location())
		
		if next.Before(now) {
			next = next.Add(1 * time.Hour)
		}
		
		return next
	}

	// Caso: cada N horas (ej: "0 */6 * * *")
	if len(hour) > 2 && hour[:2] == "*/" && minute != "*" {
		var interval int
		fmt.Sscanf(hour[2:], "%d", &interval)
		return now.Add(time.Duration(interval) * time.Hour)
	}

	// Default: 24 horas
	return now.Add(24 * time.Hour)
}

// describeCronSchedule genera una descripción legible del cron
func describeCronSchedule(cronSchedule string) string {
	fields := splitCronFields(cronSchedule)
	
	if len(fields) != 5 {
		return "Formato inválido"
	}

	minute := fields[0]
	hour := fields[1]

	// Casos comunes
	if minute == "0" && hour == "16" {
		return "Diario a las 16:00 (4 PM)"
	}
	
	if minute == "0" && hour == "*" {
		return "Cada hora en punto"
	}
	
	if minute == "*/30" && hour == "*" {
		return "Cada 30 minutos"
	}
	
	if minute == "0" && len(hour) > 2 && hour[:2] == "*/" {
		var interval int
		fmt.Sscanf(hour[2:], "%d", &interval)
		return fmt.Sprintf("Cada %d horas", interval)
	}

	if minute != "*" && hour != "*" {
		return fmt.Sprintf("Diario a las %s:%s", hour, minute)
	}

	return "Horario personalizado: " + cronSchedule
}
