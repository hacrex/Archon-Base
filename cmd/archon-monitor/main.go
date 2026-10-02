// Command archon-monitor reports local resources for the enthusiast/K3s profile.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/hacrex/Archon-Base/internal/monitor"
)

var version = "0.0.1-dev"

type daemon struct {
	collector monitor.Collector
	latest    monitor.Snapshot
	dataPath  string
}

func main() {
	listen := flag.String("listen", "127.0.0.1:9105", "HTTP listen address; loopback is the safe default")
	interval := flag.Duration("interval", 15*time.Second, "resource collection interval")
	diskPath := flag.String("disk-path", "/var/lib/rancher/k3s", "filesystem path used for capacity checks")
	dataDir := flag.String("data-dir", "/var/lib/archon-monitor", "directory for the latest JSON snapshot")
	once := flag.Bool("once", false, "collect one snapshot, print JSON, and exit")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("archon-monitor", version)
		return
	}
	if *interval <= 0 {
		log.Fatal("interval must be positive")
	}

	collector := monitor.NewCollector(*diskPath)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *once {
		snapshot, err := collector.Collect(ctx)
		if err != nil {
			log.Fatal(err)
		}
		if err := json.NewEncoder(os.Stdout).Encode(snapshot); err != nil {
			log.Fatal(err)
		}
		if !snapshot.Compliance.Ready {
			os.Exit(2)
		}
		return
	}

	if err := os.MkdirAll(*dataDir, 0o750); err != nil {
		log.Fatalf("create data directory: %v", err)
	}
	d := &daemon{collector: collector, dataPath: filepath.Join(*dataDir, "status.json")}
	if err := d.collect(ctx); err != nil {
		log.Printf("initial collection: %v", err)
	}
	go d.run(ctx, *interval)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", d.health)
	mux.HandleFunc("/status", d.status)
	mux.HandleFunc("/metrics", d.metrics)
	server := &http.Server{Addr: *listen, Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	log.Printf("archon-monitor listening on %s", *listen)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func (d *daemon) run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := d.collect(ctx); err != nil {
				log.Printf("resource collection: %v", err)
			}
		}
	}
}

func (d *daemon) collect(ctx context.Context) error {
	snapshot, err := d.collector.Collect(ctx)
	if err != nil {
		return err
	}
	d.latest = snapshot
	return monitor.WriteJSON(d.dataPath, snapshot)
}

func (d *daemon) health(w http.ResponseWriter, _ *http.Request) {
	if d.latest.CollectedAt.IsZero() {
		http.Error(w, "no snapshot", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "profile": d.latest.Profile, "collectedAt": d.latest.CollectedAt})
}

func (d *daemon) status(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(d.latest)
}

func (d *daemon) metrics(w http.ResponseWriter, _ *http.Request) {
	s := d.latest
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "archon_profile_ready %s\n", boolValue(s.Compliance.Ready))
	fmt.Fprintf(w, "archon_cpu_cores %d\n", s.CPU.Cores)
	fmt.Fprintf(w, "archon_cpu_load_1m %f\n", s.CPU.Load1m)
	fmt.Fprintf(w, "archon_cpu_load_ratio %f\n", s.CPU.LoadRatio)
	fmt.Fprintf(w, "archon_memory_total_bytes %d\n", s.Memory.TotalBytes)
	fmt.Fprintf(w, "archon_memory_available_bytes %d\n", s.Memory.AvailableBytes)
	fmt.Fprintf(w, "archon_memory_used_ratio %f\n", s.Memory.UsedRatio)
	fmt.Fprintf(w, "archon_disk_available_bytes{path=%q} %d\n", s.Disk.Path, s.Disk.AvailableBytes)
	fmt.Fprintf(w, "archon_disk_used_ratio{path=%q} %f\n", s.Disk.Path, s.Disk.UsedRatio)
	if s.GPU != nil {
		fmt.Fprintf(w, "archon_gpu_memory_total_bytes %d\n", s.GPU.TotalMemoryMB*1024*1024)
		fmt.Fprintf(w, "archon_gpu_memory_used_bytes %d\n", s.GPU.UsedMemoryMB*1024*1024)
		fmt.Fprintf(w, "archon_gpu_utilization_percent %f\n", s.GPU.Utilization)
	}
}

func boolValue(value bool) string {
	if value {
		return strconv.Itoa(1)
	}
	return strconv.Itoa(0)
}
