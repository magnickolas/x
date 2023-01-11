package brightness

import (
	_ "embed"
	"fmt"
	"os/exec"
	"strings"

	"github.com/magnickolas/x/util"
	e "github.com/pkg/errors"
	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/conf"
	"github.com/rwxrob/help"
	"github.com/rwxrob/vars"
)

var defs = map[string]string{
	"delta": "6000",
	"scale": "1.5",
}
var defKeys = util.Keys(defs)

func init() {
	util.Must(Z.Conf.SoftInit())
	util.Must(Z.Vars.SoftInit())
}

type cfg struct {
	delta brightnessT
	scale scaleT
}

func getConfig(x *Z.Cmd) (cfg, error) {
	delta, err := util.Get[uint](x, `delta`)
	if err != nil {
		return cfg{}, err
	}
	scale, err := util.Get[float64](x, `scale`)
	if err != nil {
		return cfg{}, err
	}
	return cfg{
		delta: brightnessT(delta),
		scale: scaleT(scale),
	}, nil
}

func getBrightness() (brightnessT, error) {
	output, err := exec.Command("brightnessctl", "get").Output()
	if err != nil {
		return 0, e.Wrap(err, "get brightness")
	}
	brightness, err := util.ParseUint(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, e.Wrap(err, "parse brightness")
	}
	return brightnessT(brightness), nil
}

func setBrightness(brightness uint) error {
	fmt.Println(brightness)
	_, err := exec.Command("brightnessctl", "set", fmt.Sprintf("%d", brightness)).Output()
	if err != nil {
		return e.Wrap(err, "set brightness")
	}
	return nil
}

type brightnessT int
type scaleT float64
type alterMode int

const (
	alterInc alterMode = iota
	alterDec
)

func alterBrightness(mode alterMode, delta brightnessT, scale scaleT) (brightnessT, error) {
	min_step := brightnessT(12)
	brightness, err := getBrightness()
	if err != nil {
		return 0, e.Wrap(err, "get brightness")
	}
	switch mode {
	case alterInc:
		return util.Min(
			brightnessT(scaleT(brightness)*scale)+min_step,
			brightness+delta,
		), err
	case alterDec:
		return util.Max(
			min_step,
			brightnessT(scaleT(brightness-min_step)/scale),
			brightness-delta,
		), err
	default:
		return 0, e.Wrap(err, "incorrect alter mode")
	}
}

func printBrightness() error {
	brightness, err := getBrightness()
	if err != nil {
		return e.Wrap(err, "get brightness")
	}
	fmt.Print(brightness)
	return nil
}

func alter(c cfg, mode alterMode) error {
	brightness, err := alterBrightness(mode, c.delta, c.scale)
	if err != nil {
		return e.Wrap(err, "inc brightness")
	}
	err = setBrightness(uint(brightness))
	if err != nil {
		return e.Wrap(err, "set brightness")
	}
	err = printBrightness()
	if err != nil {
		return e.Wrap(err, "print brightness")
	}
	return nil
}

func inc(x *Z.Cmd) error {
	c, err := getConfig(x)
	if err != nil {
		return e.Wrap(err, "get config")
	}
	return alter(c, alterInc)
}

func dec(x *Z.Cmd) error {
	c, err := getConfig(x)
	if err != nil {
		return e.Wrap(err, "get config")
	}
	return alter(c, alterDec)
}

var Cmd = &Z.Cmd{
	Name:    `brightness`,
	Summary: `Change screen brightness`,
	Commands: []*Z.Cmd{
		printCmd,
		help.Cmd, vars.Cmd, conf.Cmd,
		initCmd,
		incCmd, decCmd,
	},
	Shortcuts: util.ShortcutsFromDefs(defKeys),
}

var printCmd = &Z.Cmd{
	Name:     `print`,
	Summary:  `Print current brightness`,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, _ ...string) error {
		defer util.TrapPanic()
		util.Must(printBrightness())
		return nil
	},
}

var incCmd = &Z.Cmd{
	Name:     `inc`,
	Summary:  `Increase brightness`,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, args ...string) error {
		defer util.TrapPanic()
		util.Must(inc(x.Caller))
		return nil
	},
	Description: `
		Increase brightness.
	`,
}

var decCmd = &Z.Cmd{
	Name:     `dec`,
	Summary:  `Decrease brightness`,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, args ...string) error {
		defer util.TrapPanic()
		util.Must(dec(x.Caller))
		return nil
	},
	Description: `
		Decrease brightness.
	`,
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
