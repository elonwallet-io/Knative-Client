package config

import (
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

func CreateConfig() Config {
	sgx_activate, err := strconv.ParseBool(os.Getenv("SGX_ACTIVATE"))
	if err != nil {
		log.Error().Err(err).Caller().Str("Parse", "Env").Msg(err.Error())
		sgx_activate = true
	}
	image_name := os.Getenv("IMAGE_WALLET_SERVICE")
	wildcard, err := strconv.ParseBool(os.Getenv("USE_WILDCARD_CERT"))
	if err != nil {
		log.Error().Err(err).Caller().Str("Parse", "Env").Msg(err.Error())
		wildcard = false
	}
	return Config{
		Images: Images{
			WALLET_SERVICE_IMAGE: image_name,
			VAULT_IMAGE:          os.Getenv("IMAGE_VAULT"),
			PCCS_IMAGE:           os.Getenv("IMAGE_PCCS"),
		},
		FRONTEND_URL:  os.Getenv("FRONTEND_URL"),
		WILDCARD:      wildcard,
		BACKEND_URL:   os.Getenv("BACKEND_URL"),
		FRONTEND_HOST: os.Getenv("FRONTEND_HOST"),
		SGX_ACTIVATE:  sgx_activate,
	}
}
