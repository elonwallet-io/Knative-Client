package main

import (
	"backend/kubernetes_client/server"

	"github.com/rs/zerolog/log"
)

func main() {
	server, err := server.New()
	if err != nil {
		log.Error().Err(err).Caller().Str("Server", "Create").Msg("failed to create server")
		return
	}
	err = server.Run()
	if err != nil {
		log.Error().Err(err).Caller().Str("Server", "Start").Msg("failed to start server")
		return
	}
}
