package middleware

import (
	"errors"
	"net/http"

	echojwt "github.com/labstack/echo-jwt/v5"
	"github.com/labstack/echo/v5"
	"gorm.io/gorm"

	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
	"cinema-booking/internal/user_auth"
	"cinema-booking/internal/utils"
)

const (
	UserClaimsContextKey  = "user_claims"
	CurrentUserContextKey = "current_user"
)

func RequireUser(tokens *userauth.TokenManager, userTokens repositories.UserTokenRepository, users repositories.UserRepository) echo.MiddlewareFunc {
	verifyJWT := echojwt.WithConfig(echojwt.Config{
		ContextKey: UserClaimsContextKey,
		ParseTokenFunc: func(_ *echo.Context, raw string) (any, error) {
			return tokens.Parse(raw)
		},
		ErrorHandler: JWTErrorHandler,
	})

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return verifyJWT(func(c *echo.Context) error {
			claims := CurrentClaims(c)
			if claims == nil {
				return utils.APIError(http.StatusUnauthorized, "invalid access token")
			}
			ctx := c.Request().Context()

			revoked, err := userTokens.IsAccessRevoked(ctx, claims.ID)
			if err != nil {
				return utils.APIErrorFrom(http.StatusInternalServerError, "internal server error", err)
			}
			if revoked {
				return utils.APIError(http.StatusUnauthorized, apperrors.ErrAccessTokenRevoked.Error())
			}

			userID, err := claims.UserID()
			if err != nil {
				return utils.APIError(http.StatusUnauthorized, "invalid access token")
			}
			user, err := users.FindByID(ctx, userID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return utils.APIError(http.StatusUnauthorized, "user not found")
			}
			if err != nil {
				return utils.APIErrorFrom(http.StatusInternalServerError, "internal server error", err)
			}

			c.Set(CurrentUserContextKey, user)
			return next(c)
		})
	}
}

func JWTErrorHandler(_ *echo.Context, err error) error {
	switch {
	case errors.Is(err, echojwt.ErrJWTMissing):
		return utils.APIErrorFrom(http.StatusUnauthorized, "access token is required", err)
	case errors.Is(err, echojwt.ErrJWTInvalid):
		return utils.APIErrorFrom(http.StatusUnauthorized, "invalid or expired access token", err)
	default:
		return utils.APIErrorFrom(http.StatusUnauthorized, "invalid access token", err)
	}
}

func CurrentClaims(c *echo.Context) *userauth.Claims {
	claims, _ := c.Get(UserClaimsContextKey).(*userauth.Claims)
	return claims
}

func CurrentUser(c *echo.Context) *models.User {
	user, _ := c.Get(CurrentUserContextKey).(*models.User)
	return user
}
