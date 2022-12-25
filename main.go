package main

import (
	"log"

	"github.com/magnickolas/x/battery_notify"
	"github.com/magnickolas/x/battery_status"
	"github.com/magnickolas/x/bookmark"
	"github.com/magnickolas/x/emoji"
	"github.com/magnickolas/x/setup_keyboard"

	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/conf"
	"github.com/rwxrob/good"
	"github.com/rwxrob/help"
	"github.com/rwxrob/pomo"
	"github.com/rwxrob/vars"
)

func init() {}

func main() {
	log.SetFlags(0)
	Z.AllowPanic = true

	Cmd.Run()
}

var Cmd = &Z.Cmd{
	Name:    `x`,
	Summary: `magnickolas' bonzai command tree`,
	Version: `v1.0.6`,
	Source:  `git@github.com:magnickolas/x.git`,

	Commands: []*Z.Cmd{
		help.Cmd, conf.Cmd, vars.Cmd, good.Cmd,
		pomo.Cmd,
		// personal
		battery_notify.Cmd, battery_status.Cmd,
		setup_keyboard.Cmd,
		emoji.Cmd, bookmark.Cmd,
	},

	Shortcuts: Z.ArgMap{},

	Description: `
		Magnickolas' Bonzai tree
		`,
}
