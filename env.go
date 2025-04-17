package main

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	ENV_LOADED = ""

	ENV_BASEURL_SDAPI       = ""
	ENV_BASEURL_HUGGINGFACE = ""
	ENV_BASEURL_MEMEGEN     = ""

	ENV_TOKEN_HUGGING = ""
	ENV_TOKEN_GEMINI  = ""

	ENV_BAUTH_SDAPI_USER = ""
	ENV_BAUTH_SDAPI_PASS = ""
)

func EnvLoad() {
	ENV_LOADED = os.Getenv("ENV_LOADED")

	ENV_BASEURL_SDAPI = os.Getenv("BASEURL_SDAPI")
	ENV_BASEURL_HUGGINGFACE = os.Getenv("BASEURL_HUGGINGFACE")
	ENV_BASEURL_MEMEGEN = os.Getenv("BASEURL_MEMEGEN")

	ENV_TOKEN_HUGGING = os.Getenv("TOKEN_HUGGING")
	ENV_TOKEN_GEMINI = os.Getenv("TOKEN_GEMINI")

	ENV_BAUTH_SDAPI_USER = os.Getenv("BAUTH_SDAPI_USER")
	ENV_BAUTH_SDAPI_PASS = os.Getenv("BAUTH_SDAPI_PASS")
}

func EnvInit() {
	EnvLoad()
	if ENV_LOADED != "" {
		log.Info().Msg("ENV already loaded")
		return
	}

	log.Info().Msg("Loading .env file")
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	EnvLoad()
	if ENV_LOADED == "" {
		panic("NO ENV LOADED")
	}
}
