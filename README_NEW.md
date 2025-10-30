# 🗑️ Sistema de Gestión de Rutas de Recolección - Integración de Aplicaciones 2

Sistema completo para la gestión inteligente de rutas de recolección de residuos, con cálculo automático de rutas óptimas, caché distribuido y actualización programada mediante cron jobs.

## 🏗️ Arquitectura

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Cliente   │────▶│   API REST  │────▶│   MySQL     │
│  (Frontend) │     │   (Gin/Go)  │     │  (Camiones, │
└─────────────┘     └─────────────┘     │   Centros)  │
                           │             └─────────────┘
                           │
                    ┌──────┴──────┐
                    │             │
             ┌──────▼─────┐ ┌────▼──────┐
             │   Redis    │ │  MongoDB  │
             │  (Cache)   │ │  (Tachos) │
             └────────────┘ └───────────┘
                    │
             ┌──────▼─────┐
             │   Neo4j    │
             │  (Grafos)  │
             └────────────┘
```

## 🚀 Características

- ✅ **API RESTful** con Gin framework
- ✅ **Cache distribuido** con Redis para rutas óptimas
- ✅ **Base de datos relacional** (MySQL) para camiones y centros
- ✅ **Base de datos de grafos** (Neo4j) para datos geoespaciales
- ✅ **Base de datos documental** (MongoDB) para tachos de basura
- ✅ **Cron job** para actualización automática de rutas
- ✅ **Distribución inteligente** de camiones según carga de trabajo
- ✅ **Métricas con Prometheus** y visualización con Grafana
- ✅ **Documentación automática** con Swagger/OpenAPI
- ✅ **Docker Compose** para desarrollo local

## 📋 Requisitos

- **Go** 1.24+
- **Docker** y **Docker Compose**
- **Make** (opcional, para comandos simplificados)

### Servicios Externos (Configurables)
- MySQL (incluido en docker-compose)
- Redis (incluido en docker-compose)
- MongoDB (incluido en docker-compose)
- Neo4j (configurar externamente)
- Servicio de cálculo de rutas (API externa)

## 🛠️ Instalación y Configuración

### 1. Clonar el Repositorio

```bash
git clone https://github.com/ezequielNavarrete/IntegracionDeAplicaciones2.git
cd IntegracionDeAplicaciones2
```

### 2. Configurar Variables de Entorno

```bash
cp .env.example .env
```

Editar `.env` con tus configuraciones:

```env
# MongoDB
MONGODB_URI=mongodb://mongodb:27017
MONGODB_DATABASE=tachos

# MySQL
MYSQL_HOST=your_mysql_host
MYSQL_PASSWORD=your_mysql_password

# Neo4j
NEO4J_URI=bolt://your_neo4j_host:7687
NEO4J_PASSWORD=your_neo4j_password

# Redis
REDIS_URL=redis://redis:6379
REDIS_TTL_SECONDS=300

# Servicio de Rutas (API externa)
ROUTE_SERVICE_URL=https://your-route-service.com/calculate-routes

# Cron Job Security
CRON_API_KEY=$(openssl rand -hex 32)
```

### 3. Iniciar Servicios con Docker Compose

```bash
docker-compose up -d
```

Esto iniciará:
- API en `http://localhost:8080`
- Redis en `localhost:6379`
- MongoDB en `localhost:27017`
- Prometheus en `http://localhost:9090`

### 4. Verificar que todo funcione

```bash
# Health check
curl http://localhost:8080/metrics

# Ver documentación API
open http://localhost:8080/swagger/index.html
```

## 📊 Estructura del Proyecto

```
.
├── src/lambda/binService/
│   ├── config/              # Configuración de conexiones
│   │   ├── db.go           # MySQL
│   │   ├── redis.go        # Redis
│   │   ├── mongodb.go      # MongoDB ✨ NEW
│   │   └── neo4j.go        # Neo4j
│   ├── handlers/           # Controladores HTTP
│   │   ├── rutas_handler.go
│   │   ├── cron_handler.go ✨ NEW
│   │   └── routes_cache_handler.go ✨ NEW
│   ├── services/           # Lógica de negocio
│   │   ├── redis_cache.go
│   │   ├── mongodb_service.go ✨ NEW
│   │   └── route_distribution.go ✨ NEW
│   ├── middleware/         # Middlewares
│   │   └── business_metrics.go
│   ├── routes/             # Definición de rutas
│   └── main.go
├── docker-compose.yml      # Configuración Docker ✨ UPDATED
├── Dockerfile
├── CRON_SETUP.md          # Documentación del cron ✨ NEW
└── README.md
```

## 🔄 Sistema de Cron Job

### Flujo de Actualización Automática

1. **Obtener camiones disponibles** desde MySQL
2. **Consultar barrios con tachos** desde MongoDB
3. **Distribuir camiones** inteligentemente:
   - Mínimo 1 camión por barrio
   - Más camiones a barrios con más tachos
4. **Calcular rutas** llamando al servicio externo
5. **Guardar en Redis** con formato optimizado

### Configuración del Cron

Ver documentación completa en [CRON_SETUP.md](./CRON_SETUP.md)

#### Trigger Manual

```bash
curl -X POST \
  -H "Authorization: YOUR_CRON_API_KEY" \
  http://localhost:8080/cron/update-routes
```

#### Configuración en Render

```yaml
# Cron Job Settings
Schedule: 0 */6 * * *  # Cada 6 horas
Command: curl -X POST -H "Authorization: ${CRON_API_KEY}" https://tu-api.onrender.com/cron/update-routes
```

## 📡 Endpoints Principales

### Rutas

- `GET /ruta-optima/:zonaID` - Obtener ruta óptima (con caché)
- `GET /ruta-optima` - Obtener ruta por email (header)
- `GET /routes/neighborhood/:neighborhood` - Todas las rutas de un barrio ✨ NEW
- `GET /routes/neighborhood/:neighborhood/route/:routeNumber` - Ruta específica ✨ NEW

### Cron

- `POST /cron/update-routes` - Actualizar todas las rutas (requiere auth) ✨ NEW

### Camiones

- `GET /camiones` - Listar todos los camiones
- `GET /camiones/:id` - Obtener camión específico

### Centros

- `GET /centros` - Listar todos los centros
- `GET /centros/:id` - Obtener centro específico

### Tachos

- `GET /tachos` - Listar todos los tachos
- `POST /tachos` - Crear nuevo tacho
- `PUT /tachos/:id/capacidad` - Actualizar capacidad
- `PUT /tachos/:id/prioridad` - Actualizar prioridad

### Personas

- `GET /personas` - Listar todas las personas
- `GET /personas/:id` - Obtener persona específica
- `GET /personas/zona/:zona` - Personas por zona

### Monitoreo

- `GET /metrics` - Métricas de Prometheus

### Documentación

- `GET /swagger/*` - Documentación interactiva Swagger

## 🗄️ Formato de Datos en Redis

Las rutas se almacenan con claves: `barrio_{neighborhood}_ruta_{number}`

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

## 📈 Métricas y Monitoreo

### Prometheus

Acceder a: `http://localhost:9090`

Métricas disponibles:
- `rutas_optimas_calculadas_total` - Total de rutas calculadas
- `rutas_calculo_duration_seconds` - Tiempo de cálculo de rutas
- `cron_update_routes_duration_seconds` - Duración del cron job ✨ NEW
- `tachos_total` - Total de tachos
- `database_connections` - Conexiones a bases de datos

### Grafana

Configurar datasource apuntando a Prometheus y crear dashboards.

## 🧪 Testing

### Desarrollo Local

```bash
# Compilar
cd src/lambda/binService
go build -o main

# Ejecutar tests
go test ./...

# Ver cobertura
go test -cover ./...
```

### Probar MongoDB

```bash
# Conectar a MongoDB
mongosh mongodb://localhost:27017/tachos

# Ver estadísticas de tachos por barrio
use tachos;
db.bins.aggregate([
  { $group: { _id: "$neighborhood", total_tachos: { $sum: 1 } } },
  { $sort: { _id: 1 } }
]);
```

## 🔧 Comandos Útiles

```bash
# Ver logs de todos los servicios
docker-compose logs -f

# Ver logs solo de la API
docker-compose logs -f api

# Ver logs de MongoDB
docker-compose logs -f mongodb

# Reiniciar servicios
docker-compose restart

# Detener todo
docker-compose down

# Limpiar volúmenes
docker-compose down -v
```

## 🚀 Despliegue en Producción

### Render

1. Configurar todas las variables de entorno
2. Configurar el Cron Job (ver CRON_SETUP.md)
3. Deploy automático desde GitHub

### Variables Críticas

```env
MONGODB_URI=mongodb+srv://user:pass@cluster.mongodb.net/
ROUTE_SERVICE_URL=https://your-external-service.com
CRON_API_KEY=your_secure_random_key
REDIS_URL=redis://your-redis-host:6379
```

## 🤝 Contribución

1. Fork el proyecto
2. Crear una rama feature (`git checkout -b feature/AmazingFeature`)
3. Commit cambios (`git commit -m 'Add: nueva funcionalidad'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abrir Pull Request

## 📄 Licencia

Este proyecto es parte del curso de Integración de Aplicaciones 2.

## 👥 Equipo

- Ezequiel Navarrete (@ezequielNavarrete)

## 📚 Documentación Adicional

- [Configuración del Cron Job](./CRON_SETUP.md) ✨ NEW
- [Swagger/OpenAPI](http://localhost:8080/swagger/index.html)
- [Prometheus Metrics](http://localhost:8080/metrics)
