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
	"github.com/samber/lo"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/genai"
)

const (
	CHAT_HISTORY_MAX_LEN = 50
	GEMINI_MODEL         = "gemini-2.5-flash"

	KONTEXT_TIMEOUT = 300 * time.Second
)

// Generic chat reset
func ChatReset[T any](msg *events.Message, history T) T {
	WaText(msg, "No thoughts. Head's empty. 👍")
	return lo.Empty[T]()
}

var GaiChatClient *genai.Client = nil
var GaiChat *genai.Chat = nil

func ChatGaiOneText(query string, hint string) (string, error) {
	gai, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  ENV_TOKEN_GEMINI,
		Backend: genai.BackendGeminiAPI})
	if err != nil {
		return "", err
	}

	config := genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(hint)},
			Role:  "user",
		},
	}
	r, err := gai.Models.GenerateContent(context.Background(), GEMINI_MODEL, genai.Text(query), &config)
	if err != nil {
		return "", err
	}
	res := ""

	for _, part := range r.Candidates[0].Content.Parts {
		res = fmt.Sprint(res, part.Text)
	}

	return res, nil
}

func ChatGaiReset() (*genai.Chat, error) {
	return GaiChatClient.Chats.Create(context.Background(), GEMINI_MODEL, nil, nil)
}

func ChatGaiConvo(msg *events.Message) {
	var err error
	if GaiChatClient == nil {
		GaiChatClient, err = genai.NewClient(context.Background(), &genai.ClientConfig{
			APIKey:  ENV_TOKEN_GEMINI,
			Backend: genai.BackendGeminiAPI,
		})
		if err != nil {
			WaSaadStr(msg, "GAI CLI: "+err.Error())
			return
		}
	}

	if GaiChat == nil || len(GaiChat.History(false)) >= CHAT_HISTORY_MAX_LEN || WaMsgQry(msg) == "/reset" {
		if GaiChat != nil && len(GaiChat.History(false)) >= CHAT_HISTORY_MAX_LEN {
			WaReact(msg, "😵‍💫")
		}
		GaiChat, err = ChatGaiReset()
		if err != nil {
			WaSaadStr(msg, "GAI RESET: "+err.Error())
			return
		}
	}

	qryMedia := WaMsgMedia(msg)
	mime := http.DetectContentType(qryMedia)

	var r *genai.GenerateContentResponse
	if qryMedia != nil {
		r, err = GaiChat.SendMessage(context.Background(),
			*genai.NewPartFromBytes(qryMedia, mime),
			*genai.NewPartFromText(WaMsgQry(msg)),
		)
	} else {
		r, err = GaiChat.SendMessage(context.Background(),
			*genai.NewPartFromText(WaMsgQry(msg)),
		)
	}
	if err != nil {
		WaSaadStr(msg, "GAI SEND: "+err.Error())
		return
	}

	res := ""
	for _, part := range r.Candidates[0].Content.Parts {
		if len(part.Text) > 0 {
			res = fmt.Sprint(res, part.Text)
		}
		if part.InlineData != nil && len(part.InlineData.Data) > 0 {
			if strings.HasPrefix(part.InlineData.MIMEType, "image/") {
				WaImage(msg, part.InlineData.Data, part.InlineData.DisplayName)
			} else {
				WaSaadStr(msg, "GAI MIME IN: "+part.InlineData.MIMEType)
			}
		}
		if part.FileData != nil && len(part.FileData.FileURI) > 0 {
			r, err := HttpcBase.Clone().R().Get(part.FileData.FileURI)
			if err != nil {
				WaSaadStr(msg, "GAI URL FL:"+err.Error())
				return
			}
			if strings.HasPrefix(part.FileData.MIMEType, "image/") {
				WaImage(msg, r.Body(), part.FileData.FileURI)
			} else {
				WaSaadStr(msg, "GAI MIME FL: "+part.FileData.MIMEType)
			}
		}
	}

	if len(res) > 0 {
		WaText(msg, res)
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
		WaSaadStr(msg, "No image to edit ☹️")
		return
	}

	imgimg, err := WaByte2ImgImg(img)
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
	reqbody.Prompt = WaMsgQry(msg)

	r, err := HttpcBase.Clone().SetTimeout(KONTEXT_TIMEOUT).SetBasicAuth(ENV_BAUTH_SDAPI_USER, ENV_BAUTH_SDAPI_PASS).
		R().SetBody(reqbody).Post(ENV_BASEURL_KONTEXT)
	if err != nil {
		WaSaadStr(msg, "KONTEXT: "+err.Error())
		return
	}
	if r.StatusCode() != http.StatusOK {
		if r.StatusCode() == http.StatusTooManyRequests {
			WaText(msg, "MODAL ZERO")
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
	WaImage(msg, image, t_str)
}

func ChatCmdChk(msg *events.Message, cmd string) bool {
	switch cmd {
	case "!ai":
		go ChatGaiConvo(msg)
		return true
	case "!i.flx":
		go ChatKontext(msg)
		return true
	}

	return false
}
