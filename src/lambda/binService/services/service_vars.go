package services

// Variables que apuntan a las funciones reales. Permiten reemplazarlas en tests.
var (
	CreateTachoFunc  = func(req CreateTachoRequest) (*CreateTachoResponse, error) { return CreateTacho(req) }
	DeleteTachoFunc  = func(mongoID int) error { return DeleteTacho(mongoID) }
	GetAllTachosFunc = func() ([]TachoCompleto, error) { return GetAllTachos() }
)
