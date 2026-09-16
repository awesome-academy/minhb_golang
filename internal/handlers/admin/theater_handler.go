package admin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"

	"cinema-booking/internal/dto"
	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/middleware"
	"cinema-booking/internal/models"
	"cinema-booking/internal/services"
	"cinema-booking/internal/utils"
)

const theatersPath = middleware.AdminPathPrefix + "/theaters"

type TheaterRow struct {
	ID       int64
	Name     string
	Slug     string
	City     string
	Address  string
	Phone    string
	IsActive bool
}

type TheaterListView struct {
	Title      string
	AdminEmail string
	CSRFToken  string
	Flash      Flash
	Q          string
	Theaters   []TheaterRow
	Total      int64
	Page       int
	TotalPages int
	PrevPage   int
	NextPage   int
	HasPrev    bool
	HasNext    bool
}

type TheaterFormView struct {
	Title      string
	AdminEmail string
	CSRFToken  string
	Action     string
	Submit     string
	Slug       string
	IsEdit     bool
	Form       dto.AdminTheaterForm
	Errors     map[string]string
	Error      string
}

type TheaterHandler struct {
	service services.AdminTheaterService
	flash   flashCookie
}

func NewTheaterHandler(service services.AdminTheaterService, secureCookies bool) *TheaterHandler {
	return &TheaterHandler{service: service, flash: flashCookie{secure: secureCookies}}
}

func (h *TheaterHandler) List(c *echo.Context) error {
	q := strings.TrimSpace(c.QueryParam("q"))
	page := parsePage(c.QueryParam("page"))
	theaters, total, err := h.service.List(c.Request().Context(), q, page)
	if err != nil {
		return utils.ServiceError(err)
	}
	totalPages := int((total + services.TheaterPageSize - 1) / services.TheaterPageSize)
	if totalPages < 1 {
		totalPages = 1
	}
	return c.Render(http.StatusOK, "admin/theaters/list", TheaterListView{
		Title:      "Theaters",
		AdminEmail: adminEmail(c),
		CSRFToken:  csrfToken(c),
		Flash:      h.flash.pop(c),
		Q:          q,
		Theaters:   toTheaterRows(theaters),
		Total:      total,
		Page:       page,
		TotalPages: totalPages,
		PrevPage:   page - 1,
		NextPage:   page + 1,
		HasPrev:    page > 1,
		HasNext:    page < totalPages,
	})
}

func (h *TheaterHandler) New(c *echo.Context) error {
	return h.renderForm(c, http.StatusOK, h.newView(c, dto.AdminTheaterForm{IsActive: true}))
}

func (h *TheaterHandler) Create(c *echo.Context) error {
	var form dto.AdminTheaterForm
	if err := utils.BindAndValidate(c, &form); err != nil {
		return h.renderFormError(c, h.newView(c, form), err)
	}
	if _, err := h.service.Create(c.Request().Context(), form); err != nil {
		return h.renderFormError(c, h.newView(c, form), err)
	}
	h.flash.set(c, flashSuccess, "Theater created")
	return c.Redirect(http.StatusSeeOther, theatersPath)
}

func (h *TheaterHandler) Edit(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	theater, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		return utils.ServiceError(err)
	}
	return h.renderForm(c, http.StatusOK, h.editView(c, theater, formFromTheater(theater)))
}

func (h *TheaterHandler) Update(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	theater, err := h.service.Get(ctx, id)
	if err != nil {
		return utils.ServiceError(err)
	}
	var form dto.AdminTheaterForm
	if err := utils.BindAndValidate(c, &form); err != nil {
		return h.renderFormError(c, h.editView(c, theater, form), err)
	}
	if err := h.service.Update(ctx, id, form); err != nil {
		if errors.Is(err, apperrors.ErrRecordModified) || errors.Is(err, apperrors.ErrRecordTokenInvalid) {
			form.UpdatedAt = updatedAtToken(theater.UpdatedAt)
		}
		return h.renderFormError(c, h.editView(c, theater, form), err)
	}
	h.flash.set(c, flashSuccess, "Theater updated")
	return c.Redirect(http.StatusSeeOther, theatersPath)
}

func (h *TheaterHandler) ChangeStatus(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	active, err := h.service.ChangeStatus(c.Request().Context(), id)
	if err != nil {
		return utils.ServiceError(err)
	}
	if active {
		h.flash.set(c, flashSuccess, "Theater activated")
	} else {
		h.flash.set(c, flashSuccess, "Theater deactivated")
	}
	return c.Redirect(http.StatusSeeOther, theatersPath)
}

func (h *TheaterHandler) newView(c *echo.Context, form dto.AdminTheaterForm) TheaterFormView {
	return TheaterFormView{
		Title:      "New theater",
		AdminEmail: adminEmail(c),
		CSRFToken:  csrfToken(c),
		Action:     theatersPath,
		Submit:     "Create theater",
		Form:       form,
	}
}

func (h *TheaterHandler) editView(c *echo.Context, theater *models.Theater, form dto.AdminTheaterForm) TheaterFormView {
	return TheaterFormView{
		Title:      "Edit theater",
		AdminEmail: adminEmail(c),
		CSRFToken:  csrfToken(c),
		Action:     theatersPath + "/" + strconv.FormatInt(theater.ID, 10),
		Submit:     "Save changes",
		Slug:       theater.Slug,
		IsEdit:     true,
		Form:       form,
	}
}

func (h *TheaterHandler) renderForm(c *echo.Context, status int, view TheaterFormView) error {
	if view.Errors == nil {
		view.Errors = map[string]string{}
	}
	return c.Render(status, "admin/theaters/form", view)
}

func (h *TheaterHandler) renderFormError(c *echo.Context, view TheaterFormView, err error) error {
	if errors.Is(err, apperrors.ErrRecordModified) {
		view.Error = "Someone else changed this theater while you were editing. Reload to see the latest data, or save again to overwrite it."
		return h.renderForm(c, http.StatusConflict, view)
	}
	if errors.Is(err, apperrors.ErrRecordTokenInvalid) {
		view.Error = "The form was submitted without a valid version token, so nothing was saved. Submit again to retry."
		return h.renderForm(c, http.StatusBadRequest, view)
	}
	fields, message, ok := fieldErrors(err)
	if !ok {
		return utils.ServiceError(err)
	}
	view.Errors, view.Error = fields, message
	return h.renderForm(c, http.StatusUnprocessableEntity, view)
}

func toTheaterRows(theaters []models.Theater) []TheaterRow {
	rows := make([]TheaterRow, 0, len(theaters))
	for _, theater := range theaters {
		rows = append(rows, TheaterRow{
			ID:       theater.ID,
			Name:     theater.Name,
			Slug:     theater.Slug,
			City:     theater.City,
			Address:  theater.Address,
			Phone:    derefString(theater.Phone),
			IsActive: theater.IsActive,
		})
	}
	return rows
}

func formFromTheater(theater *models.Theater) dto.AdminTheaterForm {
	return dto.AdminTheaterForm{
		Name:        theater.Name,
		Address:     theater.Address,
		City:        theater.City,
		Phone:       derefString(theater.Phone),
		Description: derefString(theater.Description),
		ImageURL:    derefString(theater.ImageURL),
		IsActive:    theater.IsActive,
		UpdatedAt:   updatedAtToken(theater.UpdatedAt),
	}
}
