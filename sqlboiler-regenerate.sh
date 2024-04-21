rm -rf ./src/persistence/mysql/model_schema
mkdir -p ./src/persistence/mysql/model_schema -- "$1"

sqlboiler mysql -p model_schema -d --no-tests