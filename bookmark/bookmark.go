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

var (
	AskForDescription = "true"
	BookmarkFile      = Z.Dynamic[`homedir`].(func(...string) string)(".bookmarks")
	PickerCmd         = `["rofi", "-dmenu", "-p", "Choose bookmark"]`
	Notify            = "true"
	NotifyDuration    = "3s"
)

func init() {
	Z.Conf.SoftInit()
	Z.Vars.SoftInit()
	Z.Dynamic[`dAskForDescription`] = func() string { return AskForDescription }
	Z.Dynamic[`dBookmarkFile`] = func() string { return BookmarkFile }
	Z.Dynamic[`dPickerCmd`] = func() string { return PickerCmd }
	Z.Dynamic[`dNotify`] = func() string { return Notify }
}

func prompt(content string) (string, error) {
	return zenity.Entry(
		fmt.Sprintf("Description of `%s`", content),
		zenity.Title("Bookmark description"),
		zenity.Width(300),
	)
}

func getBookmarkContent() (string, error) {
	err := clipboard.Init()
	if err != nil {
		return "", e.Wrap(err, "init clipboard")
	}
	text := clipboard.Read(clipboard.FmtText)
	return string(bytes.TrimSpace(text)), nil
}

type bookmark struct {
	content     string
	description string
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
	f, err := os.OpenFile(path,
		os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return e.Wrap(err, "open bookmark file")
	}
	defer f.Close()

	line := b.content
	if b.description != "" {
		line += " # " + b.description
	}
	bytes, err := ioutil.ReadAll(f)
	if len(bytes) > 0 && bytes[len(bytes)-1] != '\n' {
		line = "\n" + line
	}
	_, err = f.WriteString(line)
	return e.Wrap(err, "write bookmark file")
}

func addBookmarkIfNotExists(c cfg) (bookmark, error) {
	content, err := getBookmarkContent()
	if err != nil {
		return bookmark{}, e.Wrap(err, "get bookmark content")
	}
	return addBookmarkContentIfNotExists(content, c)
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
	bookmark := bookmark{
		content:     content,
		description: description,
	}
	err = addBookmarkToFile(bookmark, c.bookmarkFile)
	if err != nil {
		return bookmark, e.Wrap(err, "add bookmark to file")
	}
	return bookmark, nil
}

func defaultPickLine(lines []string) (string, error) {
	height := util.Min(len(lines)*35, 600)
	width := util.Min(
		util.FoldLeft1(util.Max[int], util.Map(util.Len, lines))*10,
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

func bookmarkFromLine(line string) (bookmark, error) {
	bookmark := bookmark{}
	split := strings.Split(line, " # ")
	if len(split) > 1 {
		bookmark.description = split[1]
	}
	bookmark.content = strings.TrimSpace(split[0])
	return bookmark, nil
}

func pickBookmark(pickerCmd []string) (bookmark, error) {
	fileContent, err := os.ReadFile(BookmarkFile)
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
	b, err := bookmarkFromLine(chosenLine)
	if err != nil {
		return bookmark{}, e.Wrap(err, "parse bookmark line")
	}
	return b, nil
}

func outputBookmark(b bookmark) error {
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
	done := clipboard.Write(clipboard.FmtText, []byte(b.content))
	<-done
	return nil
}

func notifyBookmarkAdded(b bookmark, duration time.Duration) error {
	return util.Notify(
		fmt.Sprintf("New bookmark\n`%s`", b.content),
		util.Low, uint(duration.Milliseconds()), "",
	)
}

type cfg struct {
	askForDescription bool
	bookmarkFile      string
	pickerCmd         []string
	notify            bool
	notifyDuration    time.Duration
}

func getConfig(x *Z.Cmd) (cfg, error) {
	bookmarkFile, err := util.Get(x, `bookmarkFile`)
	if err != nil {
		return cfg{}, err
	}
	pickerCmd, err := util.GetCommand(x, `pickerCmd`)
	if err != nil {
		return cfg{}, err
	}
	askForDescription, err := util.GetBool(x, `askForDescription`)
	if err != nil {
		return cfg{}, err
	}
	notify, err := util.GetBool(x, `notify`)
	if err != nil {
		return cfg{}, err
	}
	notifyDuration, err := util.GetDuration(x, `notifyDuration`)
	if err != nil {
		return cfg{}, err
	}
	return cfg{
		askForDescription: askForDescription,
		bookmarkFile:      bookmarkFile,
		pickerCmd:         pickerCmd,
		notify:            notify,
		notifyDuration:    notifyDuration,
	}, nil
}

func cmd(x *Z.Cmd) error {
	c, err := getConfig(x)
	if err != nil {
		return e.Wrap(err, "get config")
	}
	b, err := pickBookmark(c.pickerCmd)
	if err != nil {
		return e.Wrap(err, "pick bookmark")
	}
	return outputBookmark(b)
}

func add(x *Z.Cmd, args ...string) error {
	c, err := getConfig(x.Caller)
	if err != nil {
		return e.Wrap(err, "get config")
	}
	var b bookmark
	if len(args) == 0 {
		b, err = addBookmarkIfNotExists(c)
		util.Must(err)
		return nil
	}
	content := args[0]
	b, err = addBookmarkContentIfNotExists(content, c)
	util.Must(err)
	err = notifyBookmarkAdded(b, c.notifyDuration)
	util.Must(err)
	return nil
}

var Cmd = &Z.Cmd{
	Name:    `bookmark`,
	Summary: `Manage bookmarks`,
	Commands: []*Z.Cmd{
		help.Cmd, vars.Cmd, conf.Cmd,
		initCmd,
		addCmd,
	},
	Call: func(x *Z.Cmd, args ...string) error {
		defer util.TrapPanic()
		util.Must(cmd(x))
		return nil
	},
	Shortcuts: Z.ArgMap{
		`askForDescription`: {`var`, `set`, `askForDescription`},
		`bookmarkFile`:      {`var`, `set`, `bookmarkFile`},
		`pickerCmd`:         {`var`, `set`, `pickerCmd`},
		`notify`:            {`var`, `set`, `notify`},
		`notifyDuration`:    {`var`, `set`, `notifyDuration`},
	},
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

var initCmd = &Z.Cmd{
	Name:     `init`,
	Summary:  `sets all values to defaults`,
	Commands: []*Z.Cmd{help.Cmd},

	Description: `
		The {{cmd .Name}} command sets all cached variables to their initial
		values. Any variable name from {{cmd "conf"}} will be used to
		initialize if defined.  Otherwise, the following hard-coded package
		globals will be used instead:

            askForDescription - {{dAskForDescription}}
            bookmarkFile - {{dBookmarkFile}}
            pickerCmd - {{dPickerCmd}}
            notify - {{dNotify}}
            notifyDuration - {{dNotifyDuration}}
	`,
	Call: func(x *Z.Cmd, _ ...string) error {
		defs := map[string]string{
			`askForDescription`: AskForDescription,
			`bookmarkFile`:      BookmarkFile,
			`pickerCmd`:         PickerCmd,
			`notify`:            Notify,
			`notifyDuration`:    NotifyDuration,
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
