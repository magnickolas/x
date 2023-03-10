package live_wallpaper

import (
	_ "embed"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/magnickolas/stopit"
	"github.com/magnickolas/x/util"
	e "github.com/pkg/errors"
	"github.com/postfinance/single"
	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/conf"
	"github.com/rwxrob/help"
	"github.com/rwxrob/vars"
)

var defs = map[string]string{
	"command":  `["xwinwrap", "-ni", "-fdt", "-sh", "rectangle", "-un", "-b", "-nf", "-ov", "-fs", "--", "mpv", "-wid", "WID", "--no-config", "--keepaspect=no", "--loop", "--no-border", "--vd-lavc-fast", "--x11-bypass-compositor=no", "--gapless-audio=yes", "--aid=no", "--vo=xv", "--hwdec=auto", "--really-quiet", "--input-ipc-server=/tmp/mpv-bg-socket", "/home/magnickolas/.wallpapers/live.mp4"]`,
	"startNow": "true",
}
var defKeys = util.Keys(defs)

func init() {
	util.Must(Z.Conf.SoftInit())
	util.Must(Z.Vars.SoftInit())
}

type cfg struct {
	cmd      *exec.Cmd
	startNow bool
}

func getConfig(x *Z.Cmd) (cfg, error) {
	args, err := util.Get[[]string](x, `command`)
	if err != nil {
		return cfg{}, err
	}
	cmd := exec.Command(args[0], args[1:]...)
	startNow, err := util.Get[bool](x, `startNow`)
	if err != nil {
		return cfg{}, err
	}
	return cfg{
		cmd:      cmd,
		startNow: startNow,
	}, nil
}

func runServer(x *Z.Cmd) error {
	s, err := single.New("live-wallpaper-server")
	if err != nil {
		return e.Wrap(err, "can't create singleton")
	}
	if err = s.Lock(); err != nil {
		fmt.Println("A server instance is already running")
		return nil
	}
	c, err := getConfig(x)
	if err != nil {
		return e.Wrap(err, "get config")
	}
	server, err := stopit.StopItServerWithFreePort(c.cmd, c.startNow)
	if err != nil {
		return err
	}
	x.Set("port", strconv.Itoa(server.Port))
	server.Run()
	panic("unreachable")
}

type action int

const (
	aStart action = iota
	aStop
)

func perform(a action, x *Z.Cmd) error {
	port, err := util.Get[int](x, "port")
	if err != nil {
		return err
	}
	stopit := stopit.StopIt{Port: port}
	switch a {
	case aStart:
		return stopit.Run()
	case aStop:
		return stopit.Stop()
	default:
		return e.Errorf("incorrect action %s", a)
	}
}

func start(x *Z.Cmd) error {
	return perform(aStart, x)
}

func stop(x *Z.Cmd) error {
	return perform(aStop, x)
}

var Cmd = &Z.Cmd{
	Name:    `live-wallpaper`,
	Summary: `Manage a live wallpaper`,
	Commands: []*Z.Cmd{
		help.Cmd, vars.Cmd, conf.Cmd,
		initCmd,
		runServerCmd, startCmd, stopCmd,
	},
	Shortcuts: util.ShortcutsFromDefs(defKeys),
}

var runServerCmd = &Z.Cmd{
	Name:     `run-server`,
	Summary:  `Run a process that can run/stop the live wallpaper command`,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, _ ...string) error {
		defer util.TrapPanic()
		util.Must(runServer(x.Caller))
		return nil
	},
}

var startCmd = &Z.Cmd{
	Name:     `start`,
	Aliases:  []string{`show`},
	Summary:  `Start wallpaper`,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, args ...string) error {
		defer util.TrapPanic()
		util.Must(start(x.Caller))
		return nil
	},
	Description: `
		Start wallpaper.
	`,
}

var stopCmd = &Z.Cmd{
	Name:     `stop`,
	Aliases:  []string{`kill`},
	Summary:  `Stop wallpaper`,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, args ...string) error {
		defer util.TrapPanic()
		util.Must(stop(x.Caller))
		return nil
	},
	Description: `
		Stop wallpaper.
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
