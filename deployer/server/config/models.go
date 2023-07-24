package config

type Images struct {
	WALLET_SERVICE_IMAGE string
	VAULT_IMAGE          string
	PCCS_IMAGE           string
}

type Config struct {
	Images        Images
	FRONTEND_URL  string
	FRONTEND_HOST string
	WILDCARD      bool
	BACKEND_URL   string
	SGX_ACTIVATE  bool
}
