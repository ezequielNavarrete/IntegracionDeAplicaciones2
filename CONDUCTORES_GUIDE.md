# 🚀 Sistema de Rutas y Conductores - Guía de Uso

## 📋 Resumen

Este sistema genera automáticamente conductores (personas) para cada ruta calculada y les asigna emails únicos. **Ya no necesitas preocuparte por tener suficientes emails**, el sistema los genera automáticamente.

---

## 🔄 Flujo Completo

### 1️⃣ Ejecutar el Cron Job (Generar Rutas y Conductores)

```bash
curl -X POST \
  -H "Authorization: e3a7e834f4a24bea8392fe75b905a9c193502a6f7a934974a8468ceb8da3babf" \
  http://localhost:8080/cron/update-routes
```

**Esto hace:**
- ✅ Consulta MongoDB para obtener tachos por barrio
- ✅ Distribuye camiones entre barrios
- ✅ Calcula rutas óptimas con el servicio externo
- ✅ Guarda rutas en Redis (`barrio_X_ruta_Y`)
- ✅ **Genera automáticamente un conductor por cada ruta**
- ✅ **Crea emails automáticos** con formato: `conductor.bX.rY@empresa.com`

**Ejemplo de emails generados:**
```
conductor.b8.r1@empresa.com  → Conductor del Barrio 8, Ruta 1
conductor.b8.r2@empresa.com  → Conductor del Barrio 8, Ruta 2
conductor.b12.r1@empresa.com → Conductor del Barrio 12, Ruta 1
conductor.b12.r2@empresa.com → Conductor del Barrio 12, Ruta 2
```

---

## 👥 Ver Todos los Conductores con sus Emails

### Endpoint: `GET /personas/emails`

Este endpoint te muestra **TODOS** los conductores generados con sus emails:

```bash
curl http://localhost:8080/personas/emails
```

**Respuesta:**
```json
{
  "total": 15,
  "message": "Usa el campo 'email' para consultar /v2/ruta-optima-by-header",
  "personas": [
    {
      "id": "1",
      "nombre": "Conductor_B8_R1",
      "email": "conductor.b8.r1@empresa.com",
      "neighborhood_id": 8,
      "route_number": 1,
      "truck_id": 5,
      "route_key": "barrio_8_ruta_1"
    },
    {
      "id": "2",
      "nombre": "Conductor_B8_R2",
      "email": "conductor.b8.r2@empresa.com",
      "neighborhood_id": 8,
      "route_number": 2,
      "truck_id": 7,
      "route_key": "barrio_8_ruta_2"
    }
    // ... más conductores
  ]
}
```

---

## 🎯 Consultar la Ruta de un Conductor

### Opción 1: Por Header (Recomendado para pruebas)

```bash
curl -H "X-User-Email: conductor.b8.r1@empresa.com" \
  http://localhost:8080/v2/ruta-optima-by-header
```

### Opción 2: Con JWT

```bash
curl -H "Authorization: Bearer {tu-jwt-token}" \
  http://localhost:8080/v2/ruta-optima
```

**Respuesta:**
```json
{
  "user": {
    "email": "conductor.b8.r1@empresa.com",
    "persona_id": "1",
    "nombre": "Conductor_B8_R1",
    "neighborhood": 8,
    "route_number": 1
  },
  "route": {
    "route_id": "route_abc123",
    "truck_id": 5,
    "neighborhood": 8,
    "bins_coords": [
      {"lat": -34.603722, "lon": -58.381592},  // Depot inicio
      {"lat": -34.604, "lon": -58.382},        // Tacho 1
      {"lat": -34.605, "lon": -58.383},        // Tacho 2
      // ... más tachos
      {"lat": -34.603722, "lon": -58.381592}   // Depot fin
    ],
    "path_coords": [
      // Camino completo optimizado
    ],
    "total_bins": 45,
    "total_distance_km": 12.5
  }
}
```

---

## 📊 Otros Endpoints Útiles

### Ver todas las personas
```bash
GET /personas
```

### Ver persona por ID
```bash
GET /personas/1
```

### Ver personas de un barrio
```bash
GET /personas/neighborhood/8
```

### Ver todas las rutas de un barrio
```bash
GET /routes/neighborhood/8
```

### Ver ruta específica
```bash
GET /routes/neighborhood/8/route/1
```

---

## 🔑 Formato de Emails

Los emails se generan automáticamente con este patrón:

```
conductor.b{neighborhood_id}.r{route_number}@empresa.com
```

**Ejemplos:**
- Barrio 8, Ruta 1: `conductor.b8.r1@empresa.com`
- Barrio 12, Ruta 3: `conductor.b12.r3@empresa.com`
- Barrio 5, Ruta 10: `conductor.b5.r10@empresa.com`

---

## ✨ Ventajas del Sistema

1. ✅ **No hay límite de conductores**: Se generan automáticamente según las rutas calculadas
2. ✅ **Emails únicos garantizados**: Basados en barrio + ruta
3. ✅ **Fácil de identificar**: El email te dice exactamente qué ruta tiene el conductor
4. ✅ **Regeneración automática**: Cada vez que ejecutas el cron, se regeneran todos los conductores
5. ✅ **Trazabilidad completa**: Sabes exactamente qué email corresponde a qué ruta

---

## 🧪 Flujo de Prueba Completo

### Paso 1: Ejecutar cron para generar rutas y conductores
```bash
curl -X POST \
  -H "Authorization: e3a7e834f4a24bea8392fe75b905a9c193502a6f7a934974a8468ceb8da3babf" \
  http://localhost:8080/cron/update-routes
```

### Paso 2: Ver la lista de conductores con sus emails
```bash
curl http://localhost:8080/personas/emails
```

### Paso 3: Copiar un email de la respuesta y consultar su ruta
```bash
curl -H "X-User-Email: conductor.b8.r1@empresa.com" \
  http://localhost:8080/v2/ruta-optima-by-header
```

---

## 📝 Notas Importantes

- ⚠️ Los **emails dummy originales** (`eze@example.com`, `leo@example.com`, etc.) **se mantienen** en Redis para compatibilidad
- ⚠️ Cada vez que ejecutas el cron, los conductores se **regeneran desde cero**
- ⚠️ Los emails antiguos de conductores se **limpian automáticamente** antes de regenerar
- ✅ El sistema escala automáticamente: **1 ruta = 1 conductor = 1 email único**

---

## 🆘 Troubleshooting

### No veo conductores después del cron
- Verifica que el ROUTE_SERVICE_URL esté configurado correctamente
- Revisa los logs del cron job para ver si hubo errores

### Email no encontrado
- Asegúrate de usar el formato correcto: `conductor.bX.rY@empresa.com`
- Verifica que el cron se haya ejecutado correctamente
- Consulta `/personas/emails` para ver los emails disponibles

### Ruta no encontrada
- El email existe pero la ruta puede haber expirado del cache (TTL de Redis)
- Ejecuta el cron job nuevamente para regenerar las rutas

---

## 🎉 ¡Listo!

Ya no necesitas preocuparte por la cantidad de emails. El sistema se encarga de todo automáticamente. Solo ejecuta el cron y consulta `/personas/emails` para ver todos los conductores disponibles.
