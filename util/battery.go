//go:build linux

package util

import (
	"errors"

	e "github.com/pkg/errors"
	"github.com/xellio/tools/acpi"
)

type batteryLevel int
type batteryStatus int

const (
	Charging batteryStatus = iota
	NotCharging
	Discharging
)

func makeBatteryStatus(s string) (batteryStatus, error) {
	switch s {
	case "Charging":
		return Charging, nil
	case "Not charging":
		return NotCharging, nil
	case "Discharging":
		return Discharging, nil
	default:
		return 0, errors.New("unknown battery status")
	}
}

type batteryInfo struct {
	Status batteryStatus
	Level  batteryLevel
}

func GetBatteryInfo() (batteryInfo, error) {
	info, err := acpi.Battery()
	if err != nil {
		return batteryInfo{}, e.Wrap(err, "get acpi battery info")
	}
	for _, b := range info {
		if b.Level == 0 {
			continue
		}
		status, err := makeBatteryStatus(b.Status)
		if err != nil {
			return batteryInfo{}, e.Wrap(err, "parse battery status")
		}
		return batteryInfo{
			Status: status,
			Level:  batteryLevel(b.Level),
		}, nil
	}
	return batteryInfo{}, nil
}
