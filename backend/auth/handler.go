package auth

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	svc *Service
	v   *validator.Validate
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, v: validator.New()}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Post("/auth/register", h.Register)
	app.Post("/auth/registor", h.Register)
	app.Post("/auth/login", h.Login)
}

func (h *Handler) Register(c fiber.Ctx) error {
	var dto RegisterDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.v.Struct(dto); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.svc.Register(c.Context(), dto)
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(user)
}

func (h *Handler) Login(c fiber.Ctx) error {
	var dto LoginDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.v.Struct(dto); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.svc.Login(c.Context(), dto)
	if err != nil {
		if errors.Is(err, ErrInvalidCredential) {
			return c.Status(401).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(user)
}
