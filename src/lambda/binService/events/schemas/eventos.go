package schemas

import (
	"encoding/json"
	"time"
)

// BaseEvent contiene campos comunes a todos los eventos
type BaseEvent struct {
	EventID   string    `json:"event_id"`  // UUID único del evento
	Timestamp time.Time `json:"timestamp"` // Momento del evento
	ModuloID  string    `json:"modulo_id"` // Identificador del módulo que publica
}

// ========== EVENTOS QUE PUBLICA RESIDUOS ==========

// TachoEliminadoEvent se publica cuando un tacho es eliminado del sistema
type TachoEliminadoEvent struct {
	BaseEvent
	TachoID int    `json:"tacho_id"`
	Motivo  string `json:"motivo,omitempty"`
}

// TachoCreadoEvent se publica cuando se crea un nuevo tacho
type TachoCreadoEvent struct {
	BaseEvent
	TachoID   int     `json:"tacho_id"`
	Capacidad float64 `json:"capacidad"`
	Ubicacion string  `json:"ubicacion,omitempty"`
	ZonaID    int     `json:"zona_id,omitempty"`
}

// TachoActualizadoEvent se publica cuando se actualiza un tacho
type TachoActualizadoEvent struct {
	BaseEvent
	TachoID     int     `json:"tacho_id"`
	Capacidad   float64 `json:"capacidad,omitempty"`
	NuevaZonaID int     `json:"nueva_zona_id,omitempty"`
	Detalles    string  `json:"detalles,omitempty"`
}

// TachoLlenoEvent se publica cuando un tacho alcanza su capacidad máxima
type TachoLlenoEvent struct {
	BaseEvent
	TachoID         int       `json:"tacho_id"`
	CapacidadActual float64   `json:"capacidad_actual"`
	UltimaFecha     time.Time `json:"ultima_fecha"`
	ZonaID          int       `json:"zona_id,omitempty"`
}

// AlertaResueltaEvent - Confirmación de que se resolvió una alerta (PUBLICAMOS)
type AlertaResueltaEvent struct {
	EventID           string    `json:"event_id"`
	EmergenciaID      string    `json:"_id"`              // ID de la emergencia original
	Estado            string    `json:"Estado"`           // "Resuelto"
	RutasPerjudicadas int       `json:"ruta_perjudicada"` // Cantidad de rutas afectadas
	Lng               float64   `json:"lng,omitempty"`
	Lat               float64   `json:"lat,omitempty"`
	Timestamp         time.Time `json:"timestamp"`
	ModuloID          string    `json:"modulo_id"`
}

// ========== EVENTOS QUE CONSUME RESIDUOS (de otros módulos) ==========

// ReclamoRecibidoEvent evento recibido del módulo de Reclamos
type ReclamoRecibidoEvent struct {
	BaseEvent
	ReclamoID   int    `json:"reclamo_id"`
	TachoID     int    `json:"tacho_id"`
	TipoReclamo string `json:"tipo_reclamo"` // "mal_estado", "lleno", "roto", "desbordado"
	Descripcion string `json:"descripcion,omitempty"`
	PrioridadID int    `json:"prioridad_id,omitempty"`
	UsuarioID   int    `json:"usuario_id,omitempty"`
}

// RecoleccionCompletadaEvent evento del módulo de Conductores cuando se vacía un tacho
type RecoleccionCompletadaEvent struct {
	BaseEvent
	TachoID      int       `json:"tacho_id"`
	ConductorID  int       `json:"conductor_id,omitempty"`
	CamionID     int       `json:"camion_id,omitempty"`
	FechaVaciado time.Time `json:"fecha_vaciado"`
}

// AlertaVecinalEvent - Evento de emergencias sobre tachos dañados (RECIBIMOS)
type AlertaVecinalEvent struct {
	ID             string `json:"_id"`
	IDUsuario      string `json:"idUsuario"`
	Prioridad      string `json:"prioridad"`
	ScorePrioridad int    `json:"scorePrioridad"`
	Estado         string `json:"estado"` // "Pendiente", "Resuelta", "Cancelada"
	TipoEmergencia string `json:"tipoEmergencia"`
	Origen         string `json:"origen"`
	Ubicacion      struct {
		Lat       float64 `json:"lat"`
		Lon       float64 `json:"lon"`
		Precision int     `json:"precision"`
	} `json:"ubicacion"`
	Adjuntos  []interface{} `json:"adjuntos"`
	Bateria   int           `json:"bateria"`
	Red       string        `json:"red"`
	Timestamp time.Time     `json:"timestamp"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

// RecoleccionReprogramadaEvent - Calle cortada que afecta rutas (RECIBIMOS de Movilidad)
type RecoleccionReprogramadaEvent struct {
	EventVersion string `json:"eventVersion"`
	Timestamp    string `json:"timestamp"`
	Data         struct {
		TripID string `json:"tripId"`
		Origin struct {
			Coordinates struct {
				Lng float64 `json:"lng"`
				Lat float64 `json:"lat"`
			} `json:"coordinates"`
		} `json:"origin"`
		Destination struct {
			Coordinates struct {
				Lng float64 `json:"lng"`
				Lat float64 `json:"lat"`
			} `json:"coordinates"`
		} `json:"destination"`
	} `json:"data"`
}

// ReclamoEstadoEvent - Notificación de cambio de estado de reclamo (RECIBIMOS de Reclamos)
type ReclamoEstadoEvent struct {
	EventID   string    `json:"event_id"`
	ReclamoID int       `json:"reclamo_id"`
	TachoID   int       `json:"tacho_id"`
	Estado    string    `json:"estado"` // "Resuelto", "Rechazado"
	Motivo    string    `json:"motivo"`
	Timestamp time.Time `json:"timestamp"`
}

// RecoleccionReprogramadaPayload es el payload que publicamos cuando se reprograma la recolección
type RecoleccionReprogramadaPayload struct {
	UltimaRecoleccion  time.Time `json:"ultima_recoleccion"`
	ProximaRecoleccion time.Time `json:"proxima_recoleccion"`
}

// RutaNavegacionPayload es el payload que publicamos con el progreso de navegación de una ruta
type RutaNavegacionPayload struct {
	IDRuta               string              `json:"id_ruta"`
	IndicePuntoActual    int                 `json:"indice_punto_actual"`
	TotalPuntos          int                 `json:"total_puntos"`
	PuntoActual          PuntoNavegacion     `json:"punto_actual"`
	PorcentajeProgreso   float64             `json:"porcentaje_progreso"`
	InformacionAdicional []InformacionEvento `json:"informacion_adicional,omitempty"`
}

// PuntoNavegacion representa un punto en la ruta
type PuntoNavegacion struct {
	Latitud  float64 `json:"latitud"`
	Longitud float64 `json:"longitud"`
}

// InformacionEvento contiene el ID de eventos culturales en ese punto
type InformacionEvento struct {
	IDEvento string `json:"id_evento"`
}

// ========== FORMATO ESTÁNDAR DE ENVELOPE ==========

// EventEnvelope es el formato estándar de todos los eventos según especificación
type EventEnvelope struct {
	ID        string          `json:"id"`        // UUID único del evento
	Timestamp time.Time       `json:"timestamp"` // Momento de generación (ISO 8601)
	Source    string          `json:"source"`    // Microservicio que publica (e.g., "movilidad")
	Topic     string          `json:"topic"`     // Debe coincidir con la Routing Key
	Payload   json.RawMessage `json:"payload"`   // Cuerpo del evento (JSON)
}

// ReclamoResiduoPayload es el payload específico para reclamos de residuos creados
type ReclamoResiduoPayload struct {
	IDReclamo      int       `json:"id_reclamo"`
	IDSubcategoria *int      `json:"id_subcategoria"` // Puede ser null
	Titulo         string    `json:"titulo"`
	Descripcion    string    `json:"descripcion"`
	Prioridad      string    `json:"prioridad"`
	Direccion      string    `json:"direccion"`
	Referencia     string    `json:"referencia"`
	Comuna         *int      `json:"comuna"` // Puede ser null
	Lat            string    `json:"lat"`    // Puede ser string vacío ""
	Lng            string    `json:"lng"`    // Puede ser string vacío ""
	Fecha          time.Time `json:"fecha"`
}

// ReclamoEstadoCambiadoPayload es el payload cuando cambia el estado de un reclamo
type ReclamoEstadoCambiadoPayload struct {
	IDReclamo      int       `json:"id_reclamo"`
	Comentario     string    `json:"comentario"`
	Estado         string    `json:"estado"` // "RESUELTO", "RECHAZADO", "ESPERA_INFO"
	FechaRespuesta time.Time `json:"fechaRespuesta"`
}

// EventoCulturaPayload es el payload para eventos culturales (estructura tentativa)
// TODO: Ajustar campos cuando se conozca la estructura exacta del payload
type EventoCulturaPayload struct {
	// Estructura genérica por ahora para capturar todo
	Data map[string]interface{} `json:"data,omitempty"`

	// Campos esperados según ejemplo (se ajustarán)
	Name  string `json:"name,omitempty"`
	Date  string `json:"date,omitempty"`
	Place string `json:"place,omitempty"`
}

// ========== ROUTING KEYS ==========

const (
	// ===== EVENTOS QUE PUBLICA RESIDUOS (residuos.*) =====

	// Tachos
	RoutingKeyTachoCreado      = "residuos.tacho.creado"
	RoutingKeyTachoActualizado = "residuos.tacho.actualizado"
	RoutingKeyTachoEliminado   = "residuos.tacho.eliminado"
	RoutingKeyTachoLleno       = "residuos.tacho.lleno"

	// Alertas Vecinales
	RoutingKeyAlertaResuelta = "residuos.alertavecinal.resuelta"

	// Reclamos - Cambios de estado (respuesta a otros módulos)
	RoutingKeyReclamoResueltoPub   = "residuos.reclamo.resuelto"
	RoutingKeyReclamoRechazadoPub  = "residuos.reclamo.rechazado"
	RoutingKeyReclamoEsperaInfoPub = "residuos.reclamo.espera_info"

	// Recolección - Reprogramación
	RoutingKeyRecoleccionReprogramadaPub = "residuos.recoleccion.reprogramada"

	// Navegación de rutas - Progreso en tiempo real
	RoutingKeyRutaNavegacion = "residuos.camion.posicion"

	// ===== EVENTOS QUE CONSUME RESIDUOS (de otros equipos) =====

	// De Reclamos - Formato antiguo (deprecar si es necesario)
	RoutingKeyReclamoMalEstado  = "reclamos.tacho.mal_estado"
	RoutingKeyReclamoLleno      = "reclamos.tacho.lleno"
	RoutingKeyReclamoRoto       = "reclamos.tacho.roto"
	RoutingKeyReclamoDesbordado = "reclamos.tacho.desbordado"

	// De Reclamos - Formato nuevo (envelope estándar)
	RoutingKeyReclamoResiduoCreado   = "reclamos.residuos.deshecho.creado"
	RoutingKeyReclamoResiduoDerivado = "reclamos.residuos.derivado"

	// De Emergencias
	RoutingKeyAlertaPendiente = "emergencias.alertavecinal.pendiente"

	// De Conductores
	RoutingKeyRecoleccionCompletada = "conductores.recoleccion.completada"

	// De Cultura
	RoutingKeyEventoCulturaCrear    = "cultura.evento.crear"
	RoutingKeyEventoCulturaCancelar = "cultura.evento.cancelado"
)
