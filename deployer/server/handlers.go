package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func (s *Server) deployment(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return err
	}
	if u.Name != "" {
		route := s.Clients.ReturnRouteIfServiceExists(c.Request().Context(), u.Name)
		if route == "" {
			var err error
			route, err = s.Clients.DeployServerlessWalletService(c.Request().Context(), s.Config.SGX_ACTIVATE, u.Name, s.Config.FRONTEND_URL, s.Config.FRONTEND_HOST, s.Config.BACKEND_URL, s.Config.Images.WALLET_SERVICE_IMAGE)
			if err != nil {
				log.Error().Err(err).Caller().Str("Error", err.Error()).Msg("failed to deploy the Wallet Service")
				return c.JSON(http.StatusInternalServerError, err.Error())
			}
		}
		if s.Config.WILDCARD {
			route = "https" + route[4:]
		}
		return c.JSON(http.StatusOK, Url{Url: route})
	} else {
		log.Debug().Caller().Str("Empty", "username").Msg("username can't be empty string")
		return c.String(http.StatusInternalServerError, "username can't be empty string")
	}
}

func (s *Server) deletion(c echo.Context) error {
	username := c.Param("id")
	route := s.Clients.ReturnRouteIfServiceExists(c.Request().Context(), username)
	if route != "" {
		errors := s.Clients.DeleteServiceForUser(c.Request().Context(), username)
		if len(errors) > 0 {
			log.Error().Err(errors[0]).Caller().Str("Error", errors[0].Error()).Msg("failed to delete service for user " + username)
			return c.String(http.StatusInternalServerError, errors[0].Error())
		}
	}
	return c.String(http.StatusOK, "")
}
