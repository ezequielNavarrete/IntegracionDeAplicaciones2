package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/config"
	"go.mongodb.org/mongo-driver/bson"
)

type HorarioRecoleccion struct {
	Inicio time.Time `json:"inicio"`
	Fin    time.Time `json:"fin"`
}

type NeighborhoodSchedule struct {
	Neighborhood              int                `json:"neighborhood"`
	HorarioProximaRecoleccion HorarioRecoleccion `json:"horario_proxima_recoleccion"`
}

type GlobalScheduleConfig struct {
	Horarios                  []NeighborhoodSchedule `json:"horarios"`
	UltimaActualizacionRutas  time.Time              `json:"ultima_actualizacion_rutas"`
	ProximaActualizacionRutas time.Time              `json:"proxima_actualizacion_rutas"`
	CronSchedule              string                 `json:"cron_schedule"`
}

type UpdateScheduleRequest struct {
	CronSchedule string `json:"cron_schedule" binding:"required"`
}

type UpdateNeighborhoodScheduleRequest struct {
	Neighborhood int    `json:"neighborhood" binding:"required"`
	HoraInicio   string `json:"hora_inicio" binding:"required"`
	HoraFin      string `json:"hora_fin" binding:"required"`
	Dia          int    `json:"dia"`  // Opcional: día del mes (1-31)
	Mes          int    `json:"mes"`  // Opcional: mes (1-12)
	Anio         int    `json:"anio"` // Opcional: año (ej: 2025)
}

const scheduleRedisKey = "global_schedule_config"

func GetGlobalSchedule() (*GlobalScheduleConfig, error) {
	if config.RedisClient == nil {
		return nil, fmt.Errorf("redis connection not available")
	}
	ctx := context.Background()
	data, err := config.RedisClient.Get(ctx, scheduleRedisKey).Result()
	if err != nil {
		defaultConfig := createDefaultSchedule()
		if err := saveGlobalSchedule(defaultConfig); err != nil {
			return nil, fmt.Errorf("error saving default config: %v", err)
		}
		return defaultConfig, nil
	}
	var globalConfig GlobalScheduleConfig
	if err := json.Unmarshal([]byte(data), &globalConfig); err != nil {
		return nil, fmt.Errorf("error parsing schedule config: %v", err)
	}
	return &globalConfig, nil
}

func createDefaultSchedule() *GlobalScheduleConfig {
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	horarios := []NeighborhoodSchedule{}

	// Obtener neighborhoods desde MongoDB
	neighborhoods := getActiveNeighborhoodsFromMongo()

	for _, n := range neighborhoods {
		horarios = append(horarios, NeighborhoodSchedule{
			Neighborhood: n,
			HorarioProximaRecoleccion: HorarioRecoleccion{
				Inicio: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 7, 0, 0, 0, now.Location()),
				Fin:    time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 11, 0, 0, 0, now.Location()),
			},
		})
	}
	return &GlobalScheduleConfig{
		Horarios:                  horarios,
		UltimaActualizacionRutas:  now,
		ProximaActualizacionRutas: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 3, 0, 0, 0, now.Location()),
		CronSchedule:              "0 3 * * *",
	}
}

func getActiveNeighborhoodsFromMongo() []int {
	ctx := context.Background()

	// Usar la función de configuración existente
	collection, err := config.GetMongoCollection("bins")
	if err != nil {
		fmt.Printf("Warning: No se pudo obtener colección de MongoDB, usando fallback: %v\n", err)
		return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	}

	// Obtener neighborhoods únicos desde la colección bins
	neighborhoods, err := collection.Distinct(ctx, "neighborhood", bson.M{})
	if err != nil {
		fmt.Printf("Warning: No se pudieron obtener neighborhoods desde MongoDB, usando fallback: %v\n", err)
		return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	}

	// Convertir a []int y ordenar
	result := []int{}
	for _, n := range neighborhoods {
		switch v := n.(type) {
		case int32:
			result = append(result, int(v))
		case int64:
			result = append(result, int(v))
		case int:
			result = append(result, v)
		case float64:
			result = append(result, int(v))
		}
	}

	// Ordenar para mantener consistencia
	sort.Ints(result)

	// Si no encontramos ningún neighborhood, usar fallback
	if len(result) == 0 {
		fmt.Println("Warning: No se encontraron neighborhoods en MongoDB, usando fallback")
		return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	}

	fmt.Printf("✓ Neighborhoods encontrados en MongoDB: %v\n", result)
	return result
}

func UpdateCronSchedule(req UpdateScheduleRequest) (*GlobalScheduleConfig, error) {
	if !isValidCronFormat(req.CronSchedule) {
		return nil, fmt.Errorf("formato cron inválido")
	}
	currentConfig, err := GetGlobalSchedule()
	if err != nil {
		return nil, err
	}
	currentConfig.CronSchedule = req.CronSchedule
	currentConfig.ProximaActualizacionRutas = calculateNextCronExecution(req.CronSchedule)
	if err := saveGlobalSchedule(currentConfig); err != nil {
		return nil, err
	}
	return currentConfig, nil
}

func UpdateNeighborhoodSchedule(req UpdateNeighborhoodScheduleRequest) (*GlobalScheduleConfig, error) {
	currentConfig, err := GetGlobalSchedule()
	if err != nil {
		return nil, err
	}
	horaInicio, err := time.Parse("15:04", req.HoraInicio)
	if err != nil {
		return nil, fmt.Errorf("formato de hora_inicio inválido: %v", err)
	}
	horaFin, err := time.Parse("15:04", req.HoraFin)
	if err != nil {
		return nil, fmt.Errorf("formato de hora_fin inválido: %v", err)
	}

	// Usar fecha especificada o fecha de próxima recolección por defecto
	proximaRecoleccion := currentConfig.ProximaActualizacionRutas
	year, month, day := proximaRecoleccion.Year(), proximaRecoleccion.Month(), proximaRecoleccion.Day()

	// Si se proporcionan campos de fecha, usarlos
	if req.Anio > 0 {
		year = req.Anio
	}
	if req.Mes > 0 && req.Mes <= 12 {
		month = time.Month(req.Mes)
	}
	if req.Dia > 0 && req.Dia <= 31 {
		day = req.Dia
	}

	inicio := time.Date(year, month, day, horaInicio.Hour(), horaInicio.Minute(), 0, 0, proximaRecoleccion.Location())
	fin := time.Date(year, month, day, horaFin.Hour(), horaFin.Minute(), 0, 0, proximaRecoleccion.Location())

	found := false
	for i := range currentConfig.Horarios {
		if currentConfig.Horarios[i].Neighborhood == req.Neighborhood {
			currentConfig.Horarios[i].HorarioProximaRecoleccion = HorarioRecoleccion{Inicio: inicio, Fin: fin}
			found = true
			break
		}
	}
	if !found {
		currentConfig.Horarios = append(currentConfig.Horarios, NeighborhoodSchedule{
			Neighborhood:              req.Neighborhood,
			HorarioProximaRecoleccion: HorarioRecoleccion{Inicio: inicio, Fin: fin},
		})
	}
	return currentConfig, saveGlobalSchedule(currentConfig)
}

func UpdateLastCollectionTime() error {
	currentConfig, err := GetGlobalSchedule()
	if err != nil {
		return err
	}
	now := time.Now()
	currentConfig.UltimaActualizacionRutas = now
	currentConfig.ProximaActualizacionRutas = calculateNextCronExecution(currentConfig.CronSchedule)
	for i := range currentConfig.Horarios {
		horario := &currentConfig.Horarios[i].HorarioProximaRecoleccion
		proximaRecoleccion := currentConfig.ProximaActualizacionRutas
		horario.Inicio = time.Date(proximaRecoleccion.Year(), proximaRecoleccion.Month(), proximaRecoleccion.Day(),
			horario.Inicio.Hour(), horario.Inicio.Minute(), 0, 0, proximaRecoleccion.Location())
		horario.Fin = time.Date(proximaRecoleccion.Year(), proximaRecoleccion.Month(), proximaRecoleccion.Day(),
			horario.Fin.Hour(), horario.Fin.Minute(), 0, 0, proximaRecoleccion.Location())
	}
	return saveGlobalSchedule(currentConfig)
}

func saveGlobalSchedule(schedule *GlobalScheduleConfig) error {
	if config.RedisClient == nil {
		return fmt.Errorf("redis connection not available")
	}
	data, _ := json.Marshal(schedule)
	return config.RedisClient.Set(context.Background(), scheduleRedisKey, data, 0).Err()
}

func isValidCronFormat(cron string) bool {
	return len(splitCronFields(cron)) == 5
}

func splitCronFields(cron string) []string {
	fields, current := []string{}, ""
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

func calculateNextCronExecution(cronSchedule string) time.Time {
	now := time.Now()
	fields := splitCronFields(cronSchedule)
	if len(fields) != 5 {
		return now.Add(24 * time.Hour)
	}
	minute, hour := fields[0], fields[1]
	if minute != "*" && hour != "*" {
		var targetHour, targetMinute int
		fmt.Sscanf(hour, "%d", &targetHour)
		fmt.Sscanf(minute, "%d", &targetMinute)
		next := time.Date(now.Year(), now.Month(), now.Day(), targetHour, targetMinute, 0, 0, now.Location())
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
		return next
	}
	if hour == "*" && minute != "*" {
		var targetMinute int
		fmt.Sscanf(minute, "%d", &targetMinute)
		next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), targetMinute, 0, 0, now.Location())
		if next.Before(now) {
			next = next.Add(1 * time.Hour)
		}
		return next
	}
	if len(hour) > 2 && hour[:2] == "*/" && minute != "*" {
		var interval int
		fmt.Sscanf(hour[2:], "%d", &interval)
		return now.Add(time.Duration(interval) * time.Hour)
	}
	return now.Add(24 * time.Hour)
}
