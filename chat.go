package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/samber/lo"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/genai"
)

const (
	CHAT_HISTORY_MAX_LEN = 50
	GEMINI_MODEL         = "gemini-2.5-flash"
)

// Generic chat reset
func ChatReset[T any](msg *events.Message, history T) T {
	WaText(msg, "No thoughts. Head's empty. 👍")
	return lo.Empty[T]()
}

var GaiChatClient *genai.Client = nil
var GaiChat *genai.Chat = nil

func GaiSingleText(query string, hint string) (string, error) {
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

func GaiChatReset() (*genai.Chat, error) {
	return GaiChatClient.Chats.Create(context.Background(), GEMINI_MODEL, nil, nil)
}

func GaiConvo(msg *events.Message) {
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
		GaiChat, err = GaiChatReset()
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
			r, err := HttpcBase.R().Get(part.FileData.FileURI)
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

func ChatCmdChk(msg *events.Message, cmd string) bool {
	switch cmd {
	case "!ai":
		go GaiConvo(msg)
		return true
	case "!flx":
		WaText(msg, "Flux Kontext is a work in progress. It is not yet ready for use.")
		return true
	}

	return false
}
