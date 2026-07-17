package product

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	service ProductService
}

func NewProductController(service ProductService) *ProductController {
	return &ProductController{service: service}
}

func (h *ProductController) Create(c *gin.Context) {
	producerID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current user is required"})
		return
	}

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.service.Create(c.Request.Context(), producerID, req)
	if err != nil {
		handleProductError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": product})
}

func (h *ProductController) Search(c *gin.Context) {
	page, limit := paginationFromQuery(c)
	query := c.Query("q")
	if query == "" {
		query = c.Query("search")
	}

	products, err := h.service.Search(c.Request.Context(), query, c.Query("category"), page, limit)
	if err != nil {
		handleProductError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": products})
}

func (h *ProductController) ListProducerProducts(c *gin.Context) {
	producerID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current user is required"})
		return
	}

	page, limit := paginationFromQuery(c)
	products, err := h.service.ListProducerProducts(c.Request.Context(), producerID, page, limit)
	if err != nil {
		handleProductError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": products})
}

func (h *ProductController) GetByID(c *gin.Context) {
	id, ok := idFromParam(c, "id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	product, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleProductError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *ProductController) Update(c *gin.Context) {
	producerID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current user is required"})
		return
	}
	id, ok := idFromParam(c, "id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.service.Update(c.Request.Context(), producerID, id, req)
	if err != nil {
		handleProductError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *ProductController) Delete(c *gin.Context) {
	producerID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current user is required"})
		return
	}
	id, ok := idFromParam(c, "id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), producerID, id); err != nil {
		handleProductError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProductController) UpdateStock(c *gin.Context) {
	producerID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current user is required"})
		return
	}
	id, ok := idFromParam(c, "id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req UpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.service.UpdateStock(c.Request.Context(), producerID, id, req)
	if err != nil {
		handleProductError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": product})
}

func handleProductError(c *gin.Context, err error) {
	var validationError ValidationError
	switch {
	case errors.Is(err, ErrProductNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrProductForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
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
