package auth

import (
	"net/http"

	apikey "github.com/consensys/quorum-key-manager/src/auth/authenticator/api-key"
	"github.com/consensys/quorum-key-manager/src/auth/authenticator/oidc"
	"github.com/consensys/quorum-key-manager/src/infra/log"

	"github.com/consensys/quorum-key-manager/pkg/app"
	"github.com/consensys/quorum-key-manager/src/auth/authenticator"
	authmanager "github.com/consensys/quorum-key-manager/src/auth/policy"
	manifestsmanager "github.com/consensys/quorum-key-manager/src/manifests/manager"
)

func RegisterService(a *app.App, logger log.Logger) error {
	// Load manifests service
	m := new(manifestsmanager.Manager)
	err := a.Service(m)
	if err != nil {
		return err
	}

	// Create and register the stores service
	policyMngr := authmanager.New(*m, logger)
	err = a.RegisterService(policyMngr)
	if err != nil {
		return err
	}

	return nil
}

func Middleware(a *app.App, logger log.Logger) (func(http.Handler) http.Handler, error) {
	// Load configuration
	cfg := new(Config)
	err := a.ServiceConfig(cfg)
	if err != nil {
		return nil, err
	}

	// Load policy manager service
	policyMngr := new(authmanager.Manager)
	err = a.Service(policyMngr)
	if err != nil {
		return nil, err
	}

	auths := []authenticator.Authenticator{}
	if cfg.OIDC != nil {
		oidcAuth, err := oidc.NewAuthenticator(cfg.OIDC)
		if err != nil {
			return nil, err
		} else if oidcAuth != nil {
			logger.Info("OIDC Authenticator is enabled")
			auths = append(auths, oidcAuth)
		}
	}

	if cfg.APIKEY != nil {
		apikeyAuth, err := apikey.NewAuthenticator(cfg.APIKEY)
		if err != nil {
			return nil, err
		} else if apikeyAuth != nil {
			logger.Info("APIKEY Authenticator is enabled")
			auths = append(auths, apikeyAuth)
		}
	}

	// Create middleware
	mid := authenticator.NewMiddleware(
		logger,
		auths...,
	)

	return mid.Then, nil
}