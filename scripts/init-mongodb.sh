#!/bin/bash

# Script para inicializar datos de ejemplo en MongoDB
# Uso: ./init-mongodb.sh

MONGODB_URI=${MONGODB_URI:-"mongodb://localhost:27017"}
MONGODB_DATABASE=${MONGODB_DATABASE:-"tachos"}

echo "Conectando a MongoDB en: $MONGODB_URI"
echo "Base de datos: $MONGODB_DATABASE"

# Verificar que mongosh esté instalado
if ! command -v mongosh &> /dev/null; then
    echo "Error: mongosh no está instalado"
    echo "Instalar con: brew install mongosh (macOS) o descargar de https://www.mongodb.com/try/download/shell"
    exit 1
fi

# Crear colección bins si no existe
echo "Inicializando colección 'bins'..."

mongosh "$MONGODB_URI/$MONGODB_DATABASE" --quiet --eval '
db.bins.createIndex({ "neighborhood": 1 });
db.bins.createIndex({ "id_mongo": 1 }, { unique: true });
db.bins.createIndex({ "coordinates": "2dsphere" });

// Verificar si ya existen datos
const count = db.bins.countDocuments();
console.log(`Documentos actuales en bins: ${count}`);

if (count === 0) {
    console.log("No hay datos. Puedes importar tus datos de bins aquí.");
    console.log("Ejemplo de estructura de documento:");
    console.log(JSON.stringify({
        id_mongo: 30000,
        id_mysql: 31,
        neighborhood: 8,
        coordinates: {
            lat: -34.617200,
            lon: -58.459626
        },
        priority_calculated: 11,
        total_capacity_L: 100,
        current_demand_L: 100,
        fill_level: "Urgente",
        type: "Basura",
        status: "Bueno"
    }, null, 2));
} else {
    console.log("✓ Colección bins ya tiene datos");
}
'

echo ""
echo "✓ Inicialización completada"
echo ""
echo "Para verificar los datos:"
echo "  mongosh $MONGODB_URI/$MONGODB_DATABASE"
echo "  > db.bins.countDocuments()"
echo "  > db.bins.aggregate([{ \$group: { _id: \"\$neighborhood\", total: { \$sum: 1 } } }])"
