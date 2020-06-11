package main

const envTemplate = `# Basic env configuration

# Service Handlers
USE_HTTP=[bool]
USE_GRPC=[bool]
USE_GRAPHQL=[bool]
USE_KAFKA=[bool]

HTTP_PORT=[int]
GRPC_PORT=[int]

BASIC_AUTH_USERNAME=[string]
BASIC_AUTH_PASS=[string]

USE_MONGO=[bool]
MONGODB_HOST_WRITE=[string]
MONGODB_HOST_READ=[string]
MONGODB_DATABASE_NAME=[string]

USE_SQL=[bool]
SQL_DRIVER_NAME=[string]
SQL_DB_READ_HOST=[string]
SQL_DB_READ_USER=[string]
SQL_DB_READ_PASSWORD=[string]
SQL_DB_WRITE_HOST=[string]
SQL_DB_WRITE_USER=[string]
SQL_DB_WRITE_PASSWORD=[string]
SQL_DATABASE_NAME=[string]

USE_REDIS=[bool]
REDIS_READ_HOST=[string]
REDIS_READ_PORT=[string]
REDIS_READ_AUTH=[string]
REDIS_WRITE_HOST=[string]
REDIS_WRITE_PORT=[string]
REDIS_WRITE_AUTH=[string]

USE_RSA_KEY=[bool]

KAFKA_BROKERS=[string]

`