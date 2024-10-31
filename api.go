package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/cors"
)

func ackHandler(cfg *Config, state *WashmonState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state.Lock()
		if state.CurrentMachineState == Done {
			state.CurrentMachineState = Clear
			state.LastNotificationAt = time.Time{}
		}
		state.Unlock()

		w.WriteHeader(http.StatusNoContent)

		if cfg.StateFile != "" {
			if err := state.WriteFile(cfg.StateFile); err != nil {
				log.Printf("failed to write state file: %v", err)
			}
		}
	}
}

func muteHandler(cfg *Config, state *MuteState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state.Lock()
		state.MuteUntil = time.Now().Add(3 * time.Hour)
		state.Unlock()

		w.WriteHeader(http.StatusNoContent)
	}
}

func ServeAPI(cfg *Config, state *WashmonState, muteState *MuteState) error {
	state.Lock()
	ackPat := fmt.Sprintf("POST /ack/%s", state.NotificationKey)
	mutePat := fmt.Sprintf("POST /mute/%s", state.NotificationKey)
	state.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc(ackPat, ackHandler(cfg, state))
	mux.HandleFunc(mutePat, muteHandler(cfg, muteState))
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.APIPort),
		Handler: cors.Default().Handler(mux),
	}

	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func AckEndpoint(cfg *Config, state *WashmonState) *url.URL {
	state.Lock()
	defer state.Unlock()

	retv, err := url.Parse(fmt.Sprintf("%s/ack/%s", cfg.APIRoot, state.NotificationKey))
	if err != nil {
		panic(err)
	}
	return retv
}

func MuteEndpoint(cfg *Config, state *WashmonState) *url.URL {
	state.Lock()
	defer state.Unlock()

	retv, err := url.Parse(fmt.Sprintf("%s/mute/%s", cfg.APIRoot, state.NotificationKey))
	if err != nil {
		panic(err)
	}
	return retv
}
