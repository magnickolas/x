package setup_keyboard

import (
	_ "embed"
	"os/exec"
	"strconv"

	"github.com/magnickolas/x/util"
	e "github.com/pkg/errors"
	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/conf"
	"github.com/rwxrob/help"
	"github.com/rwxrob/vars"
)

var defs = map[string]string{
	"layout":    "us,ru,ua",
	"options":   "grp:compose:ralt",
	"delay":     "220",
	"rate":      "50",
	"playSound": "true",
}
var defKeys = util.Keys(defs)

func init() {
	util.Must(Z.Conf.SoftInit())
	util.Must(Z.Vars.SoftInit())
}

//go:embed assets/sound.mp3
var sound []byte

func setupKeyboard(c cfg) error {
	err := exec.Command(
		"xset", "r", "rate",
		strconv.Itoa(c.delay), strconv.Itoa(c.rate),
	).Run()
	if err != nil {
		return e.Wrap(err, "run xset")
	}
	err = exec.Command(
		"setxkbmap",
		"-layout", c.layout,
		"-option", c.options,
	).Run()
	if err != nil {
		return e.Wrap(err, "run setxkbmap")
	}
	if c.playSound {
		util.PlaySoundBlock(sound)
	}
	return nil
}

type cfg struct {
	layout    string
	options   string
	delay     int
	rate      int
	playSound bool
}

func getConfig(x *Z.Cmd) (cfg, error) {
	layout, err := util.Get[string](x, `layout`)
	if err != nil {
		return cfg{}, err
	}
	options, err := util.Get[string](x, `options`)
	if err != nil {
		return cfg{}, err
	}
	delay, err := util.Get[int](x, `delay`)
	if err != nil {
		return cfg{}, err
	}
	rate, err := util.Get[int](x, `rate`)
	if err != nil {
		return cfg{}, err
	}
	playSound, err := util.Get[bool](x, `playSound`)
	if err != nil {
		return cfg{}, err
	}
	return cfg{
		layout:    layout,
		options:   options,
		delay:     delay,
		rate:      rate,
		playSound: playSound,
	}, nil
}

func cmd(x *Z.Cmd) error {
	c, err := getConfig(x)
	if err != nil {
		return e.Wrap(err, "get config")
	}
	return setupKeyboard(c)
}

var Cmd = &Z.Cmd{
	Name:    `setup-keyboard`,
	Summary: `setup all connected keyboards`,
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
