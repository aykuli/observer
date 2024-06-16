# Observer

* microservice Agent sends OS metrics to Server
* mocroservice Server stores metrics according to configuration

## Configuration

### Usage of server

```shell
-a string
    server address to run on (default "localhost:8080")
-d string
    database source name
-f string
    path to save metrics values (default "/tmp/metrics-db.json")
-i int
    metrics store interval in seconds (default 300)
-k string
    secret key to sign response
-r restore metrics from file (default true)
```

### Usage of agent

```shell
-a string
    report interval in second to post metric values on server (default "localhost:8080")
-k string
    secret key to sign request
-l int
    limit sequential requests to server
-p int
    metric values refreshing interval in second (default 2)
-r int
    report interval in second to post metric values on server (default 10)
```

## Build commands

```shell
cd cmd/server
go build -buildvcs=false -o server

cd ../agent
go build -buildvcs=false -o agent
```

## Run

### Using with database storage

```shell
# go to the root of project
cd ~/<project-folder>/observer

docker compose up -d

cd cmd/server

./server -d='postgresql://localhost/postgres?user=postgres&password=postgres' -i=5
cd ../agent

./agent -l=2 -r=3
```
