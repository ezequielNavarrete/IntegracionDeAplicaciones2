# Configuración del Cron Job para Actualización de Rutas

## Variables de Entorno Requeridas

### Bases de Datos
- `MONGODB_URI`: URI de conexión a MongoDB (e.g., `mongodb://localhost:27017`)
- `MONGODB_DATABASE`: Nombre de la base de datos MongoDB (default: `tachos`)
- `MYSQL_HOST`: Host del servidor MySQL
- `MYSQL_PORT`: Puerto de MySQL (default: `3306`)
- `MYSQL_USER`: Usuario de MySQL
- `MYSQL_PASSWORD`: Contraseña de MySQL
- `MYSQL_DATABASE`: Nombre de la base de datos MySQL
- `REDIS_URL`: URL de conexión a Redis (e.g., `redis://localhost:6379`)

### Servicios Externos
- `ROUTE_SERVICE_URL`: URL del servicio externo que calcula las rutas óptimas
- `NEO4J_URI`: URI de Neo4j
- `NEO4J_USERNAME`: Usuario de Neo4j
- `NEO4J_PASSWORD`: Contraseña de Neo4j

### Caché y Seguridad
- `REDIS_TTL_SECONDS`: Tiempo de vida del caché en segundos (default: `300` = 5 minutos)
- `CRON_API_KEY`: Clave secreta para autenticar las solicitudes del cron job

## Configuración en Render

### 1. Crear Cron Job

1. Ve a tu dashboard en Render
2. Haz clic en "New +" → "Cron Job"
3. Configura:
   - **Name**: `update-routes-cron`
   - **Schedule**: Formato cron (ejemplos abajo)
   - **Command**:
     ```bash
     curl -X POST \
       -H "Authorization: ${CRON_API_KEY}" \
       -H "Content-Type: application/json" \
       https://tu-api.onrender.com/cron/update-routes
     ```

### 2. Ejemplos de Schedule

- **Cada hora**: `0 * * * *`
- **Cada 6 horas**: `0 */6 * * *`
- **Todos los días a las 4 AM**: `0 4 * * *`
- **Cada 30 minutos**: `*/30 * * * *`
- **De lunes a viernes a las 8 AM**: `0 8 * * 1-5`

### 3. Configurar Variables de Entorno

En tu servicio web de Render, añade todas las variables de entorno listadas arriba.

Para generar una clave segura para `CRON_API_KEY`:
```bash
openssl rand -hex 32
```

## Endpoints Disponibles

### Cron Job
- `POST /cron/update-routes` - Actualiza todas las rutas (requiere Authorization header)

### Consulta de Rutas
- `GET /routes/neighborhood/:neighborhood` - Obtiene todas las rutas de un barrio
- `GET /routes/neighborhood/:neighborhood/route/:routeNumber` - Obtiene una ruta específica

### Ejemplos de Uso

#### Trigger manual del cron
```bash
curl -X POST \
  -H "Authorization: tu-clave-secreta" \
  https://tu-api.onrender.com/cron/update-routes
```

#### Obtener todas las rutas del barrio 8
```bash
curl https://tu-api.onrender.com/routes/neighborhood/8
```

#### Obtener ruta específica
```bash
curl https://tu-api.onrender.com/routes/neighborhood/8/route/1
```

## Formato de Datos en Redis

Las rutas se guardan con claves tipo: `barrio_8_ruta_1`

Cada ruta contiene:
```json
{
  "route_id": "route_1",
  "truck_id": 1,
  "neighborhood": 8,
  "bins_coords": [
    {"lat": -34.617, "lon": -58.459},
    {"lat": -34.638, "lon": -58.445}
  ],
  "path_coords": [
    {"lat": -34.617, "lon": -58.457},
    {"lat": -34.616, "lon": -58.458}
  ],
  "total_bins": 6,
  "total_distance_km": 11.13
}
```

## Flujo del Cron Job

1. **Obtener camiones disponibles** desde MySQL (filtrados por estado operativo)
2. **Consultar barrios con tachos** desde MongoDB
3. **Distribuir camiones** entre barrios (mínimo 1 por barrio, más según cantidad de tachos)
4. **Calcular rutas** para cada barrio llamando al servicio externo
5. **Guardar rutas simplificadas** en Redis con TTL configurado
6. **Retornar resumen** con estadísticas de éxito/errores

## Monitoreo

El cron job expone métricas en Prometheus:
- `cron_update_routes_duration_seconds` - Duración del proceso completo
- `rutas_optimas_calculadas_total` - Total de rutas calculadas por zona

Accede a las métricas en: `https://tu-api.onrender.com/metrics`

## Troubleshooting

### Error: "No hay camiones disponibles"
- Verifica que existan camiones con estado "Operativo" o "Disponible" en MySQL

### Error: "No se encontraron barrios con tachos"
- Verifica la conexión a MongoDB
- Asegúrate de que la colección `bins` tenga datos

### Error: "Error calling route service"
- Verifica que `ROUTE_SERVICE_URL` esté correctamente configurada
- Verifica que el servicio externo esté disponible

### Error de autenticación en el cron
- Verifica que el header `Authorization` contenga exactamente el valor de `CRON_API_KEY`
