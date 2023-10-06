## HTTP Sleeper

Takes a value of seconds or milliseconds and sleeps before generating an HTTP response.

## Examples

```bash
curl sleep.jarv.org/1      # sleep for 1 second
curl sleep.jarv.org/10s    # sleep for 10 seconds
curl sleep.jarv.org/100ms  # sleep for 100ms
```

## Local development

```
go run cmd/sleeper.go
```

## Docker

```
docker-compose build
docker-compose up
```
