package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/samber/lo"
	"github.com/ucukertz/hfs"
	"go.mau.fi/whatsmeow/types/events"
)

var (
	HFS_TIMEOUT = 300 * time.Second

	HUG_MAX_ATTEMPT = 10
	HUG_RETRY_SEC   = 20 * time.Second
)

type hugLegacyModel struct {
	url     string
	postpos string
}

func HugZit(msg *events.Message) {
	s := hfs.NewHfs[any, any]("mrfakename-z-image-turbo").WithBearerToken(ENV_TOKEN_HUGGING).WithTimeout(HFS_TIMEOUT)
	log.Info().Msg("HUG ZIT start")

	query := WaMsgPrompt(msg)
	t_start := time.Now()

	ucfg := GenGet(msg)
	reso := ucfg.Reso
	img, err := s.DoFD("/generate_image", query, reso.Height, reso.Width, 9, 42, true)
	if err != nil {
		WaSaad(msg, err)
		return
	}

	t_all := time.Since(t_start).Round(time.Second)
	dur := fmt.Sprintf("%s", t_all)
	WaReplyImg(msg, img, dur)
}

func HugCmdChk(msg *events.Message, cmd string) bool {
	switch cmd {
	case AdminDevDiff("!x.zit", "!z.zit"), AdminDevDiff("!ximg", "!img"):
		go HugZit(msg)
		return true
	}
	return false
}

var HugLgc = map[string]hugLegacyModel{
	"!l.sxl": {"stabilityai/stable-diffusion-xl-base-1.0", ""},
}

func HugLegacy(msg *events.Message, model string, query string, attempt int) ([]byte, error) {
	r, err := HttpcBase().
		SetBaseURL(ENV_BASEURL_HUGGINGFACE).
		SetAuthToken(ENV_TOKEN_HUGGING).
		SetTimeout(HUG_RETRY_SEC).
		R().SetBody(map[string]any{"inputs": fmt.Sprint(query, ",", HugLgc[model].postpos, lo.RandomString(6, lo.NumbersCharset))}).
		Post(HugLgc[model].url)

	if err != nil || r.StatusCode() != http.StatusOK {
		if attempt == 0 {
			WaReact(msg, "⏳")
		}
		if err != nil {
			return nil, err
		}
		if r.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("HUG %s", r.Status())
		}
	}
	return r.Body(), nil
}

func HuggingLegacy(msg *events.Message, model string, query string) {
	attempt := 0
	img := []byte{}
	err := fmt.Errorf("HUG ERRINIT")
	for attempt < HUG_MAX_ATTEMPT && err != nil {
		img, err = HugLegacy(msg, model, query, 0)
		attempt++
	}

	if err != nil {
		WaSaad(msg, err)
		return
	}
	WaReplyImg(msg, img, "")
}

func HugLegacyCmdChk(msg *events.Message, cmd string) bool {
	if _, ok := HugLgc[cmd]; ok {
		go HuggingLegacy(msg, cmd, WaMsgPrompt(msg))
		return true
	}

	return false
}
