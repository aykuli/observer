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

## Profiler

### server app folder

```shell
go tool pprof -http=":9090" -seconds=180 http://localhost:6060/debug/pprof/profile

curl -sK -v http://localhost:6060/debug/pprof/profile > profiles/base.pprof

go tool pprof -http=":9090" -seconds=180 profiles/base.pprof
```

### agent app folder

```shell
go tool pprof -http=":9091" -seconds=180 http://localhost:6061/debug/pprof/profile

curl -sK -v http://localhost:6061/debug/pprof/profile > profiles/base.pprof

go tool pprof -http=":9091" -seconds=180 profiles/base.pprof
```

### server pproff diff

```shell
$ go tool pprof -top -diff_base=profiles/result.pprof profiles/base.pprof
File: server
Type: inuse_space
Time: Jul 27, 2024 at 1:10pm (+07)
Duration: 180.01s, Total samples = 1942.81kB 
Showing nodes accounting for -1040.23kB, 53.54% of 1942.81kB total
      flat  flat%   sum%        cum   cum%
 -528.17kB 27.19% 27.19%  -528.17kB 27.19%  compress/flate.(*dictDecoder).init (inline)
 -512.06kB 26.36% 53.54%  -512.06kB 26.36%  internal/profile.(*Profile).postDecode
         0     0% 53.54%  -528.17kB 27.19%  compress/flate.NewReader
         0     0% 53.54%  -528.17kB 27.19%  compress/gzip.(*Reader).Reset
         0     0% 53.54%  -528.17kB 27.19%  compress/gzip.(*Reader).readHeader
         0     0% 53.54%  -528.17kB 27.19%  compress/gzip.NewReader (inline)
         0     0% 53.54% -1430.75kB 73.64%  github.com/aykuli/observer/cmd/server/routers.MetricsRouter.WithLogging.func2.1
         0     0% 53.54%  -528.17kB 27.19%  github.com/aykuli/observer/internal/compressor.GzipMiddleware.func1
         0     0% 53.54%  -528.17kB 27.19%  github.com/aykuli/observer/internal/compressor.newCompressReader
         0     0% 53.54%  -528.17kB 27.19%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 53.54%   902.59kB 46.46%  github.com/go-chi/chi/v5/middleware.init.0.RequestLogger.func1.1
         0     0% 53.54%  -512.06kB 26.36%  internal/profile.Parse
         0     0% 53.54%  -512.06kB 26.36%  internal/profile.parseUncompressed
         0     0% 53.54%   902.59kB 46.46%  main.main.WithLogging.func4
         0     0% 53.54%  -512.06kB 26.36%  net/http.(*ServeMux).ServeHTTP
         0     0% 53.54% -1040.23kB 53.54%  net/http.(*conn).serve
         0     0% 53.54% -1040.23kB 53.54%  net/http.HandlerFunc.ServeHTTP
         0     0% 53.54% -1040.23kB 53.54%  net/http.serverHandler.ServeHTTP
         0     0% 53.54%  -512.06kB 26.36%  net/http/pprof.Index
         0     0% 53.54%  -512.06kB 26.36%  net/http/pprof.collectProfile
         0     0% 53.54%  -512.06kB 26.36%  net/http/pprof.handler.ServeHTTP
         0     0% 53.54%  -512.06kB 26.36%  net/http/pprof.handler.serveDeltaProfile

```

### agent pproff diff

```shell
g$ go tool pprof -top -diff_base=profiles/result.pprof profiles/base.pprof
File: agent
Type: inuse_space
Time: Jul 27, 2024 at 1:19pm (+07)
Showing nodes accounting for 1944.42kB, 368.15% of 528.17kB total
      flat  flat%   sum%        cum   cum%
  902.59kB 170.89% 170.89%  1447.25kB 274.01%  compress/flate.NewWriter (inline)
  544.67kB 103.12% 274.01%   544.67kB 103.12%  compress/flate.(*compressor).initDeflate (inline)
 -528.17kB   100% 174.01%  -528.17kB   100%  compress/flate.(*dictDecoder).init (inline)
  513.31kB 97.19% 271.20%   513.31kB 97.19%  sync.(*Pool).pinSlow
  512.02kB 96.94% 368.15%   512.02kB 96.94%  net/http.init
         0     0% 368.15%   544.67kB 103.12%  compress/flate.(*compressor).init
         0     0% 368.15%  -528.17kB   100%  compress/flate.NewReader
         0     0% 368.15%  -528.17kB   100%  compress/gzip.(*Reader).Reset
         0     0% 368.15%  -528.17kB   100%  compress/gzip.(*Reader).readHeader
         0     0% 368.15%  1447.25kB 274.01%  compress/gzip.(*Writer).Write
         0     0% 368.15%  -528.17kB   100%  compress/gzip.NewReader (inline)
         0     0% 368.15%   513.31kB 97.19%  fmt.Fprintf
         0     0% 368.15%   513.31kB 97.19%  fmt.newPrinter
         0     0% 368.15%   919.09kB 174.01%  github.com/aykuli/observer/cmd/agent/client.(*MetricsClient).SendBatchMetrics
         0     0% 368.15%  1447.25kB 274.01%  github.com/aykuli/observer/internal/compressor.Compress
         0     0% 368.15%  -528.17kB   100%  github.com/go-resty/resty/v2.(*Client).execute
         0     0% 368.15%  -528.17kB   100%  github.com/go-resty/resty/v2.(*Request).Execute
         0     0% 368.15%  -528.17kB   100%  github.com/go-resty/resty/v2.(*Request).Execute.func2
         0     0% 368.15%  -528.17kB   100%  github.com/go-resty/resty/v2.(*Request).Send (inline)
         0     0% 368.15%  -528.17kB   100%  github.com/go-resty/resty/v2.Backoff
         0     0% 368.15%   919.09kB 174.01%  main.main
         0     0% 368.15%   513.31kB 97.19%  net/http.(*Request).write
         0     0% 368.15%   513.31kB 97.19%  net/http.(*persistConn).writeLoop
         0     0% 368.15%   512.02kB 96.94%  runtime.doInit (inline)
         0     0% 368.15%   512.02kB 96.94%  runtime.doInit1
         0     0% 368.15%  1431.11kB 270.96%  runtime.main
         0     0% 368.15%   513.31kB 97.19%  sync.(*Pool).Get
         0     0% 368.15%   513.31kB 97.19%  sync.(*Pool).pin
         
```

## Test coverage

In root folder:

```shell
go test -cover -v ./...
```

## Documentation

```shell
godoc -http=:8080 -play
```

### swagger generating

```shell
cd cmd/server
swag init --output ./swagger/
```

## Analyse code

```shell
go vet -vettool=/home/a/go/bin/shadow ./...
```

