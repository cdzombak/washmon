package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

var version = "<dev>"

func productIdentifier() string {
	return fmt.Sprintf("github.com/cdzombak/washmon@%s", version)
}

func main() {
	configFile := flag.String("config", "./config.json", "Configuration JSON file.")
	printVersion := flag.Bool("version", false, "Print version and exit.")
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if *configFile == "" {
		fmt.Println("-config is required.")
		os.Exit(1)
	}

	config, err := ConfigFromFile(*configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	authString := ""
	if config.InfluxUser != "" || config.InfluxPass != "" {
		authString = fmt.Sprintf("%s:%s", config.InfluxUser, config.InfluxPass)
	} else if config.InfluxToken != "" {
		authString = config.InfluxToken
	}
	influxClient := influxdb2.NewClient(config.InfluxServer, authString)
	if !config.InfluxHealthCheckDisabled {
		ctx, cancel := context.WithTimeout(context.Background(), config.InfluxTimeout())
		defer cancel()
		health, err := influxClient.Health(ctx)
		if err != nil {
			log.Fatalf("Failed to check InfluxDB health: %v", err)
		}
		if health.Status != "pass" {
			log.Fatalf("InfluxDB did not pass health check: status %s; message '%s'", health.Status, *health.Message)
		}
	}
	qAPI := influxClient.QueryAPI(config.InfluxOrg)

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	if err := RunMain(ctx, config, qAPI); err != nil {
		log.Fatal(err)
	}
}

func RunMain(ctx context.Context, cfg *Config, q api.QueryAPI) error {
	state := &WashmonState{}
	if cfg.StateFile != "" {
		s, err := StateFromFile(cfg.StateFile)
		if err != nil {
			log.Printf("failed to load state from '%s': %v", cfg.StateFile, err)
			log.Println("will start with a fresh state")
		} else {
			state = s
		}
	}
	saveState := func() {
		if cfg.StateFile != "" {
			if err := state.WriteFile(cfg.StateFile); err != nil {
				log.Printf("failed to save state to '%s': %v", cfg.StateFile, err)
			}
		}
	}
	if state.NotificationKey == "" {
		state.NotificationKey = RandAlnumString(32)
		saveState()
	}

	muteState := &MuteState{}
	go func() {
		if err := ServeAPI(cfg, state, muteState); err != nil {
			log.Fatalf("API error: %v", err)
		}
	}()

	tickInterval := 1 * time.Minute
	if runFast, err := strconv.ParseBool(os.Getenv("WM_TICK_FAST")); err == nil && runFast {
		tickInterval = 5 * time.Second
	}
	t := time.NewTicker(tickInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			priorPwrWindowMean, err := DoPowerWindowQuery(ctx, q, cfg.PriorWindowPowerMeanQuery)
			if err != nil {
				if IsQueryErrFatal(err) {
					return err
				} else {
					log.Printf("failed to query prior power window: %v", err)
					continue
				}
			}
			currentPwrWindowMean, err := DoPowerWindowQuery(ctx, q, cfg.CurrentWindowPowerMeanQuery)
			if err != nil {
				if IsQueryErrFatal(err) {
					return err
				} else {
					log.Printf("failed to query current power window: %v", err)
					continue
				}
			}

			wasMachineRunning := priorPwrWindowMean > cfg.PowerMeanRunningThreshold
			isMachineRunning := currentPwrWindowMean > cfg.PowerMeanRunningThreshold
			didMachineStop := !isMachineRunning && wasMachineRunning

			state.Lock()

			switch state.CurrentMachineState {
			case Clear:
				if isMachineRunning {
					log.Println("transition from Clear to Running")
					state.CurrentMachineState = Running
				} else if didMachineStop {
					log.Println("transition from Clear to Done")
					state.CurrentMachineState = Done
				}
				state.LastNotificationAt = time.Time{}
			case Running:
				if didMachineStop {
					log.Println("transition from Running to Done")
					state.CurrentMachineState = Done
				}
				state.LastNotificationAt = time.Time{}
			case Done:
				if isMachineRunning {
					log.Println("transition from Done to Running")
					state.CurrentMachineState = Running
				}
			default:
				panic("unhandled CurrentMachineState")
			}

			needsNotify := state.CurrentMachineState == Done &&
				time.Since(state.LastNotificationAt) > cfg.NotifyEvery()

			state.Unlock()
			saveState() // must not be called while holding the lock

			muteState.Lock()
			needsNotify = needsNotify && time.Now().After(muteState.MuteUntil)
			muteState.Unlock()

			if needsNotify {
				go func() {
					if err := SendDoneNotification(ctx, cfg, AckEndpoint(cfg, state), MuteEndpoint(cfg, state)); err != nil {
						log.Printf("failed to send done notification: %v", err)
					} else {
						state.Lock()
						state.LastNotificationAt = time.Now()
						state.Unlock()
						saveState()
					}
				}()
			}
		}
	}
}
