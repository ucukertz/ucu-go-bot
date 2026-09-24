package main

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/provider"
	"github.com/zendev-sh/goai/provider/google"
	"go.mau.fi/whatsmeow/types/events"
)

const (
	CHAT_HISTORY_MAX_LEN = 50
	CHAT_MAX_ATTEMPT     = 10
	GEMINI_MODEL         = "gemini-2.5-flash"
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

func ChatGaiObjOnce[T any](msg *events.Message, query string, hint string) (T, error) {
	var zero T
	var err error
	model := google.Chat(GEMINI_MODEL, google.WithAPIKey(ENV_TOKEN_GEMINI))

	for attempt := 1; attempt <= CHAT_MAX_ATTEMPT; attempt++ {
		result, gerr := goai.GenerateObject[T](context.Background(), model,
			goai.WithSystem(hint),
			goai.WithPrompt(query),
		)
		if err = gerr; err == nil {
			return result.Object, nil
		}

		log.Trace().Int("attempt", attempt).Err(err).Msg("ChatGaiObjOnce RETRY")
		if attempt == 1 && msg != nil {
			WaReact(msg, "⏳")
		}
		if attempt < CHAT_MAX_ATTEMPT {
			AdminBackoff(attempt)
		}
	}

	return zero, err
}

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

	qryMedia, _ := WaMsgMediaOrQuoted(msg)

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

func ChatCmdChk(msg *events.Message, cmd string) bool {
	switch cmd {
	case AdminDevDiff("!xai", "!ai"):
		go ChatGaiConvo(msg)
		return true
	}

	return false
}
