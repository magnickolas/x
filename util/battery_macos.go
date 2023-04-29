//go:build darwin

package util

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"

	e "github.com/pkg/errors"
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
	case "charging":
		return Charging, nil
	case "discharging":
		return Discharging, nil
	case "AC":
		return NotCharging, nil
	default:
		return 0, errors.New("unknown battery status")
	}
}

type batteryInfo struct {
	Status batteryStatus
	Level  batteryLevel
}

func GetBatteryInfo() (batteryInfo, error) {
	cmd := exec.Command("pmset", "-g", "batt")
	out, err := cmd.Output()
	if err != nil {
		return batteryInfo{}, e.Wrap(err, "get battery info")
	}
	line := strings.Split(string(out), "\n")[1]
	fields := strings.FieldsFunc(line, Split)
	level, err := strconv.Atoi(fields[2])
	if err != nil {
		return batteryInfo{}, e.Wrap(err, "cannot parse battery level")
	}
	status, err := makeBatteryStatus(fields[3])
	if err != nil {
		return batteryInfo{}, e.Wrap(err, "unknown battery status")
	}
	return batteryInfo{status, batteryLevel(level)}, nil
}

func Split(r rune) bool {
	return r == ' ' || r == '\t' || r == ';' || r == '%'
}
