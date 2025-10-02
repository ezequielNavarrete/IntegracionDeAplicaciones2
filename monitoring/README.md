# 🔗 Monitoring Setup

El sistema de monitoreo con Prometheus está ahora en un repositorio separado:

## 📊 **Repositorio de Prometheus:**
**🔗 https://github.com/ezequielNavarrete/prometheus-integracion-app**

## 🎯 **Lo que tienes aquí:**
- ✅ **Dashboards de Grafana** en `monitoring/dashboards/`
- ✅ **Middleware de métricas** en `src/lambda/binService/middleware/metrics.go`
- ✅ **Endpoint /metrics** funcionando en tu aplicación

## 🚀 **Lo que está separado:**
- 🔗 **Prometheus Agent** → Repositorio separado
- 🔗 **Configuración de Grafana Cloud** → Variables de entorno
- 🔗 **Despliegue en Render** → Servicio independiente

## 📋 **Para usar:**
1. **Tu app** expone métricas en `/metrics`
2. **Prometheus** (repo separado) recolecta esas métricas
3. **Grafana Cloud** muestra los dashboards
4. **¡Profit!** 🎉

**¡Todo está configurado y funcionando! Solo falta desplegar Prometheus.**