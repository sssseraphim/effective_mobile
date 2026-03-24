package handlers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/sssseraphim/effective_mobile/internal/models"
	"github.com/sssseraphim/effective_mobile/internal/repository"
	"net/http"
	"strconv"
	"time"
)

type SubscriptionHandler struct {
	repo *repository.Repository
}

func NewSubscriptionHandler(repo *repository.Repository) *SubscriptionHandler {
	return &SubscriptionHandler{repo: repo}
}

// CreateSubscription godoc
// @Summary Create a new subscription
// @Description Create a new subscription record
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var req models.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Warn("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		log.WithError(err).Warn("Invalid start_date format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, expected MM-YYYY"})
		return
	}

	var endDate *time.Time
	if req.EndDate != "" {
		parsed, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			log.WithError(err).Warn("Invalid end_date format")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, expected MM-YYYY"})
			return
		}
		endDate = &parsed
	}

	subscription := &models.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if err := h.repo.Create(subscription); err != nil {
		log.WithError(err).Error("Failed to create subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

// GetSubscription godoc
// @Summary Get subscription by ID
// @Description Get a single subscription by its UUID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} models.Subscription
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.WithError(err).Warn("Invalid UUID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
		return
	}

	subscription, err := h.repo.GetByID(id)
	if err != nil {
		log.WithError(err).Error("Failed to get subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription"})
		return
	}

	if subscription == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// ListSubscriptions godoc
// @Summary List all subscriptions
// @Description Get a list of subscriptions with pagination
// @Tags subscriptions
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} models.Subscription
// @Failure 500 {object} map[string]string
// @Router /subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	subscriptions, err := h.repo.List(limit, offset)
	if err != nil {
		log.WithError(err).Error("Failed to list subscriptions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list subscriptions"})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

// UpdateSubscription godoc
// @Summary Update a subscription
// @Description Update an existing subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param subscription body models.UpdateSubscriptionRequest true "Subscription update data"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.WithError(err).Warn("Invalid UUID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
		return
	}

	existing, err := h.repo.GetByID(id)
	if err != nil {
		log.WithError(err).Error("Failed to get subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription"})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	var req models.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Warn("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ServiceName != "" {
		existing.ServiceName = req.ServiceName
	}
	if req.Price != nil {
		existing.Price = *req.Price
	}
	if req.StartDate != "" {
		startDate, err := time.Parse("01-2006", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format"})
			return
		}
		existing.StartDate = startDate
	}
	if req.EndDate != nil {
		if *req.EndDate != "" {
			endDate, err := time.Parse("01-2006", *req.EndDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format"})
				return
			}
			existing.EndDate = &endDate
		} else {
			existing.EndDate = nil
		}
	}

	if err := h.repo.Update(id, existing); err != nil {
		log.WithError(err).Error("Failed to update subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, existing)
}

// DeleteSubscription godoc
// @Summary Delete a subscription
// @Description Delete a subscription by ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 204 {object} nil
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.WithError(err).Warn("Invalid UUID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		if err.Error() == "subscription not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}
		log.WithError(err).Error("Failed to delete subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete subscription"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetTotalCost godoc
// @Summary Get total cost of subscriptions
// @Description Calculate total cost of subscriptions for a period with filters
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User ID (UUID)"
// @Param service_name query string false "Service name"
// @Param start_date query string true "Start date (MM-YYYY)"
// @Param end_date query string true "End date (MM-YYYY)"
// @Success 200 {object} models.TotalCostResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/total-cost [get]
func (h *SubscriptionHandler) GetTotalCost(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	start, err := time.Parse("01-2006", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, expected MM-YYYY"})
		return
	}

	end, err := time.Parse("01-2006", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, expected MM-YYYY"})
		return
	}

	var userID uuid.UUID
	userIDStr := c.Query("user_id")
	if userIDStr != "" {
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format"})
			return
		}
	}

	serviceName := c.Query("service_name")

	total, err := h.repo.GetTotalCost(userID, serviceName, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil && err != sql.ErrNoRows {
		log.WithError(err).Error("Failed to calculate total cost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate total cost"})
		return
	}

	c.JSON(http.StatusOK, models.TotalCostResponse{TotalCost: total})
}
