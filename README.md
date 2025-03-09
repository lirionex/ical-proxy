# iCal Proxy 📅

The **iCal Proxy** is a lightweight HTTP server written in Go that acts as a caching proxy for iCalendar (iCal) feeds.
It forwards requests to upstream iCal URLs based on predefined aliases and caches the responses to reduce load and latency.
It's main purpose is to hide credentials when sharing private calendars. The *arr services for example all include the API key in the calendar url. 

## Features 🚀
- **Alias-based URL mapping**: Use simple aliases to access iCal feeds.
- **Configurable Caching**: Customize cache duration via config (default: 30 minutes).
- **Configuration via YAML**: Easy-to-edit configuration file for mappings and cache settings.
- **Graceful error handling**: Provides meaningful HTTP responses and logs errors.
- **Environment variable support**: Customize config and bind address using environment variables.

---
## Example Config

```yaml
cache_ttl: 24h
mappings:
  sonarr: "https://sonarr.host/feed/v3/calendar/Sonarr.ics?apikey=ABC"
  radarr: "https://radarr.host/feed/v3/calendar/Radarr.ics?apikey=DEF"
```

---

## Installation 🛠️

### From Source
Build the binary
```bash
git clone git@github.com:lirionex/ical-proxy.git

cd ical-proxy

go build -o ical-proxy
```
Run with env vars
```bash
export CONFIG_PATH="~/config.yaml"
./ical-proxy
```

## Docker

```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  --name ical-proxy \
  ghcr.io/lirionex/ical-proxy/ical-proxy:latest
```

## Docker Compose

```yaml
services:
    ical-proxy:
        image: ghcr.io/lirionex/ical-proxy/ical-proxy:latest
        container_name: ical-proxy
        volumes:
          - ./config.yaml:/app/config.yaml
        ports:
          - "8080:8080"
```

## Kubernetes
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ical-proxy-config
data:
  config.yaml: |
    cache_ttl: 24h
    mappings:
      sonarr: "https://sonarr.host/feed/v3/calendar/Sonarr.ics?apikey=ABC"
      radarr: "https://radarr.host/feed/v3/calendar/Radarr.ics?apikey=DEF"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ical-proxy
  labels:
    app: ical-proxy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ical-proxy
  template:
    metadata:
      labels:
        app: ical-proxy
    spec:
      containers:
        - name: ical-proxy
          image: ghcr.io/lirionex/ical-proxy/ical-proxy:latest
          ports:
            - containerPort: 8080
          volumeMounts:
            - name: config-volume
              mountPath: /app/config.yaml
              subPath: config.yaml
      volumes:
        - name: config-volume
          configMap:
            name: ical-proxy-config
---
apiVersion: v1
kind: Service
metadata:
  name: ical-proxy
spec:
  selector:
    app: ical-proxy
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
```