package layout

import (
	"bytes"
	_ "embed"
	"fmt"
	"os/exec"

	"github.com/magnickolas/x/util"
	e "github.com/pkg/errors"
	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/conf"
	"github.com/rwxrob/help"
	"github.com/rwxrob/vars"
	"golang.org/x/exp/slices"
)

var defs = map[string]string{
	"layouts":        `["us", "ru"]`,
	"extraLayout":    "ua",
	"previousLayout": "us",
}
var defKeys = util.Keys(defs)

func init() {
	util.Must(Z.Conf.SoftInit())
	util.Must(Z.Vars.SoftInit())
}

type layoutT string

type cfg struct {
	layouts        []layoutT
	extraLayout    layoutT
	previousLayout layoutT
}

func getConfig(x *Z.Cmd) (cfg, error) {
	layoutsS, err := util.Get[[]string](x, `layouts`)
	if err != nil {
		return cfg{}, err
	}
	extraLayout, err := util.Get[string](x, `extraLayout`)
	if err != nil {
		return cfg{}, err
	}
	previousLayout, err := util.Get[string](x, `previousLayout`)
	if err != nil {
		return cfg{}, err
	}
	layouts := make([]layoutT, 0, len(layoutsS))
	for _, layout := range layoutsS {
		layouts = append(layouts, layoutT(layout))
	}
	return cfg{
		layouts:        layouts,
		extraLayout:    layoutT(extraLayout),
		previousLayout: layoutT(previousLayout),
	}, nil
}

func getCurrentLayout() (layoutT, error) {
	output, err := exec.Command("xkb-switch").Output()
	if err != nil {
		return layoutT(""), e.Wrap(err, "get layout")
	}
	return layoutT(bytes.TrimSpace(output)), nil
}

func getNextLayout(curLayout layoutT, layouts []layoutT, extraLayout layoutT, previousLayout layoutT) (layoutT, error) {
	if curLayout == extraLayout {
		return previousLayout, nil
	}
	if i := slices.Index(layouts, curLayout); i != -1 {
		return layouts[(i+1)%len(layouts)], nil
	}
	return layoutT(""), e.Errorf("unknown layout %s", curLayout)
}

func getNextExtraLayout(curLayout layoutT, extraLayout layoutT, previousLayout layoutT) layoutT {
	if curLayout == extraLayout {
		return previousLayout
	}
	return extraLayout
}

func switchLayout(layout layoutT) error {
	return exec.Command("xkb-switch", "-s", string(layout)).Run()
}

func switchL(x *Z.Cmd) error {
	c, err := getConfig(x)
	if err != nil {
		return e.Wrap(err, "get config")
	}
	curLayout, err := getCurrentLayout()
	if err != nil {
		return e.Wrap(err, "get current layout")
	}
	nextLayout, err := getNextLayout(curLayout, c.layouts, c.extraLayout, c.previousLayout)
	if err != nil {
		return e.Wrap(err, "get next layout")
	}
	x.Set(`previousLayout`, string(nextLayout))
	return switchLayout(nextLayout)
}

func switchExtra(x *Z.Cmd) error {
	c, err := getConfig(x)
	if err != nil {
		return e.Wrap(err, "get config")
	}
	curLayout, err := getCurrentLayout()
	if err != nil {
		return e.Wrap(err, "get current layout")
	}
	nextLayout := getNextExtraLayout(curLayout, c.extraLayout, c.previousLayout)
	return switchLayout(nextLayout)
}

func printCurrentLayout() error {
	layout, err := getCurrentLayout()
	if err != nil {
		return e.Wrap(err, "get layout")
	}
	fmt.Println(layout)
	return nil
}

var Cmd = &Z.Cmd{
	Name:    `layout`,
	Summary: `Manage keyboard layouts`,
	Commands: []*Z.Cmd{
		printCmd,
		help.Cmd, vars.Cmd, conf.Cmd,
		initCmd,
		switchCmd, extraCmd,
	},
	Shortcuts: util.ShortcutsFromDefs(defKeys),
}

var printCmd = &Z.Cmd{
	Name:     `print`,
	Summary:  `Print current layout`,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, _ ...string) error {
		defer util.TrapPanic()
		util.Must(printCurrentLayout())
		return nil
	},
}

var switchCmd = &Z.Cmd{
	Name:     `switch`,
	Summary:  `Switch to next layout`,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, args ...string) error {
		defer util.TrapPanic()
		util.Must(switchL(x.Caller))
		return nil
	},
}

var extraCmd = &Z.Cmd{
	Name:     `extra`,
	Summary:  `Switch to extra layout`,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, args ...string) error {
		defer util.TrapPanic()
		util.Must(switchExtra(x.Caller))
		return nil
	},
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
