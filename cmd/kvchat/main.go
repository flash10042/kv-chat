package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/flash10042/kv-chat/internal/persistence"
	"github.com/flash10042/kv-chat/internal/protocol"
	"github.com/flash10042/kv-chat/internal/server"
	"github.com/flash10042/kv-chat/internal/store"
)

var Version = "0.1.0"

type Config struct {
	Address string `json:"address"`
	AOFPath string `json:"aof_path"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config := loadConfig()

	storage := store.NewStorage()

	var aof *persistence.AOF
	if config.AOFPath != "" {
		err := replayAOF(storage, config.AOFPath)
		if err != nil {
			log.Fatalf("Failed to replay AOF: %v", err)
		}

		aof = persistence.NewAOF(config.AOFPath)
		defer aof.Close()
		log.Printf("AOF enabled: %s", config.AOFPath)
	} else {
		log.Printf("AOF disabled")
	}

	startServer(ctx, storage, aof, config.Address)
}

func loadConfig() *Config {
	var (
		addressFlag = flag.String("address", "", "Server address (default: :6379)")
		aofPathFlag = flag.String("aof-path", "", "Path to AOF file (if not provided, AOF is disabled)")
		configFile  = flag.String("config", "", "Path to JSON config file")
	)
	flag.Parse()

	config := &Config{
		Address: ":6379", // default value
		AOFPath: "",
	}

	// Load from config file if provided
	if *configFile != "" {
		fileConfig, err := loadConfigFromFile(*configFile)
		if err != nil {
			log.Fatalf("Failed to load config file: %v", err)
		}
		// Merge config file values
		if fileConfig.Address != "" {
			config.Address = fileConfig.Address
		}
		if fileConfig.AOFPath != "" {
			config.AOFPath = fileConfig.AOFPath
		}
	}

	// CLI flags override config file values
	if *addressFlag != "" {
		config.Address = *addressFlag
	}
	if *aofPathFlag != "" {
		config.AOFPath = *aofPathFlag
	}

	return config
}

func loadConfigFromFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func startServer(ctx context.Context, storage *store.Storage, aof *persistence.AOF, address string) {
	var wg sync.WaitGroup

	listener, err := net.Listen("tcp", address)
	if err != nil {
		// Might be better to return an error and let the caller handle it
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	log.Printf("Listening on %s", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				break

			}
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		wg.Go(func() {
			server.HandleConnection(conn, storage, aof)
		})
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Server shutdown complete")
	case <-time.After(30 * time.Second):
		log.Println("Server shutdown forced")
	}
}

func replayAOF(storage *store.Storage, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		args, err := protocol.ReadCommand(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		protocol.DispatchCommand(protocol.DispatchModePrivate, args, storage, nil)
	}
	return nil
}
