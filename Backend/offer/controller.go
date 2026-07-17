package offer

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OfferController struct {
	service OfferService
}

func NewOfferController(service OfferService) *OfferController {
	return &OfferController{service: service}
}

func (h *OfferController) Create(c *gin.Context) {
	producerID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current user is required"})
		return
	}

	var req CreateOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	offer, err := h.service.Create(c.Request.Context(), producerID, req)
	if err != nil {
		handleOfferError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": offer})
}

func (h *OfferController) ListProducerOffers(c *gin.Context) {
	producerID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current user is required"})
		return
	}

	page, limit := paginationFromQuery(c)
	offers, err := h.service.ListProducerOffers(c.Request.Context(), producerID, page, limit)
	if err != nil {
		handleOfferError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": offers})
}

func (h *OfferController) GetByID(c *gin.Context) {
	id, ok := idFromParam(c, "id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offer id"})
		return
	}

	offer, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleOfferError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": offer})
}

func (h *OfferController) Update(c *gin.Context) {
	producerID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current user is required"})
		return
	}
	id, ok := idFromParam(c, "id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offer id"})
		return
	}

	var req UpdateOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	offer, err := h.service.Update(c.Request.Context(), producerID, id, req)
	if err != nil {
		handleOfferError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": offer})
}

func (h *OfferController) Cancel(c *gin.Context) {
	producerID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current user is required"})
		return
	}
	id, ok := idFromParam(c, "id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offer id"})
		return
	}

	if err := h.service.Cancel(c.Request.Context(), producerID, id); err != nil {
		handleOfferError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *OfferController) ListByDemandGroup(c *gin.Context) {
	demandGroupID, ok := idFromParam(c, "id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid demand group id"})
		return
	}

	page, limit := paginationFromQuery(c)
	offers, err := h.service.ListByDemandGroup(c.Request.Context(), demandGroupID, page, limit)
	if err != nil {
		handleOfferError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": offers})
}

func handleOfferError(c *gin.Context, err error) {
	var validationError ValidationError
	switch {
	case errors.Is(err, ErrOfferNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrOfferForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, ErrDuplicateOffer):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, ErrOfferCancelled):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, ErrDemandGroupIDZero):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.As(err, &validationError):
		c.JSON(http.StatusBadRequest, gin.H{"error": validationError.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

func paginationFromQuery(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	return page, limit
}

func idFromParam(c *gin.Context, name string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || value == 0 {
		return 0, false
	}
	return uint(value), true
}

func currentUserID(c *gin.Context) (uint, bool) {
	keys := []string{"CurrentUserID", "currentUserID", "user_id", "userID", "producer_id", "producerID"}
	for _, key := range keys {
		value, exists := c.Get(key)
		if !exists {
			continue
		}
		if id, ok := toUint(value); ok && id > 0 {
			return id, true
		}
	}
	return 0, false
}

func toUint(value interface{}) (uint, bool) {
	switch v := value.(type) {
	case uint:
		return v, true
	case uint8:
		return uint(v), true
	case uint16:
		return uint(v), true
	case uint32:
		return uint(v), true
	case uint64:
		return uint(v), true
	case int:
		if v < 0 {
			return 0, false
		}
		return uint(v), true
	case int8:
		if v < 0 {
			return 0, false
		}
		return uint(v), true
	case int16:
		if v < 0 {
			return 0, false
		}
		return uint(v), true
	case int32:
		if v < 0 {
			return 0, false
		}
		return uint(v), true
	case int64:
		if v < 0 {
			return 0, false
		}
		return uint(v), true
	case float64:
		if v < 0 {
			return 0, false
		}
		return uint(v), true
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return uint(parsed), true
	default:
		return 0, false
	}
}
