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

var (
	Layout    = "us,ru"
	Options   = "grp:alt_space_toggle,compose:ralt"
	Delay     = "220"
	Rate      = "50"
	PlaySound = "true"
)

func init() {
	Z.Conf.SoftInit()
	Z.Vars.SoftInit()
	Z.Dynamic[`dLayout`] = func() string { return Layout }
	Z.Dynamic[`dOptions`] = func() string { return Options }
	Z.Dynamic[`dDelay`] = func() string { return Delay }
	Z.Dynamic[`dRate`] = func() string { return Rate }
	Z.Dynamic[`dPlaySound`] = func() string { return PlaySound }
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
	layout, err := x.Get(`layout`)
	if err != nil {
		return cfg{}, e.Wrap(err, "get layout")
	}
	options, err := x.Get(`options`)
	if err != nil {
		return cfg{}, e.Wrap(err, "get options")
	}
	delayS, err := x.Get(`delay`)
	if err != nil {
		return cfg{}, e.Wrap(err, "get delay")
	}
	delay, err := strconv.Atoi(delayS)
	if err != nil {
		return cfg{}, e.Wrap(err, "parse delay")
	}
	rateS, err := x.Get(`rate`)
	if err != nil {
		return cfg{}, e.Wrap(err, "get rate")
	}
	rate, err := strconv.Atoi(rateS)
	if err != nil {
		return cfg{}, e.Wrap(err, "parse rate")
	}
	playSoundS, err := x.Get(`playSound`)
	if err != nil {
		return cfg{}, e.Wrap(err, "get playSound")
	}
	playSound, err := strconv.ParseBool(playSoundS)
	if err != nil {
		return cfg{}, e.Wrap(err, "parse playSound")
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
	Shortcuts: Z.ArgMap{
		`layout`:    {`var`, `set`, `layout`},
		`options`:   {`var`, `set`, `options`},
		`delay`:     {`var`, `set`, `delay`},
		`rate`:      {`var`, `set`, `rate`},
		`playSound`: {`var`, `set`, `playSound`},
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

            layout - {{dLayout}}
            options - {{dOptions}}
            delay - {{dDelay}}
            rate - {{dRate}}
            playSound - {{dPlaySound}}
	`,
	Call: func(x *Z.Cmd, _ ...string) error {
		defs := map[string]string{
			`layout`:    Layout,
			`options`:   Options,
			`delay`:     Delay,
			`rate`:      Rate,
			`playSound`: PlaySound,
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
