package services

// Variables que apuntan a las funciones reales. Permiten reemplazarlas en tests.
var (
	CreateTachoFunc  = func(req CreateTachoRequest) (*CreateTachoResponse, error) { return CreateTacho(req) }
	DeleteTachoFunc  = func(customID string) error { return DeleteTacho(customID) }
	GetAllTachosFunc = func() ([]TachoCompleto, error) { return GetAllTachos() }
)
