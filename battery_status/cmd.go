package battery_status

import (
	_ "embed"
	"fmt"
	"sort"
	"time"

	"github.com/magnickolas/x/util"
	e "github.com/pkg/errors"
	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/conf"
	"github.com/rwxrob/help"
	"github.com/rwxrob/vars"
)

var defs = map[string]string{
	"charging":              `{"20": " ", "40": " ", "60": " ", "80": " ", "100": " "}`,
	"discharging":           `{"20": " ", "40": " ", "60": " ", "80": " ", "100": " "}`,
	"notCharging":           `{"100": ""}`,
	"format":                "{status} {level}%",
	"chargingFrameDuration": "1s",
}
var defKeys = util.Keys(defs)

func init() {
	util.Must(Z.Conf.SoftInit())
	util.Must(Z.Vars.SoftInit())
}

func outputBatteryStatus(c cfg) error {
	info, err := util.GetBatteryInfo()
	if err != nil {
		e.Wrap(err, "get battery info")
	}
	var levelMap map[int]string
	if info.Status == util.Discharging {
		levelMap = c.discharging
	} else if info.Status == util.Charging {
		levelMap = c.charging
	} else if info.Status == util.NotCharging {
		levelMap = c.notCharging
	} else {
		return e.New("unknown status")
	}
	thresholds := util.Keys(levelMap)
	sort.Ints(thresholds)
	var thresholdIndex int
	for i, threshold := range thresholds {
		if int(info.Level) <= threshold {
			thresholdIndex = i
			break
		}
	}
	var statusSymbol string
	frameDuration := c.chargingFrameDuration.Milliseconds()
	if info.Status == util.Charging && frameDuration > 0 {
		if len(thresholds) < 2 {
			return e.Errorf("there should be at least two statuses for animation")
		}
		thresholdIndex = util.Min(thresholdIndex, len(thresholds)-2)
		if int(float64(time.Now().UnixMilli()/frameDuration)+0.5)%2 == 0 {
			statusSymbol = levelMap[thresholds[thresholdIndex]]
		} else {
			statusSymbol = levelMap[thresholds[thresholdIndex+1]]
		}
	} else {
		statusSymbol = levelMap[thresholds[thresholdIndex]]
	}
	args := map[string]interface{}{
		"status": statusSymbol,
		"level":  info.Level,
	}
	_, err = fmt.Println(util.Fprint(c.format, args))
	if err != nil {
		return e.Wrap(err, "print battery status")
	}
	return nil
}

type cfg struct {
	charging              map[int]string
	discharging           map[int]string
	notCharging           map[int]string
	format                string
	chargingFrameDuration time.Duration
}

func getConfig(x *Z.Cmd) (cfg, error) {
	charging, err := util.Get[map[int]string](x, "charging")
	if err != nil {
		return cfg{}, err
	}
	discharging, err := util.Get[map[int]string](x, "discharging")
	if err != nil {
		return cfg{}, err
	}
	notCharging, err := util.Get[map[int]string](x, "notCharging")
	if err != nil {
		return cfg{}, err
	}
	format, err := util.Get[string](x, "format")
	if err != nil {
		return cfg{}, err
	}
	chargingFrameDuration, err := util.Get[time.Duration](x, "chargingFrameDuration")
	if err != nil {
		return cfg{}, err
	}
	return cfg{
		charging:              charging,
		discharging:           discharging,
		notCharging:           notCharging,
		format:                format,
		chargingFrameDuration: chargingFrameDuration,
	}, nil
}

func cmd(x *Z.Cmd) error {
	c, err := getConfig(x)
	if err != nil {
		return e.Wrap(err, "get config")
	}
	return outputBatteryStatus(c)
}

var Cmd = &Z.Cmd{
	Name:    `battery-status`,
	Summary: `output configured symbol if battery is charging`,
	Commands: []*Z.Cmd{
		help.Cmd, vars.Cmd, conf.Cmd,
		initCmd,
	},
	Call: func(x *Z.Cmd, args ...string) error {
		defer util.TrapPanic()
		util.Must(cmd(x))
		return nil
	},
	Shortcuts: util.ShortcutsFromDefs(defKeys),
}

var initCmd = &Z.Cmd{
	Name:     `init`,
	Summary:  `sets all values to defaults`,
	Commands: []*Z.Cmd{help.Cmd},

	Call: func(x *Z.Cmd, _ ...string) error {
		for k, dv := range defs {
			v, _ := x.Caller.C(k)
			if v == "null" {
				v = dv
			}
			x.Caller.Set(k, v)
		}
		return nil
	},
}
