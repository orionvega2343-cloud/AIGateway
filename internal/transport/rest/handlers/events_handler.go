package handlers

import (
	"AIGateway/internal/service"
	"AIGateway/internal/worker"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EventsHandler interface {
	Post(c *gin.Context)
	GetEventById(c *gin.Context)
}

type EventsHandlerImpl struct {
	Es *service.EventsService
	P  *worker.Pipeline
}

func NewEventsHandler(es *service.EventsService, p *worker.Pipeline) *EventsHandlerImpl {
	return &EventsHandlerImpl{Es: es, P: p}
}

func (h *EventsHandlerImpl) Post(c *gin.Context) {
	//DTO тип, для предотвращения передачи ненужных данных
	var e struct {
		ExternalId string `json:"external_id"`
		Payload    string `json:"payload"`
	}

	//Бинд из DTO в JSON
	err := c.ShouldBind(&e)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := h.Es.CreateEvent(c.Request.Context(), e.ExternalId, e.Payload)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateEvent) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	_ = h.P.Producer(c.Request.Context(), event.Id, e.Payload)

	c.JSON(http.StatusOK, event)
}

func (h *EventsHandlerImpl) GetEventById(c *gin.Context) {
	id := c.Param("id")

	parsed, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	event, err := h.Es.GetEventById(c.Request.Context(), parsed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}
