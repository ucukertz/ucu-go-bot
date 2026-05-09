package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/http"
	"strings"
	"time"

	"github.com/nfnt/resize"
	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/provider"
	"github.com/zendev-sh/goai/provider/google"
	"go.mau.fi/whatsmeow/types/events"
)

const (
	CHAT_HISTORY_MAX_LEN = 50
	CHAT_MAX_ATTEMPT     = 10
	GEMINI_MODEL         = "gemini-2.5-flash"

	KONTEXT_TIMEOUT = 300 * time.Second
)

// ChatNew creates a new empty message history
func ChatNew() []provider.Message {
	return []provider.Message{}
}

// ChatReset clears the history and notifies the user
func ChatReset(msg *events.Message) []provider.Message {
	WaReplyText(msg, "No thoughts. Head's empty. 👍")
	return ChatNew()
}

var GaiHistory []provider.Message = nil

func ChatGaiTextOnce(msg *events.Message, query string, hint string) (string, error) {
	var err error
	attempt := 0
	model := google.Chat(GEMINI_MODEL, google.WithAPIKey(ENV_TOKEN_GEMINI))
	searchTool := google.Tools.GoogleSearch()

	for attempt < CHAT_MAX_ATTEMPT {
		attempt++

		result, gerr := goai.GenerateText(context.Background(), model,
			goai.WithSystem(hint),
			goai.WithPrompt(query),
			goai.WithTools(goai.Tool{
				Name:                   searchTool.Name,
				ProviderDefinedType:    searchTool.ProviderDefinedType,
				ProviderDefinedOptions: searchTool.ProviderDefinedOptions,
			}),
			goai.WithMaxSteps(5),
		)
		err = gerr
		if err == nil {
			return result.Text, nil
		}

		log.Trace().Int("attempt", attempt).Err(err).Msg("ChatGaiTextOnce RETRY")
		if attempt == 1 && msg != nil {
			WaReact(msg, "⏳")
		}

		if attempt < CHAT_MAX_ATTEMPT {
			AdminBackoff(attempt)
		}
	}

	if err != nil {
		if msg != nil {
			if strings.Contains(err.Error(), "demand") {
				WaReact(msg, "🤕")
			} else {
				WaSaadStr(msg, "GAI SEND: "+err.Error())
			}
		}
		return "", err
	}

	return "", nil
}

func ChatGaiConvo(msg *events.Message) {
	if GaiHistory == nil || len(GaiHistory) >= CHAT_HISTORY_MAX_LEN || WaMsgPrompt(msg) == "/reset" {
		if GaiHistory != nil && len(GaiHistory) >= CHAT_HISTORY_MAX_LEN {
			WaReact(msg, "😵‍💫")
		}
		if WaMsgPrompt(msg) == "/reset" {
			GaiHistory = ChatReset(msg)
			return
		}
		GaiHistory = ChatNew()
	}

	qryMedia := WaMsgMedia(msg)
	if qryMedia == nil {
		qryMedia = WaMsgMediaQuoted(msg)
	}

	prompt := WaMsgPrompt(msg)
	userMsg := provider.Message{
		Role:    provider.RoleUser,
		Content: []provider.Part{{Type: provider.PartText, Text: prompt}},
	}

	if qryMedia != nil {
		mime := http.DetectContentType(qryMedia)
		userMsg.Content = append([]provider.Part{
			{
				Type:      provider.PartImage,
				URL:       "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(qryMedia),
				MediaType: mime,
			},
		}, userMsg.Content...)
	}

	GaiHistory = append(GaiHistory, userMsg)

	var err error
	var result *goai.TextResult
	attempt := 0
	model := google.Chat(GEMINI_MODEL, google.WithAPIKey(ENV_TOKEN_GEMINI))
	searchTool := google.Tools.GoogleSearch()

	for attempt < CHAT_MAX_ATTEMPT {
		attempt++
		result, err = goai.GenerateText(context.Background(), model,
			goai.WithMessages(GaiHistory...),
			goai.WithTools(goai.Tool{
				Name:                   searchTool.Name,
				ProviderDefinedType:    searchTool.ProviderDefinedType,
				ProviderDefinedOptions: searchTool.ProviderDefinedOptions,
			}),
			goai.WithMaxSteps(5),
		)

		if err == nil {
			break
		}

		log.Trace().Int("attempt", attempt).Err(err).Msg("ChatGaiConvo RETRY")
		if attempt == 1 {
			WaReact(msg, "⏳")
		}

		if attempt < CHAT_MAX_ATTEMPT {
			AdminBackoff(attempt)
		}
	}

	if err != nil {
		if strings.Contains(err.Error(), "demand") {
			WaReact(msg, "🤕")
		} else {
			WaSaadStr(msg, "GAI SEND: "+err.Error())
		}
		return
	}

	GaiHistory = append(GaiHistory, result.ResponseMessages...)

	res := result.Text
	if len(res) > 0 {
		WaReplyText(msg, res)
	}
}

func ChatKontext(msg *events.Message) {
	t_start := time.Now()
	defer func() {
		log.Info().Str("took", fmt.Sprintf("%s", time.Since(t_start).Round(time.Second))).Msg("KONTEXT END")
	}()

	var reqbody struct {
		Prompt string `json:"prompt,omitempty"`
		Img    string `json:"input_image,omitempty"`
	}

	img := WaMsgMedia(msg)
	if img == nil {
		img = WaMsgMediaQuoted(msg)
		if img == nil {
			WaSaadStr(msg, "No image to edit ☹️")
			return
		}
	}

	imgimg, err := PicImgImgFromBytes(img)
	if err != nil {
		log.Error().Err(err).Msg("IMGCONV")
		WaSaadStr(msg, "IMGCONV: "+err.Error())
		return
	}

	imgr := resize.Thumbnail(1344, 1344, imgimg, resize.Lanczos3)
	thumbbuf := new(bytes.Buffer)
	if err := png.Encode(thumbbuf, imgr); err != nil {
		log.Error().Err(err).Msg("PNG ENCODE")
		WaSaadStr(msg, "PNG ENCODE: "+err.Error())
		return
	}

	reqbody.Img = base64.StdEncoding.EncodeToString(thumbbuf.Bytes())
	reqbody.Prompt = WaMsgPrompt(msg)

	r, err := HttpcBase().SetTimeout(KONTEXT_TIMEOUT).SetBasicAuth(ENV_BAUTH_SDAPI_USER, ENV_BAUTH_SDAPI_PASS).
		R().SetBody(reqbody).Post(ENV_BASEURL_KONTEXT)
	if err != nil {
		WaSaadStr(msg, "KONTEXT: "+err.Error())
		return
	}
	if r.StatusCode() != http.StatusOK {
		if r.StatusCode() == http.StatusTooManyRequests {
			WaReact(msg, "💸")
			return
		} else {
			WaSaadStr(msg, "KONTEXT DED "+r.Status())
			return
		}
	}

	image, err := base64.StdEncoding.DecodeString(r.String())
	if err != nil {
		WaSaadStr(msg, "KONTEXT DECODE: "+err.Error())
		return
	}
	t_all := time.Since(t_start).Round(time.Second)
	t_str := fmt.Sprintf("%s", t_all)
	WaReplyImg(msg, image, t_str)
}

func ChatCmdChk(msg *events.Message, cmd string) bool {
	switch cmd {
	case AdminDevDiff("!xai", "!ai"):
		go ChatGaiConvo(msg)
		return true
	case AdminDevDiff("!x.flx", "!m.flx"):
		go ChatKontext(msg)
		return true
	}

	return false
}
