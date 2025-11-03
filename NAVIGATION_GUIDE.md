# 🧭 Guía de Navegación Punto a Punto

## 📋 Descripción

Sistema de navegación secuencial para conductores de camiones recolectores. Permite avanzar punto por punto a través de la ruta asignada, mostrando siempre la ubicación actual y la siguiente.

---

## 🎯 Endpoints Disponibles

### 1. **Navegación por Barrio y Ruta** (Requiere conocer el barrio y número de ruta)

#### **Iniciar Navegación**
```http
GET /v2/ruta-navegacion/{neighborhood}/{route}/start
```

**Ejemplo:**
```bash
curl http://localhost:8080/v2/ruta-navegacion/8/1/start
```

#### **Navegar a un Punto Específico**
```http
GET /v2/ruta-navegacion/{neighborhood}/{route}?index={index}
```

**Ejemplo:**
```bash
# Punto inicial (depot)
curl http://localhost:8080/v2/ruta-navegacion/8/1?index=0

# Primer tacho
curl http://localhost:8080/v2/ruta-navegacion/8/1?index=1

# Segundo tacho
curl http://localhost:8080/v2/ruta-navegacion/8/1?index=2
```

---

### 2. **Navegación por Email del Conductor** (Más fácil para el front)

#### **Iniciar Navegación**
```http
GET /v2/ruta-navegacion-by-header/start
Header: X-User-Email: conductor.b8.r1@empresa.com
```

**Ejemplo:**
```bash
curl -H "X-User-Email: conductor.b8.r1@empresa.com" \
     http://localhost:8080/v2/ruta-navegacion-by-header/start
```

#### **Navegar a un Punto Específico**
```http
GET /v2/ruta-navegacion-by-header?index={index}
Header: X-User-Email: conductor.b8.r1@empresa.com
```

**Ejemplo:**
```bash
# Punto inicial
curl -H "X-User-Email: conductor.b8.r1@empresa.com" \
     "http://localhost:8080/v2/ruta-navegacion-by-header?index=0"

# Siguiente punto
curl -H "X-User-Email: conductor.b8.r1@empresa.com" \
     "http://localhost:8080/v2/ruta-navegacion-by-header?index=1"
```

---

## 📦 Respuesta del Endpoint

### Estructura de `RouteNavigationResponse`

```json
{
  "route_id": "route_1",
  "current_index": 0,
  "total_points": 15,
  "current_point": {
    "lat": -34.60853946523944,
    "lon": -58.4167656303893
  },
  "next_point": {
    "lat": -34.60901234567890,
    "lon": -58.4171234567890
  },
  "is_first_point": true,
  "is_last_point": false,
  "progress_percentage": 0.0,
  "next_point_key": "/v2/ruta-navegacion-by-header?index=1",
  "previous_point_key": null
}
```

### Campos de la Respuesta

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `route_id` | string | ID de la ruta |
| `current_index` | int | Índice del punto actual (0-based) |
| `total_points` | int | Total de puntos en la ruta |
| `current_point` | object | Coordenadas del punto actual |
| `next_point` | object | Coordenadas del siguiente punto (null si es el último) |
| `is_first_point` | bool | `true` si es el primer punto (depot) |
| `is_last_point` | bool | `true` si es el último punto (depot final) |
| `progress_percentage` | float | Porcentaje de progreso (0-100) |
| `next_point_key` | string | URL para obtener el siguiente punto |
| `previous_point_key` | string | URL para volver al punto anterior |

---

## 🚀 Flujo de Navegación para el Frontend

### Opción 1: Usar `next_point_key` (Recomendado)

```javascript
// 1. Iniciar navegación
const startResponse = await fetch('/v2/ruta-navegacion-by-header/start', {
  headers: { 'X-User-Email': 'conductor.b8.r1@empresa.com' }
});
const navigation = await startResponse.json();

// 2. Mostrar punto actual
console.log('Punto actual:', navigation.current_point);
console.log('Siguiente punto:', navigation.next_point);
console.log('Progreso:', navigation.progress_percentage + '%');

// 3. Botón "Siguiente" - Usar next_point_key
if (!navigation.is_last_point) {
  const nextResponse = await fetch(navigation.next_point_key, {
    headers: { 'X-User-Email': 'conductor.b8.r1@empresa.com' }
  });
  const nextNavigation = await nextResponse.json();
  // Actualizar UI con nuevo punto
}

// 4. Botón "Anterior" - Usar previous_point_key
if (!navigation.is_first_point) {
  const prevResponse = await fetch(navigation.previous_point_key, {
    headers: { 'X-User-Email': 'conductor.b8.r1@empresa.com' }
  });
  const prevNavigation = await prevResponse.json();
  // Actualizar UI con punto anterior
}
```

### Opción 2: Manejar el índice manualmente

```javascript
let currentIndex = 0;
const email = 'conductor.b8.r1@empresa.com';

async function getCurrentPoint() {
  const response = await fetch(
    `/v2/ruta-navegacion-by-header?index=${currentIndex}`,
    { headers: { 'X-User-Email': email } }
  );
  return await response.json();
}

// Siguiente punto
async function nextPoint() {
  currentIndex++;
  return await getCurrentPoint();
}

// Punto anterior
async function previousPoint() {
  currentIndex--;
  return await getCurrentPoint();
}

// Iniciar
async function start() {
  currentIndex = 0;
  return await getCurrentPoint();
}
```

---

## 🎨 Ejemplo de UI Sugerida

```
┌─────────────────────────────────────────┐
│  🚛 Ruta de Recolección                │
│  Progreso: 35.7% (5/14 puntos)         │
├─────────────────────────────────────────┤
│  📍 PUNTO ACTUAL                        │
│  Lat: -34.6085, Lon: -58.4167          │
│  [Depot Inicial]                        │
│                                         │
│  ⬇️                                     │
│                                         │
│  📌 SIGUIENTE PUNTO                     │
│  Lat: -34.6090, Lon: -58.4171          │
│  [Tacho #1]                            │
│                                         │
│  [◀ Anterior]  [Completar ▶]           │
└─────────────────────────────────────────┘
```

---

## 🔄 Ciclo Completo de una Ruta

1. **Inicio** (`index=0`): Depot inicial
   - `is_first_point = true`
   - `next_point` disponible

2. **Puntos intermedios** (`index=1 a N-2`): Tachos
   - `is_first_point = false`
   - `is_last_point = false`
   - `next_point` y `previous_point_key` disponibles

3. **Final** (`index=N-1`): Depot final (regreso)
   - `is_last_point = true`
   - `next_point = null`
   - Solo `previous_point_key` disponible

---

## 📝 Ejemplos de Casos de Uso

### Caso 1: Conductor inicia su jornada

```bash
# Conductor se loguea con su email
curl -H "X-User-Email: conductor.b8.r1@empresa.com" \
     http://localhost:8080/v2/ruta-navegacion-by-header/start

# Respuesta:
{
  "current_index": 0,
  "total_points": 15,
  "current_point": { "lat": -34.608, "lon": -58.416 },
  "next_point": { "lat": -34.609, "lon": -58.417 },
  "is_first_point": true,
  "is_last_point": false,
  "progress_percentage": 0.0,
  "next_point_key": "/v2/ruta-navegacion-by-header?index=1"
}
```

### Caso 2: Conductor avanza al siguiente tacho

```bash
# Usa el next_point_key de la respuesta anterior
curl -H "X-User-Email: conductor.b8.r1@empresa.com" \
     "http://localhost:8080/v2/ruta-navegacion-by-header?index=1"

# Respuesta:
{
  "current_index": 1,
  "total_points": 15,
  "current_point": { "lat": -34.609, "lon": -58.417 },
  "next_point": { "lat": -34.610, "lon": -58.418 },
  "is_first_point": false,
  "is_last_point": false,
  "progress_percentage": 7.14,
  "next_point_key": "/v2/ruta-navegacion-by-header?index=2",
  "previous_point_key": "/v2/ruta-navegacion-by-header?index=0"
}
```

### Caso 3: Conductor termina la ruta

```bash
# Llega al último punto (depot final)
curl -H "X-User-Email: conductor.b8.r1@empresa.com" \
     "http://localhost:8080/v2/ruta-navegacion-by-header?index=14"

# Respuesta:
{
  "current_index": 14,
  "total_points": 15,
  "current_point": { "lat": -34.608, "lon": -58.416 },
  "next_point": null,
  "is_first_point": false,
  "is_last_point": true,
  "progress_percentage": 100.0,
  "next_point_key": null,
  "previous_point_key": "/v2/ruta-navegacion-by-header?index=13"
}
```

---

## ⚠️ Manejo de Errores

### Error 400: Índice fuera de rango
```json
{
  "error": "Índice fuera de rango. Debe estar entre 0 y 14"
}
```

### Error 404: Ruta no encontrada
```json
{
  "error": "Ruta no encontrada: no route found for neighborhood 8, route 99"
}
```

### Error 404: Usuario no encontrado
```json
{
  "error": "Usuario no encontrado: email no encontrado en Redis"
}
```

---

## 🎯 Ventajas de este Sistema

1. **✅ Simple para el frontend**: Solo necesita el email del conductor
2. **✅ Navegación secuencial**: Avanza punto por punto de forma natural
3. **✅ Control de progreso**: Sabes en qué punto estás y cuántos faltan
4. **✅ URLs autogeneradas**: `next_point_key` y `previous_point_key` facilitan la navegación
5. **✅ Flags útiles**: `is_first_point` y `is_last_point` para UI condicional
6. **✅ Sin estado en el cliente**: El índice se pasa en cada request

---

## 🔗 Relación con Otros Endpoints

Este sistema complementa los endpoints existentes:

- **GET `/v2/ruta-optima-by-header`**: Obtiene la ruta completa
- **GET `/v2/ruta-navegacion-by-header`**: Navega punto por punto
- **GET `/personas/emails`**: Lista todos los conductores con sus emails

---

## 🚀 Próximos Pasos (Ideas Futuras)

1. **Estado de completado**: Marcar puntos como "visitados"
2. **Reordenamiento dinámico**: Saltar puntos o cambiar orden
3. **Tiempo estimado**: Calcular tiempo entre puntos
4. **Notificaciones**: Alertar cuando faltan pocos puntos
5. **Historial**: Ver rutas completadas anteriormente

---

¡Sistema de navegación listo para usar! 🎉
