# Métricas de Prometheus - IntegracionDeAplicaciones2

Este documento describe las métricas de Prometheus implementadas en la aplicación.

## 🚀 Endpoint de Métricas

Las métricas están disponibles en: **`/metrics`**

Ejemplo: `http://localhost:8080/metrics`

## 📊 Métricas HTTP Automáticas

Estas métricas se capturan automáticamente para todos los endpoints:

### `http_requests_total`
- **Tipo**: Counter
- **Descripción**: Total de requests HTTP
- **Labels**: `method`, `endpoint`, `status`
- **Ejemplo**: `http_requests_total{method="GET",endpoint="/tachos",status="200"} 15`

### `http_request_duration_seconds`
- **Tipo**: Histogram
- **Descripción**: Duración de requests HTTP en segundos
- **Labels**: `method`, `endpoint`, `status`
- **Buckets**: `[0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]`

### `http_requests_in_progress`
- **Tipo**: Gauge
- **Descripción**: Número de requests HTTP en progreso
- **Sin labels**

## 🗂️ Métricas de Negocio

### Tachos

#### `tachos_total`
- **Tipo**: Gauge
- **Descripción**: Total de tachos en el sistema
- **Uso**: `middleware.UpdateTachosMetrics(count)`

#### `tachos_capacidad_percentage`
- **Tipo**: Gauge
- **Descripción**: Porcentaje de capacidad actual de tachos
- **Labels**: `tacho_id`, `zona`
- **Uso**: `middleware.UpdateTachoCapacidad(tachoID, zona, capacidad)`

#### `tachos_prioridad`
- **Tipo**: Gauge
- **Descripción**: Nivel de prioridad de tachos (1-5, mayor = más urgente)
- **Labels**: `tacho_id`, `zona`
- **Uso**: `middleware.UpdateTachoPrioridad(tachoID, zona, prioridad)`

### Rutas

#### `rutas_optimas_calculadas_total`
- **Tipo**: Counter
- **Descripción**: Total de rutas óptimas calculadas
- **Labels**: `zona_id`
- **Uso**: `middleware.IncrementRutasOptimas(zonaID)`

#### `rutas_calculo_duration_seconds`
- **Tipo**: Histogram
- **Descripción**: Tiempo de cálculo de rutas óptimas
- **Labels**: `zona_id`
- **Buckets**: `[0.1, 0.5, 1.0, 2.0, 5.0, 10.0]`
- **Uso**: `middleware.ObserveRutaCalculoTime(zonaID, duration)`

### Personas

#### `personas_total`
- **Tipo**: Gauge
- **Descripción**: Total de personas en Redis
- **Uso**: `middleware.UpdatePersonasMetrics(total)`

#### `personas_por_zona`
- **Tipo**: Gauge
- **Descripción**: Número de personas por zona
- **Labels**: `zona`
- **Uso**: `middleware.UpdatePersonasPorZona(zona, count)`

### Emergencias

#### `emergencias_enviadas_total`
- **Tipo**: Counter
- **Descripción**: Total de emergencias enviadas
- **Labels**: `tipo`, `zona`
- **Uso**: `middleware.IncrementEmergencias(tipo, zona)`

### Base de Datos

#### `database_connections`
- **Tipo**: Gauge
- **Descripción**: Número actual de conexiones a base de datos
- **Labels**: `database_type` (mysql, redis, neo4j)
- **Uso**: `middleware.UpdateDatabaseConnections(dbType, count)`

#### `database_errors_total`
- **Tipo**: Counter
- **Descripción**: Total de errores de base de datos
- **Labels**: `database_type`, `operation`
- **Uso**: `middleware.IncrementDatabaseErrors(dbType, operation)`

## 🛠️ Cómo Usar las Métricas

### En un Handler

```go
import (
    "time"
    "github.com/ezequielNavarrete/IntegracionDeAplicaciones2/src/lambda/binService/middleware"
)

func MiHandler(c *gin.Context) {
    start := time.Now()
    
    // Tu lógica aquí...
    
    // Registrar métricas
    duration := time.Since(start).Seconds()
    middleware.IncrementRutasOptimas("zona1")
    middleware.ObserveRutaCalculoTime("zona1", duration)
}
```

### Actualizar Métricas de Tachos

```go
// Cuando actualizas la capacidad de un tacho
middleware.UpdateTachoCapacidad("tacho_123", "zona_norte", 85.5)

// Cuando actualizas la prioridad
middleware.UpdateTachoPrioridad("tacho_123", "zona_norte", 4.0)

// Actualizar total de tachos
middleware.UpdateTachosMetrics(totalTachos)
```

## 📈 Consultas Útiles en Prometheus

### HTTP
```promql
# Rate de requests por minuto
rate(http_requests_total[5m])

# Percentil 95 de latencia
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Requests por endpoint
sum(rate(http_requests_total[5m])) by (endpoint)
```

### Negocio
```promql
# Tachos con alta capacidad (>80%)
tachos_capacidad_percentage > 80

# Rate de cálculo de rutas por zona
rate(rutas_optimas_calculadas_total[5m])

# Tiempo promedio de cálculo de rutas
rate(rutas_calculo_duration_seconds_sum[5m]) / rate(rutas_calculo_duration_seconds_count[5m])

# Emergencias por tipo
sum(rate(emergencias_enviadas_total[5m])) by (tipo)
```

## 🚀 Despliegue

Estas métricas están listas para ser consumidas por:
- Tu instancia de Prometheus local/Render
- Grafana Cloud
- Cualquier sistema compatible con métricas de Prometheus

Las métricas HTTP se capturan automáticamente. Las métricas de negocio deben ser llamadas manualmente en tus handlers según sea necesario.

---
¡Ya tienes observabilidad completa en tu aplicación! 📊