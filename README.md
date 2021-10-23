# pg-db-operator

This is a proof-of-concept of a simple Kubernetes operator that provisions a user and database in Postgres.

## But why?
Imagine a multi tenancy scenario with a central platform team offering individual PostgreSQL databases to a multitude of consuming teams/services/projects whose application's run in Kubernetes.
The idea would be that a team in need of a database can declare it with a simple manifest like this:
```yaml
apiVersion: experimental.bujarmurati.com/v1
kind: PostgresDatabase
metadata:
  name: pg-db-test
spec:
  name: test_db
  targetSecret:
    name: secret-pg-db-test
    userNamePostfix: "@servername"
    
```
After which an operator will create a new PostgreSQL database, create a new user with a generated password, grant all privileges to the new database to the new user and generate a Kubernetes secret in the same namespace as the ```PostgresDatabase``` resource.

```yaml
apiVersion: v1
data:
  # shown in plaintext for legibility
  PGDATABASE: "test_db"
  PGPASSWORD: "a_randomly_generated_password"
  PGUSER: "test_db@servername"
kind: Secret
metadata:
  name: secret-pg-db-test
type: Opaque
```

This gives consumers a self-service way to get a database while keeping the 

## Installation
TODO: this needs a helm chart!

## Configuration

| Environment variable | Description |
| --- | --- |
| ```PGPASSWORD``` | Required. The password used by the operator to connect to the PostgreSQL server.|
| ```PGUSER``` | Required. The username used by the operator to connect to the PostgreSQL server.|
| ```PGHOST``` | Required. The hostname of the PostgreSQL server.|
| ```PGPORT``` | Required. The port of the PostgreSQL server.|
| ```PG_DB_OPERATOR_PASSWORD_ROTATION_INTERVAL_SECONDS``` | Optional. The interval (in seconds) for password rotations. If unset, no password rotations will be performed. |

In addition to the ones mentioned above, any of the standard [PostgreSQL environment variables](https://www.postgresql.org/docs/current/libpq-envars.html) can be set if needed.

As a prerequisite, the credentials used for the operator should be roughly equivalent to those of the [PostgreSQL server admin account in Azure](https://docs.microsoft.com/en-us/azure/postgresql/howto-create-users#the-server-admin-account).

## Usage

| Field | Example | Description |
| --- | --- | ---
| ```spec.name``` | "mydatabase" | Shared name of the database AND user to create. |
| ```spec.targetSecret.name``` | "secret-mydatabase" | Name of the secret to create. |
| ```spec.targetSecret.userNamePostFix``` | "@myserver" | Optional postfix to append to the PostgreSQL internal username. (To connect as the "mydatabase" user, a consuming app might have to pass "mydatabase@myserver" at connection time)

## Behavior

### Deletion
When a ```PostgresDatabase``` resource is deleted, only the associated secret will be deleted, but none of the database objects. This is a precaution against accidental loss of data.

## Development

### Requirements

- Python 3.9
- poetry
- Docker
- minikube

### Environment setup

Installing Python dependencies

```bash
poetry install
```

Installing the CRD:
```bash
kubectl apply -f manifests/crd.yaml
```

### Running integration tests

```bash
docker-compose up -d -V
poetry run pytest
docker-compose down
```

### Exploratory testing

```bash
kubectl create ns dev
poetry run kopf run pg_db_operator/operator.py
```