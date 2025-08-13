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

var hugs = map[string]hugmodel{
	"!img":    {"black-forest-labs/FLUX.1-dev", ""},
	"!i.sxl":  {"stabilityai/stable-diffusion-xl-base-1.0", ""},
	"!i.std":  {"stabilityai/stable-diffusion-2-1", ""},
	"!i.some": {"NoCrypt/SomethingV2_2", "masterpiece, best quality, ultra-detailed"},
	"!i.cntr": {"gsdf/Counterfeit-V2.5", "masterpiece, best quality, ultra-detailed"},
	"!i.modi": {"nitrosocke/mo-di-diffusion", "modern disney style"},
	"!i.prot": {"darkstorm2150/Protogen_x3.4_Official_Release", "modelshoot style, analog style, mdjrny-v4 style"},
	"!i.pix":  {"nerijs/pixel-art-xl", ""},
	"!i.logo": {"artificialguybr/LogoRedmond-LogoLoraForSDXL-V2", "LogoRedmAF"},
	"!i.mid":  {"prompthero/openjourney", "mdjrny-v4 style"},
}

func Hug(msg *events.Message, model string, query string, attempt int) ([]byte, error) {
	r, err := HttpcBase.Clone().
		SetBaseURL(ENV_BASEURL_HUGGINGFACE).
		SetAuthToken(ENV_TOKEN_HUGGING).
		SetTimeout(HUG_RETRY_SEC).
		R().SetBody(map[string]any{"inputs": fmt.Sprint(query, ",", hugs[model].postpos, lo.RandomString(6, lo.NumbersCharset))}).
		Post(hugs[model].url)

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

func Hugging(msg *events.Message, model string, query string) {
	attempt := 0
	img := []byte{}
	err := fmt.Errorf("HUG ERRINIT")
	for attempt < HUG_MAX_ATTEMPT && err != nil {
		img, err = Hug(msg, model, query, 0)
		attempt++
	}

	if err != nil {
		WaSaad(msg, err)
		return
	}
	WaImage(msg, img, "")
}

func HugCmdChk(msg *events.Message, cmd string) bool {
	if _, ok := hugs[cmd]; !ok {
		return false
	}

	go Hugging(msg, cmd, WaMsgQry(msg))
	return true
}
