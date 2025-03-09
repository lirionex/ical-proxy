# iCal Proxy 📅

The **iCal Proxy** is a lightweight HTTP server written in Go that acts as a caching proxy for iCalendar (iCal) feeds. It forwards requests to upstream iCal URLs based on predefined aliases and caches the responses to reduce load and latency.

## Features 🚀
- **Alias-based URL mapping**: Use simple aliases to access iCal feeds.
- **Configurable Caching**: Customize cache duration via config (default: 30 minutes).
- **Configuration via YAML**: Easy-to-edit configuration file for mappings and cache settings.
- **Graceful error handling**: Provides meaningful HTTP responses and logs errors.
- **Environment variable support**: Customize config and bind address using environment variables.

---

## Installation 🛠️