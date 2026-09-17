package user

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"

	"cinema-booking/internal/dto"
	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/middleware"
	"cinema-booking/internal/services"
	"cinema-booking/internal/utils"
)

type AuthHandler struct {
	authService services.UserAuthService
}

func NewAuthHandler(authService services.UserAuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// @Summary Register a new user
// @Description Creates a user account with role `user` and returns the user together with an access token.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration payload"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} utils.ValidationError "validation failed or dateOfBirth must be in the past"
// @Failure 409 {object} utils.APIErrorResponse "email already exists"
// @Failure 500 {object} utils.APIErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *echo.Context) error {
	var request dto.RegisterRequest
	if err := utils.BindAndValidate(c, &request); err != nil {
		return err
	}

	user, token, err := h.authService.Register(c.Request().Context(), services.RegisterUserInput{
		Email:       request.Email,
		Password:    request.Password,
		FullName:    request.FullName,
		Phone:       request.Phone,
		DateOfBirth: request.DateOfBirth,
		AvatarURL:   request.AvatarURL,
	})
	if err != nil {
		return authError(err)
	}

	return c.JSON(http.StatusCreated, dto.AuthResponse{
		User:   dto.NewUserResponse(user),
		Tokens: dto.NewTokenResponse(token),
	})
}

// @Summary Log in
// @Description Verifies email and password and returns an access token.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Credentials"
// @Success 200 {object} dto.TokenResponse
// @Failure 400 {object} utils.ValidationError
// @Failure 401 {object} utils.APIErrorResponse "invalid email or password"
// @Failure 500 {object} utils.APIErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *echo.Context) error {
	var request dto.LoginRequest
	if err := utils.BindAndValidate(c, &request); err != nil {
		return err
	}

	token, err := h.authService.Login(c.Request().Context(), request.Email, request.Password)
	if err != nil {
		return authError(err)
	}

	return c.JSON(http.StatusOK, dto.NewTokenResponse(token))
}

// @Summary Log out
// @Description Revokes the current access token; it is rejected by every protected endpoint until it expires.
// @Tags auth
// @Produce json
// @Success 200 {object} dto.MessageResponse
// @Failure 401 {object} utils.APIErrorResponse
// @Failure 500 {object} utils.APIErrorResponse
// @Security bearerauth
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *echo.Context) error {
	if err := h.authService.Logout(c.Request().Context(), middleware.CurrentClaims(c)); err != nil {
		return utils.ServiceError(err)
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "ok"})
}

// @Summary Get the current user
// @Tags auth
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} utils.APIErrorResponse
// @Security bearerauth
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *echo.Context) error {
	user := middleware.CurrentUser(c)
	if user == nil {
		return utils.APIError(http.StatusUnauthorized, "access token required")
	}
	return c.JSON(http.StatusOK, dto.NewUserResponse(user))
}

func authError(err error) error {
	switch {
	case errors.Is(err, apperrors.ErrEmailTaken):
		return utils.APIError(http.StatusConflict, err.Error())
	case errors.Is(err, apperrors.ErrDateOfBirthInFuture):
		return utils.APIError(http.StatusBadRequest, err.Error())
	case errors.Is(err, apperrors.ErrInvalidCredentials):
		return utils.APIError(http.StatusUnauthorized, err.Error())
	default:
		return utils.ServiceError(err)
	}
}
