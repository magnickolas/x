package bookmark

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/magnickolas/x/util"
	"github.com/ncruces/zenity"
	e "github.com/pkg/errors"
	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/conf"
	"github.com/rwxrob/help"
	"github.com/rwxrob/vars"
	"golang.design/x/clipboard"
	"golang.org/x/term"
)

var defs = map[string]string{
	"askForDescription":    "true",
	"bookmarkFile":         Z.Dynamic[`homedir`].(func(...string) string)(".bookmarks"),
	"pickerCmd":            `["rofi", "-dmenu", "-i", "-p", "Choose bookmark"]`,
	"notify":               "true",
	"notifyDuration":       "3s",
	"typeKeys":             "false",
	"unixPrimarySelection": "false",
	"editor":               "",
	"sort":                 "-date",
}
var defKeys = util.Keys(defs)
var initDefs = "bookmark_defs"

func init() {
	util.Must(Z.Conf.SoftInit())
	util.Must(Z.Vars.SoftInit())
	util.InitFromDefs(Z.Dynamic, initDefs, defs, defKeys)
}

func prompt(content string) (string, error) {
	return zenity.Entry(
		fmt.Sprintf("Description of `%s`", content),
		zenity.Title("Bookmark description"),
		zenity.Width(300),
	)
}

func getBookmarkContent(unixPrimarySelection bool) (string, error) {
	var text []byte
	if unixPrimarySelection {
		var err error
		text, err = exec.Command("xsel", "-o").Output()
		if err != nil {
			return "", e.Wrap(err, "run xsel to get primary selection")
		}
	} else {
		err := clipboard.Init()
		if err != nil {
			return "", e.Wrap(err, "init clipboard")
		}
		text = clipboard.Read(clipboard.FmtText)
	}
	return string(bytes.TrimSpace(text)), nil
}

type bookmark struct {
	content     string
	description string
}

func (b bookmark) serialize() string {
	line := b.content
	if b.description != "" {
		line += " # " + b.description
	}
	return line
}

func (b *bookmark) deserialize(line string) {
	split := strings.SplitN(line, " # ", 2)
	if len(split) > 1 {
		b.description = split[1]
	}
	b.content = strings.TrimSpace(split[0])
}

func getBookmarkDescription(content string) (string, error) {
	description, err := prompt(content)
	if err != nil {
		return "", e.Wrap(err, "prompt for description")
	}
	return description, nil
}

func doesBookmarkExist(content string, bookmarkFile string) (bool, error) {
	bytes, err := os.ReadFile(bookmarkFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		} else {
			return false, e.Wrap(err, "read bookmark file")
		}
	}
	re := regexp.MustCompile(
		fmt.Sprintf(`(?m)^%s(\s+#.*)?$`,
			regexp.QuoteMeta(string(content))))
	return re.Find(bytes) != nil, nil
}

func addBookmarkToFile(b bookmark, path string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return e.Wrap(err, "open bookmark file")
	}
	defer f.Close()

	line := b.serialize()
	bytes, err := ioutil.ReadAll(f)
	if len(bytes) > 0 && bytes[len(bytes)-1] != '\n' {
		line = "\n" + line
	}
	_, err = f.WriteString(line)
	return e.Wrap(err, "write bookmark file")
}

func addBookmarkIfNotExists(c cfg) error {
	content, err := getBookmarkContent(c.unixPrimarySelection)
	if err != nil {
		return e.Wrap(err, "get bookmark content")
	}
	if content == "" {
		return nil
	}
	b, err := addBookmarkContentIfNotExists(content, c)
	if err != nil {
		return e.Wrap(err, "add bookmark content")
	}
	if c.notify {
		err = notifyBookmarkAdded(b.content, c.notifyDuration)
		if err != nil {
			return e.Wrap(err, "notify bookmark added")
		}
	}
	return nil
}

func addBookmarkContentIfNotExists(content string, c cfg) (bookmark, error) {
	exists, err := doesBookmarkExist(content, c.bookmarkFile)
	if err != nil {
		return bookmark{}, e.Wrap(err, "check if bookmark exists")
	}
	if exists {
		return bookmark{}, nil
	}
	var description string
	if c.askForDescription {
		description, err = getBookmarkDescription(content)
		if err != nil {
			return bookmark{}, e.Wrap(err, "get bookmark description")
		}
	}
	b := bookmark{
		content:     content,
		description: description,
	}
	err = addBookmarkToFile(b, c.bookmarkFile)
	if err != nil {
		return b, e.Wrap(err, "add bookmark to file")
	}
	return b, nil
}

func defaultPickLine(lines []string) (string, error) {
	height := util.Min(len(lines)*35, 600)
	width := util.Min(
		util.FoldLeft1(util.Max2[int], util.Map(util.Len, lines))*10,
		800,
	)
	line, err := zenity.List(
		"Choose bookmark",
		lines,
		zenity.Title("Choose bookmark"),
		zenity.Width(uint(width)),
		zenity.Height(uint(height)),
	)
	if err != nil {
		return "", e.Wrap(err, "pick bookmark")
	}
	return line, nil
}

func pickWithCmd(text string, pickerCmd []string) (string, error) {
	cmd := exec.Command(pickerCmd[0], pickerCmd[1:]...)
	cmd.Stdin = strings.NewReader(text)
	bs, err := cmd.Output()
	if err != nil {
		return "", e.Wrap(err, "get output of custom pick command")
	}
	return string(bytes.TrimSpace(bs)), nil
}

func pickBookmark(pickerCmd []string, bookmarkFile string) (bookmark, error) {
	fileContent, err := os.ReadFile(bookmarkFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return bookmark{}, nil
		} else {
			return bookmark{}, e.Wrap(err, "read bookmark file")
		}
	}
	if len(fileContent) == 0 {
		return bookmark{}, nil
	}
	s := string(fileContent)
	var chosenLine string
	if len(pickerCmd) != 0 {
		chosenLine, err = pickWithCmd(s, pickerCmd)
		if err != nil {
			return bookmark{}, e.Wrap(err, "pick line with custom command")
		}
	} else {
		lines := strings.Split(s, "\n")
		chosenLine, err = defaultPickLine(lines)
		if err != nil {
			return bookmark{}, e.Wrap(err, "run default pick line command")
		}
	}
	var b bookmark
	b.deserialize(chosenLine)
	if err != nil {
		return bookmark{}, e.Wrap(err, "parse bookmark line")
	}
	return b, nil
}

func typeKeys(keys string) error {
	return exec.Command("xdotool", "type", keys).Run()
}

func outputBookmark(b bookmark, doTypeKeys bool) error {
	if b == (bookmark{}) {
		return nil
	}
	err := clipboard.Init()
	if err != nil {
		return e.Wrap(err, "init clipboard")
	}
	line := b.content + " # " + b.description
	if term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Println(line)
	}
	if doTypeKeys {
		err := typeKeys(b.content)
		if err != nil {
			return e.Wrap(err, "type keys")
		}
	}
	done := clipboard.Write(clipboard.FmtText, []byte(b.content))
	<-done
	return nil
}

func notifyBookmarkAdded(content string, duration time.Duration) error {
	return util.Notify(
		fmt.Sprintf("New bookmark\n`%s`", content),
		util.Low, uint(duration.Milliseconds()), "",
	)
}

type cfg struct {
	askForDescription    bool
	bookmarkFile         string
	pickerCmd            []string
	notify               bool
	notifyDuration       time.Duration
	typeKeys             bool
	unixPrimarySelection bool
	sort                 string
}

type cfgEdit struct {
	editorPath string
	c          cfg
}

func getConfig(x *Z.Cmd) (cfg, error) {
	bookmarkFile, err := util.Get[string](x, `bookmarkFile`)
	if err != nil {
		return cfg{}, err
	}
	pickerCmd, err := util.Get[[]string](x, `pickerCmd`)
	if err != nil {
		return cfg{}, err
	}
	askForDescription, err := util.Get[bool](x, `askForDescription`)
	if err != nil {
		return cfg{}, err
	}
	notify, err := util.Get[bool](x, `notify`)
	if err != nil {
		return cfg{}, err
	}
	notifyDuration, err := util.Get[time.Duration](x, `notifyDuration`)
	if err != nil {
		return cfg{}, err
	}
	typeKeys, err := util.Get[bool](x, `typeKeys`)
	if err != nil {
		return cfg{}, err
	}
	unixPrimarySelection, err := util.Get[bool](x, `unixPrimarySelection`)
	if err != nil {
		return cfg{}, err
	}
	sort, err := util.GetEnum(x, `sort`, []string{"-date", "date"})
	if err != nil {
		return cfg{}, err
	}
	return cfg{
		askForDescription:    askForDescription,
		bookmarkFile:         bookmarkFile,
		pickerCmd:            pickerCmd,
		notify:               notify,
		notifyDuration:       notifyDuration,
		typeKeys:             typeKeys,
		unixPrimarySelection: unixPrimarySelection,
		sort:                 sort,
	}, nil
}

func getEditConfig(x *Z.Cmd) (cfgEdit, error) {
	c, err := getConfig(x)
	if err != nil {
		return cfgEdit{}, err
	}
	editor, err := util.Get[string](x, `editor`)
	if err != nil {
		return cfgEdit{}, err
	}
	editorPath, err := util.FindEditor(editor)
	if err != nil {
		return cfgEdit{}, err
	}
	return cfgEdit{
		editorPath: editorPath,
		c:          c,
	}, nil
}

func cmd(x *Z.Cmd) error {
	c, err := getConfig(x)
	if err != nil {
		return e.Wrap(err, "get config")
	}
	b, err := pickBookmark(c.pickerCmd, c.bookmarkFile)
	if err != nil {
		return e.Wrap(err, "pick bookmark")
	}
	return outputBookmark(b, c.typeKeys)
}

func add(x *Z.Cmd, args ...string) error {
	c, err := getConfig(x.Caller)
	if err != nil {
		return e.Wrap(err, "get config")
	}
	if len(args) == 0 {
		util.Must(addBookmarkIfNotExists(c))
		return nil
	}
	content := args[0]
	_, err = addBookmarkContentIfNotExists(content, c)
	util.Must(err)
	return nil
}

func edit(x *Z.Cmd) error {
	c, err := getEditConfig(x.Caller)
	if err != nil {
		return e.Wrap(err, "get config")
	}
	return util.EditFile(c.c.bookmarkFile, c.editorPath)
}

var Cmd = &Z.Cmd{
	Name:    `bookmark`,
	Summary: `Manage bookmarks`,
	Commands: []*Z.Cmd{
		help.Cmd, vars.Cmd, conf.Cmd,
		initCmd,
		addCmd, editCmd,
	},
	Call: func(x *Z.Cmd, args ...string) error {
		defer util.TrapPanic()
		util.Must(cmd(x))
		return nil
	},
	Shortcuts: util.ShortcutsFromDefs(defKeys),
}

var addCmd = &Z.Cmd{
	Name:     `add`,
	Summary:  `add a bookmark`,
	Usage:    `[content]`,
	MaxArgs:  1,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, args ...string) error {
		defer util.TrapPanic()
		util.Must(add(x, args...))
		return nil
	},
	Description: `
		Add a bookmark.

		If content is not specified, the current clipboard content is used.
	`,
}

var editCmd = &Z.Cmd{
	Name:     `edit`,
	Summary:  `edit bookmarks file`,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, args ...string) error {
		defer util.TrapPanic()
		util.Must(edit(x))
		return nil
	},
	Description: `
		Edit a bookmark.

		The editor is 'editor' variable is set, else $VISUAL if set, else $EDITOR if set, else {{cmd "vi"}}.
	`,
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
