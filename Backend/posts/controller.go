package posts

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"garuda-hacks/backend/auth"
	"garuda-hacks/backend/users"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

func (h *Controller) CreateSupply(c *gin.Context) {
	principal, ok := principalFromContext(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	var req CreateSupplyPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	post, err := h.service.CreateSupply(c.Request.Context(), principal, req)
	if err != nil {
		handlePostError(c, err)
		return
	}
	respondJSON(c, http.StatusCreated, "supply post created successfully", post)
}

func (h *Controller) ListSupply(c *gin.Context) {
	principal, _ := principalFromContext(c)
	filter, err := filterFromQuery(c)
	if err != nil {
		handlePostError(c, err)
		return
	}
	items, err := h.service.ListSupply(c.Request.Context(), principal, filter)
	if err != nil {
		handlePostError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, "", ListResponse{Items: items, Page: filter.Page, Limit: filter.Limit})
}

func (h *Controller) GetSupply(c *gin.Context) {
	principal, _ := principalFromContext(c)
	post, err := h.service.GetSupply(c.Request.Context(), principal, c.Param("id"))
	if err != nil {
		handlePostError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, "", post)
}

func (h *Controller) UpdateSupply(c *gin.Context) {
	principal, ok := principalFromContext(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	var req UpdateSupplyPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	post, err := h.service.UpdateSupply(c.Request.Context(), principal, c.Param("id"), req)
	if err != nil {
		handlePostError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, "supply post updated successfully", post)
}

func (h *Controller) DeleteSupply(c *gin.Context) {
	principal, ok := principalFromContext(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	if err := h.service.DeleteSupply(c.Request.Context(), principal, c.Param("id")); err != nil {
		handlePostError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Controller) CloseSupply(c *gin.Context) {
	principal, ok := principalFromContext(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	post, err := h.service.CloseSupply(c.Request.Context(), principal, c.Param("id"))
	if err != nil {
		handlePostError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, "supply post closed successfully", post)
}

func (h *Controller) MySupply(c *gin.Context) {
	principal, ok := principalFromContext(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	filter, err := filterFromQuery(c)
	if err != nil {
		handlePostError(c, err)
		return
	}
	items, err := h.service.ListMySupply(c.Request.Context(), principal, filter)
	if err != nil {
		handlePostError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, "", ListResponse{Items: items, Page: filter.Page, Limit: filter.Limit})
}

func (h *Controller) CreateDemand(c *gin.Context) {
	principal, ok := principalFromContext(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	var req CreateDemandPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	post, err := h.service.CreateDemand(c.Request.Context(), principal, req)
	if err != nil {
		handlePostError(c, err)
		return
	}
	respondJSON(c, http.StatusCreated, "demand post created successfully", post)
}

func (h *Controller) ListDemand(c *gin.Context) {
	principal, _ := principalFromContext(c)
	filter, err := filterFromQuery(c)
	if err != nil {
		handlePostError(c, err)
		return
	}
	items, err := h.service.ListDemand(c.Request.Context(), principal, filter)
	if err != nil {
		handlePostError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, "", ListResponse{Items: items, Page: filter.Page, Limit: filter.Limit})
}

func (h *Controller) GetDemand(c *gin.Context) {
	principal, _ := principalFromContext(c)
	post, err := h.service.GetDemand(c.Request.Context(), principal, c.Param("id"))
	if err != nil {
		handlePostError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, "", post)
}

func (h *Controller) UpdateDemand(c *gin.Context) {
	principal, ok := principalFromContext(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	var req UpdateDemandPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	post, err := h.service.UpdateDemand(c.Request.Context(), principal, c.Param("id"), req)
	if err != nil {
		handlePostError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, "demand post updated successfully", post)
}

func (h *Controller) DeleteDemand(c *gin.Context) {
	principal, ok := principalFromContext(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	if err := h.service.DeleteDemand(c.Request.Context(), principal, c.Param("id")); err != nil {
		handlePostError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Controller) CloseDemand(c *gin.Context) {
	principal, ok := principalFromContext(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	post, err := h.service.CloseDemand(c.Request.Context(), principal, c.Param("id"))
	if err != nil {
		handlePostError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, "demand post closed successfully", post)
}

func (h *Controller) MyDemand(c *gin.Context) {
	principal, ok := principalFromContext(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	filter, err := filterFromQuery(c)
	if err != nil {
		handlePostError(c, err)
		return
	}
	items, err := h.service.ListMyDemand(c.Request.Context(), principal, filter)
	if err != nil {
		handlePostError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, "", ListResponse{Items: items, Page: filter.Page, Limit: filter.Limit})
}

func (h *Controller) Feed(c *gin.Context) {
	principal, _ := principalFromContext(c)
	filter, err := filterFromQuery(c)
	if err != nil {
		handlePostError(c, err)
		return
	}
	items, err := h.service.Feed(c.Request.Context(), principal, filter)
	if err != nil {
		handlePostError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, "", ListResponse{Items: items, Page: filter.Page, Limit: filter.Limit})
}

func (h *Controller) Search(c *gin.Context) {
	principal, _ := principalFromContext(c)
	filter, err := filterFromQuery(c)
	if err != nil {
		handlePostError(c, err)
		return
	}
	items, err := h.service.Search(c.Request.Context(), principal, filter)
	if err != nil {
		handlePostError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, "", ListResponse{Items: items, Page: filter.Page, Limit: filter.Limit})
}

func handlePostError(c *gin.Context, err error) {
	var validationError ValidationError
	switch {
	case errors.Is(err, ErrUnauthorized):
		respondError(c, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrInvalidRole), errors.Is(err, ErrPostForbidden):
		respondError(c, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrPostNotFound):
		respondError(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidPostType), errors.Is(err, ErrInvalidSort), errors.Is(err, ErrInvalidQueryFilter):
		respondError(c, http.StatusBadRequest, err.Error())
	case errors.As(err, &validationError):
		respondError(c, http.StatusBadRequest, validationError.Error())
	default:
		respondError(c, http.StatusInternalServerError, "internal server error")
	}
}

func principalFromContext(c *gin.Context) (Principal, bool) {
	if claims, ok := auth.GetCurrentUser(c.Request.Context()); ok {
		return Principal{UserID: claims.UserID, Role: claims.Role}, true
	}

	userID, idOK := stringFromGin(c, "CurrentUserID", "currentUserID", "user_id", "userID")
	role, roleOK := stringFromGin(c, "CurrentUserRole", "currentUserRole", "role", "userRole")
	if !idOK {
		return Principal{}, false
	}
	return Principal{UserID: userID, Role: users.UserRole(strings.ToUpper(role))}, roleOK
}

func stringFromGin(c *gin.Context, keys ...string) (string, bool) {
	for _, key := range keys {
		value, exists := c.Get(key)
		if !exists {
			continue
		}
		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v), true
			}
		case users.UserRole:
			return string(v), true
		}
	}
	return "", false
}

func filterFromQuery(c *gin.Context) (QueryFilter, error) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	filter := QueryFilter{
		Type:               c.Query("type"),
		Query:              firstQuery(c, "q", "search"),
		Category:           c.Query("category"),
		Subcategory:        c.Query("subcategory"),
		Location:           c.Query("location"),
		Status:             c.Query("status"),
		Unit:               c.Query("unit"),
		AvailabilityStatus: c.Query("availability_status"),
		Sort:               c.Query("sort"),
		Page:               page,
		Limit:              limit,
	}
	var err error
	if filter.PriceMin, err = floatQuery(c, "price_min"); err != nil {
		return filter, err
	}
	if filter.PriceMax, err = floatQuery(c, "price_max"); err != nil {
		return filter, err
	}
	if filter.BudgetMin, err = floatQuery(c, "budget_min"); err != nil {
		return filter, err
	}
	if filter.BudgetMax, err = floatQuery(c, "budget_max"); err != nil {
		return filter, err
	}
	if filter.QuantityMin, err = floatQuery(c, "quantity_min"); err != nil {
		return filter, err
	}
	if filter.QuantityMax, err = floatQuery(c, "quantity_max"); err != nil {
		return filter, err
	}
	if filter.NeededFrom, err = dateQuery(c, "needed_from"); err != nil {
		return filter, err
	}
	if filter.NeededUntil, err = dateQuery(c, "needed_until"); err != nil {
		return filter, err
	}
	if filter.CreatedFrom, err = dateQuery(c, "created_from"); err != nil {
		return filter, err
	}
	if filter.CreatedUntil, err = dateQuery(c, "created_until"); err != nil {
		return filter, err
	}
	filter.Page, filter.Limit, filter.Offset = normalizePagination(filter.Page, filter.Limit)
	return filter, nil
}

func firstQuery(c *gin.Context, names ...string) string {
	for _, name := range names {
		if value := c.Query(name); value != "" {
			return value
		}
	}
	return ""
}

func floatQuery(c *gin.Context, name string) (*float64, error) {
	value := strings.TrimSpace(c.Query(name))
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, ErrInvalidQueryFilter
	}
	return &parsed, nil
}

func dateQuery(c *gin.Context, name string) (*time.Time, error) {
	value := strings.TrimSpace(c.Query(name))
	if value == "" {
		return nil, nil
	}
	for _, layout := range []string{"2006-01-02", time.RFC3339} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return &parsed, nil
		}
	}
	return nil, ErrInvalidQueryFilter
}

func respondJSON(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"error":   message,
	})
}
