package battery_status

import (
	_ "embed"
	"fmt"
	"sort"

	"github.com/magnickolas/x/util"
	e "github.com/pkg/errors"
	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/conf"
	"github.com/rwxrob/help"
	"github.com/rwxrob/vars"
)

var defs = map[string]string{
	"charging":    `{"20": "  ", "40": "  ", "60": "  ", "80": "  ", "100": "  "}`,
	"discharging": `{"20": " ", "40": " ", "60": " ", "80": " ", "100": " "}`,
	"notCharging": `{"100": ""}`,
	"format":      "{status} {level}%",
}
var defKeys = util.Keys(defs)
var initDefs = "battery_status_defs"

func init() {
	util.Must(Z.Conf.SoftInit())
	util.Must(Z.Vars.SoftInit())
	util.InitFromDefs(Z.Dynamic, initDefs, defs, defKeys)
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
	var statusSymbol string
	for _, threshold := range thresholds {
		if int(info.Level) <= threshold {
			statusSymbol = levelMap[threshold]
			break
		}
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
	charging    map[int]string
	discharging map[int]string
	notCharging map[int]string
	format      string
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
	return cfg{
		charging:    charging,
		discharging: discharging,
		notCharging: notCharging,
		format:      format,
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

	Description: `
		The {{cmd .Name}} command sets all cached variables to their initial
		values. Any variable name from {{cmd "conf"}} will be used to
		initialize if defined.  Otherwise, the following hard-coded package
		globals will be used instead:

{{` + initDefs + `}}`,
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
