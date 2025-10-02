package middleware

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Business metrics - Tachos
	tachosTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "tachos_total",
			Help: "Total number of tachos in the system",
		},
	)

	tachosCapacidad = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tachos_capacidad_percentage",
			Help: "Current capacity percentage of tachos",
		},
		[]string{"tacho_id", "zona"},
	)

	tachosPrioridad = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tachos_prioridad",
			Help: "Priority level of tachos (1-5, higher is more urgent)",
		},
		[]string{"tacho_id", "zona"},
	)

	// Business metrics - Rutas
	rutasOptimas = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rutas_optimas_calculadas_total",
			Help: "Total number of optimal routes calculated",
		},
		[]string{"zona_id"},
	)

	rutasTiempoCalculo = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rutas_calculo_duration_seconds",
			Help:    "Time taken to calculate optimal routes",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0},
		},
		[]string{"zona_id"},
	)

	// Business metrics - Personas
	personasTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "personas_total",
			Help: "Total number of personas in Redis",
		},
	)

	personasPorZona = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "personas_por_zona",
			Help: "Number of personas per zone",
		},
		[]string{"zona"},
	)

	// Business metrics - Emergencias
	emergenciasEnviadas = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "emergencias_enviadas_total",
			Help: "Total number of emergencies sent",
		},
		[]string{"tipo", "zona"},
	)

	// Database connection metrics
	databaseConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_connections",
			Help: "Current number of database connections",
		},
		[]string{"database_type"}, // mysql, redis, neo4j
	)

	databaseErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"database_type", "operation"},
	)
)

// Business metrics helper functions

// UpdateTachosMetrics updates tacho-related metrics
func UpdateTachosMetrics(total int) {
	tachosTotal.Set(float64(total))
}

// UpdateTachoCapacidad updates the capacity of a specific tacho
func UpdateTachoCapacidad(tachoID, zona string, capacidad float64) {
	tachosCapacidad.WithLabelValues(tachoID, zona).Set(capacidad)
}

// UpdateTachoPrioridad updates the priority of a specific tacho
func UpdateTachoPrioridad(tachoID, zona string, prioridad float64) {
	tachosPrioridad.WithLabelValues(tachoID, zona).Set(prioridad)
}

// IncrementRutasOptimas increments the counter for optimal routes calculated
func IncrementRutasOptimas(zonaID string) {
	rutasOptimas.WithLabelValues(zonaID).Inc()
}

// ObserveRutaCalculoTime observes the time taken to calculate a route
func ObserveRutaCalculoTime(zonaID string, duration float64) {
	rutasTiempoCalculo.WithLabelValues(zonaID).Observe(duration)
}

// UpdatePersonasMetrics updates persona-related metrics
func UpdatePersonasMetrics(total int) {
	personasTotal.Set(float64(total))
}

// UpdatePersonasPorZona updates the number of personas per zone
func UpdatePersonasPorZona(zona string, count int) {
	personasPorZona.WithLabelValues(zona).Set(float64(count))
}

// IncrementEmergencias increments emergency counter
func IncrementEmergencias(tipo, zona string) {
	emergenciasEnviadas.WithLabelValues(tipo, zona).Inc()
}

// UpdateDatabaseConnections updates database connection metrics
func UpdateDatabaseConnections(dbType string, count int) {
	databaseConnections.WithLabelValues(dbType).Set(float64(count))
}

// IncrementDatabaseErrors increments database error counter
func IncrementDatabaseErrors(dbType, operation string) {
	databaseErrors.WithLabelValues(dbType, operation).Inc()
}