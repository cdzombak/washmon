package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type WashmonMachineState int

const (
	Clear WashmonMachineState = iota
	Running
	Done
)

// allowed transitions:
// Clear -> Running
// Clear -> Done (unusual, but allowed if e.g. the program started right before the machine stopped)
// Running -> Done
// Done -> Running
// Done -> Clear
//
// when in Done, notifications are sent periodically until Clear.

type WashmonState struct {
	sync.Mutex

	NotificationKey     string              `json:"notification_key"`
	CurrentMachineState WashmonMachineState `json:"current_machine_state"`
	LastNotificationAt  time.Time           `json:"last_notification_at"`
}

func (s *WashmonState) WriteFile(filename string) error {
	s.Lock()
	defer s.Unlock()

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed writing state to '%s': %w", filename, err)
	}

	return nil
}

func StateFromFile(filename string) (*WashmonState, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed reading state from '%s': %w", filename, err)
	}

	state := WashmonState{}
	err = json.Unmarshal(data, &state)
	if err != nil {
		return nil, fmt.Errorf("failed parsing state from '%s': %w", filename, err)
	}

	return &state, nil
}

type MuteState struct {
	sync.Mutex

	MuteUntil time.Time
}
