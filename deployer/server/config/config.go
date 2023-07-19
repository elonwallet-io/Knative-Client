package config

import (
	"os"
	"strconv"
)

func createConfig() Config {
	sgx_activate, err := strconv.ParseBool(os.Getenv("SGX_ACTIVATE"))
	if err != nil {
		panic(err)
	}
	image_name := os.Getenv("IMAGE_WALLET_SERVICE")
	if sgx_activate {
		image_name = image_name + "-sgx"
	}
	wildcard, err := strconv.ParseBool(os.Getenv("USE_WILDCARD_CERT"))
	if err != nil {
		panic(err)
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
