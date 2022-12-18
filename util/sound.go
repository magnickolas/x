package util

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	e "github.com/pkg/errors"
)

func PlaySoundBlock(data []byte) error {
	streamer, format, err := mp3.Decode(ioutil.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return e.Wrap(err, "decode mp3")
	}
	defer streamer.Close()

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		return e.Wrap(err, "init speaker")
	}

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
	return nil
}
