package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/samber/lo"
	"go.mau.fi/whatsmeow/types/events"
)

var (
	HUG_MAX_ATTEMPT = 10
	HUG_RETRY_SEC   = 20 * time.Second
)

type hugmodel struct {
	url     string
	postpos string
}

var HugLgc = map[string]hugmodel{
	"!img":   {"stabilityai/stable-diffusion-xl-base-1.0", ""},
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
	if _, ok := HugLgc[cmd]; !ok {
		return false
	}

	go HuggingLegacy(msg, cmd, WaMsgPrompt(msg))
	return true
}
