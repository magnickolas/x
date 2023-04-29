package emoji

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
	"golang.design/x/clipboard"
	"gopkg.in/yaml.v2"
)

var defs = map[string]string{
	"pickerCmd": `["rofi", "-dmenu", "-p", "emoji", "-i", "-sort"]`,
}
var defKeys = util.Keys(defs)

type cfg struct {
	pickerCmd []string
}

//go:embed assets/emoji.yaml
var emojisYaml string

func getMap() (yaml.MapSlice, error) {
	emojis := yaml.MapSlice{}
	err := yaml.Unmarshal([]byte(emojisYaml), &emojis)
	if err != nil {
		return nil, e.Wrap(err, "unmarshal emoji list")
	}
	return emojis, nil
}

type emojiT struct {
	name string
	char string
}

func getEmojis() ([]emojiT, error) {
	emojis, err := getMap()
	if err != nil {
		return nil, e.Wrap(err, "get emojis map")
	}
	emojiList := make([]emojiT, 0, len(emojis))
	for _, kv := range emojis {
		emoji, err := util.ParseUnicode(kv.Key.(string))
		if err != nil {
			return nil, e.Wrapf(err, "parse unicode %s", kv.Key.(string))
		}
		emojiList = append(emojiList, emojiT{
			name: kv.Value.(string),
			char: emoji,
		})
	}
	return emojiList, nil
}

func pasteEmoji(emoji string) error {
	err := clipboard.Init()
	if err != nil {
		return e.Wrap(err, "init clipboard")
	}
	done := clipboard.Write(clipboard.FmtText, []byte(emoji))
	<-done
	return nil
}

func selectEmoji(pickerCmd []string) (string, error) {
	emojis, err := getEmojis()
	if err != nil {
		return "", e.Wrap(err, "get emojis")
	}
	cmd := exec.Command(pickerCmd[0], pickerCmd[1:]...)
	cmd.Stdin = strings.NewReader(strings.Join(util.Map(func(e emojiT) string {
		return fmt.Sprintf("%s %s", e.char, e.name)
	}, emojis), "\n"))
	out, err := cmd.Output()
	if err != nil {
		return "", e.Wrap(err, "run picker command")
	}

	s := strings.Split(string(out), " ")[0]
	return s, nil
}

func selectAndYankEmoji(pickerCmd []string) error {
	emoji, err := selectEmoji(pickerCmd)
	if err != nil {
		return e.Wrap(err, "select emoji")
	}
	return pasteEmoji(emoji)
}

func getConfig(x *Z.Cmd) (cfg, error) {
	pickerCmd, err := util.Get[[]string](x, `pickerCmd`)
	if err != nil {
		return cfg{}, err
	}
	return cfg{
		pickerCmd: pickerCmd,
	}, nil
}

var Cmd = &Z.Cmd{
	Name:     `emoji`,
	Summary:  `choose and paste emoji`,
	Commands: []*Z.Cmd{help.Cmd, vars.Cmd, conf.Cmd, previewCmd},
	Call: func(x *Z.Cmd, args ...string) error {
		defer util.TrapPanic()
		c, err := getConfig(x)
		if err != nil {
			return e.Wrap(err, "get config")
		}
		util.Must(selectAndYankEmoji(c.pickerCmd))
		return nil
	},
}

var previewCmd = &Z.Cmd{
	Name:     `preview`,
	Summary:  `Preview emojis`,
	Usage:    `[width]`,
	NumArgs:  1,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, args ...string) error {
		defer util.TrapPanic()
		width, err := util.ParseUint(args[0])
		util.Must(err)
		util.Must(preview(width))
		return nil
	},
}

func preview(width uint) error {
	emojis, err := getEmojis()
	if err != nil {
		return e.Wrap(err, "get emojis")
	}
	var w uint
	for _, e := range emojis {
		fmt.Printf("%s ", e.char)
		w++
		if w == width {
			fmt.Println()
			w = 0
		}
	}
	return nil
}
