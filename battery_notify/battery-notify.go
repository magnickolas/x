package battery_notify

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/magnickolas/x/util"
	e "github.com/pkg/errors"
	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/conf"
	"github.com/rwxrob/help"
	"github.com/rwxrob/vars"
)

var (
	CacheFile = "/tmp/tmp.battery_notify_timestamp"
	Threshold = "15"
	Delay     = "20m"
	PlaySound = "true"
)

func init() {
	Z.Conf.SoftInit()
	Z.Vars.SoftInit()
	Z.Dynamic[`dCacheFile`] = func() string { return CacheFile }
	Z.Dynamic[`dThreshold`] = func() string { return Threshold }
	Z.Dynamic[`dDelay`] = func() string { return Delay }
	Z.Dynamic[`dPlaySound`] = func() string { return PlaySound }
}

func setup_env() error {
	err := os.Setenv("DISPLAY", ":0")
	if err != nil {
		return err
	}
	err = os.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path=/run/user/1000/bus")
	return err
}

//go:embed assets/battery.ico
var batteryImage []byte

//go:embed assets/battery.mp3
var batterySound []byte

func getBatteryImageFile() (*os.File, error) {
	f, err := os.CreateTemp("", "battery*.ico")
	if err != nil {
		return nil, e.Wrap(err, "create temp file")
	}
	_, err = f.Write(batteryImage)
	if err != nil {
		return nil, e.Wrap(err, "write battery image")
	}
	return f, nil
}

func batteryNotify(c cfg) error {
	setup_env()

	perm := os.FileMode(0644)
	var prev_ts int64
	if prev_ts_str, err := os.ReadFile(c.cacheFile); err == nil {
		stat, err := os.Stat(c.cacheFile)
		if err != nil {
			return e.Wrap(err, "stat cache file")
		}
		perm = stat.Mode().Perm()
		prev_ts, err = strconv.ParseInt(string(prev_ts_str), 10, 64)
		if err != nil {
			return e.Wrap(err, "invalid cached timestamp")
		}
	} else if errors.Is(err, os.ErrNotExist) {
		prev_ts = 0
	} else {
		return e.Wrap(err, "read cache file")
	}

	ts := time.Now().Unix()

	if ts-prev_ts >= c.delaySeconds {
		info, err := util.GetBatteryInfo()
		if err != nil {
			return e.Wrap(err, "get battery info")
		}
		if info.Status == util.Discharging && int(info.Level) < c.threshold {
			batteryImageFile, err := getBatteryImageFile()
			if err != nil {
				return e.Wrap(err, "get battery image file")
			}
			defer os.RemoveAll(batteryImageFile.Name())
			err = exec.Command(
				"notify-send",
				"-u", "critical",
				"-i", batteryImageFile.Name(),
				fmt.Sprintf(
					"Battery is at %d%%", info.Level,
				),
			).Run()
			if err != nil {
				return e.Wrap(err, "send notification")
			}
			if c.playSound {
				err = util.PlaySoundBlock(batterySound)
				if err != nil {
					return e.Wrap(err, "play sound")
				}
			}
			err = os.WriteFile(
				c.cacheFile,
				[]byte(fmt.Sprint(ts)),
				perm,
			)
			if err != nil {
				return e.Wrap(err, "failed to write cache file")
			}
		}
	}
	return nil
}

type cfg struct {
	cacheFile    string
	delaySeconds int64
	threshold    int
	playSound    bool
}

type Server struct {
	Name string
}

func getConfig(x *Z.Cmd) (cfg, error) {
	cacheFile, err := x.Get(`cacheFile`)
	if err != nil {
		return cfg{}, err
	}
	thresholdS, err := x.Get(`threshold`)
	if err != nil {
		return cfg{}, err
	}
	threshold, err := strconv.Atoi(thresholdS)
	if err != nil {
		return cfg{}, e.Wrap(err, "parse threshold")
	}
	delayS, err := x.Get(`delay`)
	if err != nil {
		return cfg{}, err
	}
	delay, err := time.ParseDuration(delayS)
	if err != nil {
		return cfg{}, e.Wrap(err, "parse delay")
	}
	playSoundS, err := x.Get(`playSound`)
	if err != nil {
		return cfg{}, err
	}
	playSound, err := strconv.ParseBool(playSoundS)
	if err != nil {
		return cfg{}, e.Wrap(err, "parse playSound")
	}
	return cfg{
		cacheFile:    cacheFile,
		threshold:    threshold,
		delaySeconds: int64(delay.Seconds()),
		playSound:    playSound,
	}, nil
}

func cmd(x *Z.Cmd) error {
	c, err := getConfig(x)
	if err != nil {
		return e.Wrap(err, "get config")
	}
	return batteryNotify(c)
}

var Cmd = &Z.Cmd{
	Name:    `battery-notify`,
	Summary: `notify if the battery is low`,
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
		`cacheFile`: {`var`, `set`, `cacheFile`},
		`threshold`: {`var`, `set`, `threshold`},
		`delay`:     {`var`, `set`, `delay`},
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

		    cacheFile - {{dCacheFile}}
		    threshold - {{dThreshold}}
		    delay     - {{dDelay}}
		    playSound - {{dPlaySound}}
	`,
	Call: func(x *Z.Cmd, _ ...string) error {
		defs := map[string]string{
			`cacheFile`: CacheFile,
			`threshold`: Threshold,
			`delay`:     Delay,
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
