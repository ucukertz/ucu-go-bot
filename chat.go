package main

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/genai"
)

const (
	CHAT_HISTORY_MAX_LEN = 30
)

func ChatReset[T any](msg *events.Message, history T) T {
	WaText(msg, "No thoughts. Head's empty. 👍")
	return lo.Empty[T]()
}

var gai_history = []*genai.Content{}

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
	r, err := gai.Models.GenerateContent(context.Background(), "gemini-2.0-flash", genai.Text(query), &config)
	if err != nil {
		return "", err
	}
	res := ""

	for _, part := range r.Candidates[0].Content.Parts {
		res = fmt.Sprint(res, part.Text)
	}

	return res, nil
}

func GaiText(msg *events.Message) {
	if WaMsgQry(msg) == "/reset" {
		gai_history = ChatReset(msg, gai_history)
		return
	}

	r, err := GaiSingleText(WaMsgQry(msg), "")
	if err != nil {
		WaSaad(msg, err)
		return
	}

	if len(gai_history) > CHAT_HISTORY_MAX_LEN {
		WaReact(msg, "😵")
		gai_history = []*genai.Content{}
	}

	WaText(msg, r)
}

func ChatCmdChk(msg *events.Message, cmd string) bool {
	switch cmd {
	case "!gai":
		go GaiText(msg)
		return true
	}

	return false
}
