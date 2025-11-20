package services

// SwapCoordinates invierte lat/lng para corregir la inconsistencia de la BD
// La BD almacena las coordenadas invertidas, esta función las devuelve correctamente
func SwapCoordinates(lat, lon float64) (float64, float64) {
	// Invertir: lo que está en lat es realmente lon, y viceversa
	return lon, lat // retorna (lat_real, lng_real)
}

// SwapCoordinatesStrings invierte coordenadas en formato string
func SwapCoordinatesStrings(lat, lon string) (string, string) {
	return lon, lat
}
