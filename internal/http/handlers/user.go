package handlers

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/http/middleware"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/http/response"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/users"
)

// UserService defines the required user domain interactions.
type UserService interface {
	Register(ctx context.Context, req users.RegisterRequest) (*users.RegisterResult, error)
	Authenticate(ctx context.Context, req users.AuthenticateRequest) (*users.AuthenticateResult, error)
	GetProfile(ctx context.Context, userID string) (*users.Profile, error)
	UpdateProfile(ctx context.Context, userID string, req users.UpdateProfileRequest) (*users.Profile, error)
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	AssignRole(ctx context.Context, userID, role string) error
	Permissions(ctx context.Context, userID string) ([]string, error)
	HasPermission(ctx context.Context, userID, permission string) (bool, error)
}

// UserHandler exposes HTTP handlers for user operations.
type UserHandler struct {
	svc UserService
}

// NewUserHandler constructs the handler.
func NewUserHandler(svc UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// RegisterUserRoutes binds the routes to the application.
func RegisterUserRoutes(app fiber.Router, handler *UserHandler, auth fiber.Handler) {
	usersGroup := app.Group("/users")
	usersGroup.Post("/register", handler.register)
	usersGroup.Post("/login", handler.login)

	authenticated := usersGroup.Group("")
	authenticated.Use(auth)
	authenticated.Get("/me", handler.profile)
	authenticated.Patch("/me", handler.updateProfile)
	authenticated.Post("/me/change-password", handler.changePassword)

	admin := app.Group("/admin")
	admin.Use(auth)
	admin.Post("/users/:id/roles", handler.assignRole)
	admin.Get("/users/:id/permissions", handler.permissions)
}

func (h *UserHandler) register(c *fiber.Ctx) error {
	var req registerRequest
	if err := parseJSON(c, &req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	result, err := h.svc.Register(c.Context(), users.RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Accepted(c, "registration accepted", map[string]any{
		"userId": result.UserID,
		"tokens": tokenPair(result.Tokens),
	})
}

func (h *UserHandler) login(c *fiber.Ctx) error {
	var req loginRequest
	if err := parseJSON(c, &req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	res, err := h.svc.Authenticate(c.Context(), users.AuthenticateRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, users.ErrInvalidCredentials) {
			return response.Unauthorized(c, "invalid credentials")
		}
		if errors.Is(err, users.ErrUserDisabled) {
			return response.Forbidden(c, "user disabled")
		}
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, "login successful", map[string]any{
		"userId": res.UserID,
		"tokens": tokenPair(res.Tokens),
	})
}

func (h *UserHandler) profile(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	prof, err := h.svc.GetProfile(c.Context(), userID)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return response.NotFound(c, "user not found")
		}
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, "profile retrieved", profilePayload(prof))
}

func (h *UserHandler) updateProfile(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	var req updateProfileRequest
	if err := parseJSON(c, &req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	prof, err := h.svc.UpdateProfile(c.Context(), userID, users.UpdateProfileRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return response.NotFound(c, "user not found")
		}
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, "profile updated", profilePayload(prof))
}

func (h *UserHandler) changePassword(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	var req changePasswordRequest
	if err := parseJSON(c, &req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.svc.ChangePassword(c.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, users.ErrInvalidCredentials) {
			return response.Unauthorized(c, "current password incorrect")
		}
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, "password changed", nil)
}

func (h *UserHandler) assignRole(c *fiber.Ctx) error {
	actor := middleware.UserID(c)
	has, err := h.svc.HasPermission(c.Context(), actor, "roles:assign")
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	if !has {
		return response.Forbidden(c, "insufficient permissions")
	}

	target := c.Params("id")
	var req assignRoleRequest
	if err := parseJSON(c, &req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.svc.AssignRole(c.Context(), target, req.Role); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, "role assigned", fiber.Map{"userId": target, "role": req.Role})
}

func (h *UserHandler) permissions(c *fiber.Ctx) error {
	actor := middleware.UserID(c)
	has, err := h.svc.HasPermission(c.Context(), actor, "roles:view")
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	if !has {
		return response.Forbidden(c, "insufficient permissions")
	}

	target := c.Params("id")
	perms, err := h.svc.Permissions(c.Context(), target)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, "permissions retrieved", fiber.Map{"userId": target, "permissions": perms})
}

func parseJSON(c *fiber.Ctx, out any) error {
	if err := c.BodyParser(out); err != nil {
		return err
	}
	return nil
}

type registerRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type updateProfileRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type assignRoleRequest struct {
	Role string `json:"role"`
}

type tokenResponse struct {
	AccessToken      string `json:"accessToken"`
	RefreshToken     string `json:"refreshToken"`
	AccessExpiresIn  int64  `json:"accessExpiresIn"`
	RefreshExpiresIn int64  `json:"refreshExpiresIn"`
}

func tokenPair(pair users.TokenPair) tokenResponse {
	return tokenResponse{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		AccessExpiresIn:  pair.AccessExpiresIn,
		RefreshExpiresIn: pair.RefreshExpiresIn,
	}
}

func profilePayload(profile *users.Profile) fiber.Map {
	return fiber.Map{
		"id":          profile.ID,
		"email":       profile.Email,
		"firstName":   profile.FirstName,
		"lastName":    profile.LastName,
		"status":      profile.Status,
		"roles":       profile.Roles,
		"retrievedAt": time.Now().UTC().Format(time.RFC3339),
	}
}
