package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rekberkuy/core-service/internal/usecase"
)

type CatalogHandler struct {
	catalogUsecase *usecase.CatalogUsecase
}

func NewCatalogHandler(cu *usecase.CatalogUsecase) *CatalogHandler {
	return &CatalogHandler{catalogUsecase: cu}
}

// GetCategoryCatalogHandler serves the public 3-tier taxonomy.
func (h *CatalogHandler) GetCategoryCatalogHandler(c *gin.Context) {
	catalog, err := h.catalogUsecase.GetCategoryCatalog(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": catalog})
}

// ListMarketplaceVendorsHandler browses the vendor marketplace; supports
// ?category=CATERING plus pagination.
func (h *CatalogHandler) ListMarketplaceVendorsHandler(c *gin.Context) {
	limit, offset := pagination(c)
	vendors, err := h.catalogUsecase.ListMarketplaceVendors(c.Request.Context(), c.Query("category"), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": vendors})
}
