## Prerequisites

- [Go v1.18](https://golang.org/dl)

## Set GO env variable

```
# prevent download GO modules from public github.com
go env -w "GOPRIVATE=github.com/teamyapp/*" 
```

## Install CLI

```
go install github.com/teamyapp/cloud/cli@latest
```

## DB

### Generate info for new database

```bash
cli db new -n [dbName]
```

Eg.

```bash
cli db new -n testing
```

### Migrate

Under your project's root directory, run:

```bash
cli db migrate --steps [steps]
```

Eg.

```bash
cli db migrate --steps 1
```

#### Generate migration files

Under your project's root directory, run:

```bash
cli db migrate new
```

### Initialize DB with preset data

Under your project's root directory, run:

```bash
cli db seed
```

### Create DB seed file

Under your project's root directory, run:

```bash
cli db seed new
```
