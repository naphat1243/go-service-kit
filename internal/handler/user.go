package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/naphat1243/go-service-kit/internal/pkg/apperror"
	"github.com/naphat1243/go-service-kit/internal/service"
)

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) Register(e *echo.Echo) {
	g := e.Group("/api/v1/users")
	g.POST("", h.create)
	g.GET("", h.list)
	g.GET("/:id", h.get)
	g.PATCH("/:id", h.update)
	g.DELETE("/:id", h.delete)
}

func (h *UserHandler) create(c echo.Context) error {
	var in service.CreateUserInput
	if err := c.Bind(&in); err != nil {
		return apperror.Invalid("malformed JSON body")
	}
	u, err := h.users.Create(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, u)
}

func (h *UserHandler) get(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	u, err := h.users.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, u)
}

func (h *UserHandler) list(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	users, err := h.users.List(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"items": users, "count": len(users)})
}

func (h *UserHandler) update(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var in service.UpdateUserInput
	if err := c.Bind(&in); err != nil {
		return apperror.Invalid("malformed JSON body")
	}
	u, err := h.users.Update(c.Request().Context(), id, in)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, u)
}

func (h *UserHandler) delete(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	if err := h.users.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func parseID(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, apperror.Invalid("id must be a valid UUID")
	}
	return id, nil
}
