package util

import (
	"errors"
	"fmt"
	"os/exec"

	e "github.com/pkg/errors"
)

type urgency int

const (
	Low urgency = iota
	Normal
	Critical
)

func (u urgency) String() (string, error) {
	switch u {
	case Low:
		return "low", nil
	case Normal:
		return "normal", nil
	case Critical:
		return "critical", nil
	default:
		return "", errors.New("unknown urgency")
	}
}

func Notify(msg string, urgency urgency, timeout uint, iconPath string) error {
	name := "notify-send"
	urgencyS, err := urgency.String()
	if err != nil {
		return e.Wrap(err, "get urgency string")
	}

	args := []string{"-u", urgencyS}
	if timeout > 0 {
		args = append(args, "-t", fmt.Sprint(timeout))
	}
	if iconPath != "" {
		args = append(args, "-i", iconPath)
	}
	args = append(args, msg)
	err = exec.Command(
		name,
		args...,
	).Run()
	return e.Wrap(err, "send notification")
}
