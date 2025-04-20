package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/samber/lo"
	"go.mau.fi/whatsmeow/types/events"
)

var (
	SDAPI_TIMEOUT     = 50 * time.Second
	SDAPI_MAX_ATTEMPT = 10
)

type SdCkpt struct {
	name      string
	group     string
	sampler   string
	n_sample  uint8
	cfg_scale float32
	prompt    SdPrompt
}

type SdPrompt struct {
	prepos  string
	postpos string
	preneg  string
	postneg string
}

func (ckpt *SdCkpt) Create(name string, group string) *SdCkpt {
	ckpt.name = name
	ckpt.group = lo.If(len(group) > 0, group).Else("xl-booru")
	return ckpt
}

func (ckpt *SdCkpt) Sampling(sampler string, n_sample uint8, cfg_scale float32) *SdCkpt {
	ckpt.sampler = lo.If(len(sampler) > 0, sampler).Else("Euler a")
	ckpt.n_sample = lo.If(n_sample != 0, n_sample).Else(32)
	ckpt.cfg_scale = lo.If(cfg_scale != 0, cfg_scale).Else(7.0)
	return ckpt
}

func (ckpt *SdCkpt) AddPrompt(prepos string, postpos string, preneg string, postneg string) *SdCkpt {
	ckpt.prompt.prepos = fmt.Sprint(ckpt.prompt.prepos, prepos)
	ckpt.prompt.postpos = fmt.Sprint(ckpt.prompt.postpos, postpos)
	ckpt.prompt.preneg = fmt.Sprint(ckpt.prompt.preneg, preneg)
	ckpt.prompt.postneg = fmt.Sprint(ckpt.prompt.postneg, postneg)
	return ckpt
}

var xbpostpos_default = "BREAK vibrant_colors, colorful, masterpiece, best quality, amazing quality, very aesthetic, absurdres, newest, "
var xbpostneg_default = fmt.Sprintln("chibi, bald, bad anatomy, poorly drawn, deformed anatomy, deformed fingers, censored, mosaic_censoring, bar_censor, shota, white_pupils, empty_eyes, multicolored_hair,",
	"BREAK lowres, (bad quality, worst quality:1.2), sketch, jpeg artifacts, censor, blurry, watermark, ")

var SdCkpts = map[string]SdCkpt{
	"!i.wai": *new(SdCkpt).Create("wai", "").Sampling("", 0, 0).AddPrompt("", xbpostpos_default, "", xbpostneg_default),
	"!i.mei": *new(SdCkpt).Create("mei", "").Sampling("", 0, 0).AddPrompt("", xbpostpos_default, "", xbpostneg_default),
	"!i.fwa": *new(SdCkpt).Create("fuwa", "").Sampling("", 0, 0).AddPrompt("", xbpostpos_default, "", xbpostneg_default),
	"!i.fwt": *new(SdCkpt).Create("fuwa", "").Sampling("", 16, 2).AddPrompt("", xbpostpos_default+SdTurbo, "", xbpostneg_default),
}

func SdApi(msg *events.Message, cmd string) {
	bluff := lo.ValueOr(SdBluff, msg.Info.Sender.User, false)
	if bluff {
		SdBluff[msg.Info.Sender.User] = false
		time.Sleep(30 * time.Second)
		WaReact(msg, "⏳")
		sec := rand.Intn(10) + 10
		time.Sleep(time.Duration(sec) * time.Second)
		img, err := os.ReadFile("assets/bluff.jpg")
		if err != nil {
			WaSaad(msg, err)
			return
		}
		WaImage(msg, img, "")
		return
	}

	ckpt, ok := SdCkpts[cmd]
	if !ok {
		WaSaadStr(msg, "SD CKPT NOT FOUND")
		return
	}

	log.Info().Str("ckpt", ckpt.name).Msg("SD START")
	t_start := time.Now()
	defer func() {
		log.Info().Str("took", fmt.Sprintf("%s", time.Since(t_start).Round(time.Second))).Msg("SD END")
	}()

	prompt := new(SdPrompt)
	prompt.prepos = WaMsgQry(msg)
	prompt.prepos = strings.ReplaceAll(prompt.prepos, "\n", " ")
	prompt.prepos = strings.ToLower(prompt.prepos)

	// Handle tuning
	for _, chara := range lo.Keys(SdChars) {
		if strings.Contains(prompt.prepos, chara) {
			log.Info().Str("chara", chara).Msg("SD TUNE")
			prompt.prepos = strings.ReplaceAll(prompt.prepos, chara, "")

			prompt.prepos = lo.If(strings.Contains(prompt.prepos, "cosplay"), strings.ReplaceAll(prompt.prepos, "cosplay", "")).
				Else(fmt.Sprint(SdChars[chara].clothes, ", ", prompt.prepos))
			prompt.prepos = fmt.Sprint(SdChars[chara].traits, ", ", prompt.prepos)

			prompt.postpos = fmt.Sprint(prompt.postpos, ", ", SdChars[chara].postpos)
			prompt.postneg = fmt.Sprint(prompt.postneg, ", ", SdChars[chara].postneg)
		}
	}

	// Negative prompt
	if strings.Contains(prompt.prepos, "nega") {
		split := strings.Split(prompt.prepos, "nega")
		prompt.prepos = split[0] + ", "
		prompt.preneg = split[1] + ", "
	} else {
		prompt.prepos = prompt.prepos + ", "
	}

	// Handle prompt
	pos := strings.Join([]string{ckpt.prompt.prepos, prompt.prepos, prompt.postpos, ckpt.prompt.postpos}, " ")
	neg := strings.Join([]string{ckpt.prompt.preneg, prompt.preneg, prompt.postneg, ckpt.prompt.postneg}, " ")

	for range 3 {
		pos = strings.ReplaceAll(pos, "\n", ", ")
		neg = strings.ReplaceAll(neg, "\n", ", ")
		pos = strings.ReplaceAll(pos, "  ", " ")
		neg = strings.ReplaceAll(neg, "  ", " ")
		pos = strings.ReplaceAll(pos, " , ", ", ")
		neg = strings.ReplaceAll(neg, " , ", ", ")
		pos = strings.ReplaceAll(pos, ", , ,", ",")
		neg = strings.ReplaceAll(neg, ", , ,", ",")
		pos = strings.ReplaceAll(pos, ", ,", ",")
		neg = strings.ReplaceAll(neg, ", ,", ",")
		pos = strings.ReplaceAll(pos, ",,", ",")
		neg = strings.ReplaceAll(neg, ",,", ",")
		pos = strings.TrimLeft(pos, ", ")
		neg = strings.TrimLeft(neg, ", ")
		pos = strings.TrimRight(pos, ", ")
		neg = strings.TrimRight(neg, ", ")
	}

	attempt := 0
	SdHttpc := HttpcBase.SetBaseURL(ENV_BASEURL_SDAPI).SetBasicAuth(ENV_BAUTH_SDAPI_USER, ENV_BAUTH_SDAPI_PASS).SetTimeout(SDAPI_TIMEOUT)

	// Check server readiness
	for {
		if attempt > SDAPI_MAX_ATTEMPT {
			WaSaadStr(msg, "SD DED")
			return
		}
		r, err := SdHttpc.R().Get("/sdapi/v1/sd-models")
		if err != nil {
			attempt++
		} else if r.StatusCode() != http.StatusOK {
			attempt++
		} else if r.StatusCode() == http.StatusOK {
			break
		}
	}

	WaReact(msg, "⏳")
	log.Info().Msg("[SD POSITIVE] " + pos)
	log.Info().Msg("[SD NEGATIVE] " + neg)

	// Start generation
	reso := lo.ValueOr(SdActiveReso, msg.Info.Sender.User, SdResos["sq"])
	for {
		if attempt > SDAPI_MAX_ATTEMPT {
			WaSaadStr(msg, "SD CANNOT REAL GEN")
			return
		}
		body := map[string]any{
			"prompt":            pos,
			"negative_prompt":   neg,
			"sampler_name":      ckpt.sampler,
			"steps":             ckpt.n_sample,
			"cfg_scale":         ckpt.cfg_scale,
			"width":             reso.Width,
			"height":            reso.Height,
			"override_settings": map[string]any{"sd_model_checkpoint": ckpt.name, "CLIP_stop_at_last_layers": 2},
		}
		r, err := SdHttpc.R().SetBody(body).Post("/sdapi/v1/txt2img")
		if err != nil {
			attempt++
		} else if r.StatusCode() != http.StatusOK {
			attempt++
		} else if r.StatusCode() == http.StatusOK {
			type SdImages struct {
				Images []string `json:"images"`
			}

			var res SdImages
			err = json.Unmarshal(r.Body(), &res)
			if err != nil {
				WaSaad(msg, err)
			}

			if len(res.Images) > 0 {
				image, err := base64.StdEncoding.DecodeString(res.Images[0])
				if err != nil {
					WaSaadStr(msg, "SD INVALID IMG")
					return
				}
				WaImage(msg, image, fmt.Sprintf("%s", time.Since(t_start).Round(time.Second)))
			}
			return
		}
	}
}

type SdReso struct {
	name   string
	Width  int
	Height int
}

var SdBluff = map[string]bool{}
var SdActiveReso = map[string]SdReso{}

var SdResos = map[string]SdReso{
	"sq": {name: "1024x1024", Width: 1024, Height: 1024},
	"w1": {name: "1152x896", Width: 1152, Height: 896},
	"h1": {name: "896x1152", Width: 896, Height: 1152},
	"w2": {name: "1216x832", Width: 1216, Height: 832},
	"h2": {name: "832x1216", Width: 832, Height: 1216},
	"w3": {name: "1344x768", Width: 1344, Height: 768},
	"h3": {name: "768x1344", Width: 768, Height: 1344},
}

func SdCmdChk(msg *events.Message, cmd string) bool {
	if _, ok := SdCkpts[cmd]; ok {
		go SdApi(msg, cmd)
		return true
	}

	switch cmd {
	case "!i.reso":
		if reso, ok := SdResos[WaMsgQry(msg)]; ok {
			SdActiveReso[msg.Info.Sender.User] = reso
			WaText(msg, fmt.Sprintf("A1111 resolution set to %s for you 🫶", reso.name))
		} else {
			WaText(msg, "Resolution not found. Choices: \nsq, h1, h2, h3, w1, w2, w3\n\nExample: `!i.reso sq`")
		}
		return true
	case "!i.bluff":
		SdBluff[msg.Info.Sender.User] = true
		WaReact(msg, "😏")
		return true
	case "!s.tune":
		go SdTune(msg)
		return true
	case "!s.bake":
		go SdBake(msg)
		return true
	case "!s.take":
		go SdTake(msg)
		return true
	}

	return false
}
