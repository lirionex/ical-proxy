package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Config structure for mapping aliases to upstream URLs
type Config struct {
	Mappings map[string]string `yaml:"mappings"`
	CacheTTL string            `yaml:"cache_ttl"`
}

type CacheEntry struct {
	Data      []byte
	Timestamp time.Time
}

type Cache struct {
	entries map[string]CacheEntry
	mutex   sync.Mutex
}

var (
	config   Config
	cacheTTL = 30 * time.Minute
	cache    = Cache{entries: make(map[string]CacheEntry)}
)

func (c *Cache) Get(alias string) ([]byte, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, found := c.entries[alias]
	if found && time.Since(entry.Timestamp) < cacheTTL {
		return entry.Data, true
	}
	return nil, false
}

func (c *Cache) Set(alias string, data []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.entries[alias] = CacheEntry{Data: data, Timestamp: time.Now()}
}

func loadConfig() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/app/config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}
	if len(config.Mappings) == 0 {
		log.Fatalf("Config file contains no mappings")
	}

	// Parse cache TTL if specified
	if config.CacheTTL != "" {
		parsedTTL, err := time.ParseDuration(config.CacheTTL)
		if err != nil {
			log.Fatalf("Invalid cache_ttl format in config: %v", err)
		}
		cacheTTL = parsedTTL
	}
	log.Printf("Cache TTL set to %v", cacheTTL)
}

func fetchCalendar(url string) ([]byte, error) {
	if url == "" {
		return nil, errors.New("empty URL")
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch calendar: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Failed to close response body: %v", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching calendar: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	alias := r.URL.Path[1:]
	log.Printf("Received request from %s: %s %s", clientIP, r.Method, r.URL.Path)

	if alias == "" {
		http.Error(w, "Missing alias", http.StatusBadRequest)
		log.Printf("Bad request from %s: Missing alias", clientIP)
		return
	}

	url, exists := config.Mappings[alias]
	if !exists {
		http.Error(w, "Not found", http.StatusNotFound)
		log.Printf("Not found: %s requested %s", clientIP, alias)
		return
	}

	if data, found := cache.Get(alias); found {
		log.Printf("Cache hit for alias %s", alias)
		w.Header().Set("Content-Type", "text/calendar")
		if _, err := w.Write(data); err != nil {
			log.Printf("Error writing cached response to %s: %v", clientIP, err)
		}
		return
	}

	log.Printf("Cache miss for alias %s, fetching from upstream %s", alias, url)
	data, err := fetchCalendar(url)
	if err != nil {
		log.Printf("Error fetching calendar for alias %s: %v", alias, err)
		http.Error(w, "Failed to fetch calendar", http.StatusInternalServerError)
		return
	}

	cache.Set(alias, data)

	w.Header().Set("Content-Type", "text/calendar")
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing response to %s: %v", clientIP, err)
	}
	log.Printf("Served calendar for alias %s to %s", alias, clientIP)
}

func main() {
	loadConfig()

	bindAddress := os.Getenv("BIND_ADDRESS")
	if bindAddress == "" {
		bindAddress = ":8080"
	}

	http.HandleFunc("/", handler)
	server := &http.Server{
		Addr:         bindAddress,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting iCal Proxy on %s", bindAddress)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
