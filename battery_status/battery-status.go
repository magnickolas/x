package battery_status

import (
	_ "embed"
	"fmt"

	"github.com/magnickolas/x/util"
	e "github.com/pkg/errors"
	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/conf"
	"github.com/rwxrob/help"
	"github.com/rwxrob/vars"
)

const (
	Charging    = "⚡"
	Discharging = "🔋"
	LowBattery  = "🪫"
	NotCharging = "✔"
	Threshold   = "20"
)

func init() {
	Z.Conf.SoftInit()
	Z.Vars.SoftInit()
	Z.Dynamic[`dCharging`] = func() string { return Charging }
	Z.Dynamic[`dDischarging`] = func() string { return Discharging }
	Z.Dynamic[`dLowBattery`] = func() string { return LowBattery }
	Z.Dynamic[`dNotCharging`] = func() string { return NotCharging }
	Z.Dynamic[`dThreshold`] = func() string { return Threshold }
}

func outputBatteryStatus(c cfg) error {
	info, err := util.GetBatteryInfo()
	if err != nil {
		e.Wrap(err, "get battery info")
	}
	var statusSymbol string
	if info.Status == util.Discharging {
		if int(info.Level) <= c.threshold {
			statusSymbol = LowBattery
		} else {
			statusSymbol = Discharging
		}
	} else if info.Status == util.Charging {
		statusSymbol = Charging
	} else if info.Status == util.NotCharging {
		statusSymbol = NotCharging
	} else {
		return e.New("unknown status")
	}
	_, err = fmt.Printf("%s %d%%", statusSymbol, int(info.Level))
	if err != nil {
		return e.Wrap(err, "print battery status")
	}
	return nil
}

type cfg struct {
	charging    string
	discharging string
	lowBattery  string
	notCharging string
	threshold   int
}

func getConfig(x *Z.Cmd) (cfg, error) {
	charging, err := util.Get(x, `charging`)
	if err != nil {
		return cfg{}, err
	}
	discharging, err := util.Get(x, `discharging`)
	if err != nil {
		return cfg{}, err
	}
	lowBattery, err := util.Get(x, `lowBattery`)
	if err != nil {
		return cfg{}, err
	}
	notCharging, err := util.Get(x, `notCharging`)
	if err != nil {
		return cfg{}, err
	}
	threshold, err := util.GetInt(x, `threshold`)
	if err != nil {
		return cfg{}, err
	}
	return cfg{
		charging:    charging,
		discharging: discharging,
		lowBattery:  lowBattery,
		notCharging: notCharging,
		threshold:   threshold,
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
	Shortcuts: Z.ArgMap{
		`charging`:    {`var`, `set`, `charging`},
		`discharging`: {`var`, `set`, `discharging`},
		`lowBattery`:  {`var`, `set`, `lowBattery`},
		`notCharging`: {`var`, `set`, `notCharging`},
		`threshold`:   {`var`, `set`, `threshold`},
	},
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

            threshold - {{dThreshold}}
            charging - {{dCharging}}
            discharging - {{dDischarging}}
            lowBattery - {{dLowBattery}}
            notCharging - {{dNotCharging}}
	`,
	Call: func(x *Z.Cmd, _ ...string) error {
		defs := map[string]string{
			`charging`:    Charging,
			`discharging`: Discharging,
			`lowBattery`:  LowBattery,
			`notCharging`: NotCharging,
			`threshold`:   Threshold,
		}
		for key, def := range defs {
			val, _ := x.Caller.C(key)
			if val == "null" {
				val = def
			}
			x.Caller.Set(key, val)
		}
		return nil
	},
}
