package baseStation

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"simpleServer/internal/baseStation/database"
	"simpleServer/internal/config"
	"simpleServer/internal/middleware"
	"simpleServer/internal/middleware/handler"
	"simpleServer/pkg/logging"
	"simpleServer/pkg/validate"
)

type Handler struct {
	baseStationDB database.BaseStationDB
}

func NewHandler(baseStationDB database.BaseStationDB) *Handler {
	return &Handler{
		baseStationDB: baseStationDB,
	}
}

func (h *Handler) GetCluster(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		type RequestUri struct {
			id int64 `uri:"id"`
		}
		var uri RequestUri
		if err := c.ShouldBindUri(&uri); err != nil {
			logger.Errorf("baseStations.GetCluster failed to bind", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&uri, "uri", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidUriValue, "invalid nw, se or zoom in uri", details)
		}

		ctx := context.Background()
		bs, err := h.baseStationDB.GetBaseStationById(ctx, uint64(uri.id))
		if err != nil {
			logger.Errorf("baseStations.GetCluster failed to cluster", "err", err)
			return handler.NewInternalErrorResponse(fmt.Errorf("Can't obtain clusters"))
		}
		return handler.NewSuccessResponse(http.StatusOK, NewBaseStationResponse(bs))
	})
}

func (h *Handler) GetClusters(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		type RequestUri struct {
			N    float64 `uri:"n"`
			W    float64 `uri:"w"`
			S    float64 `uri:"s"`
			E    float64 `uri:"e"`
			Zoom float32 `uri:"zoom"`
		}
		var uri RequestUri
		if err := c.ShouldBindUri(&uri); err != nil {
			logger.Errorf("baseStations.GetClusters failed to bind", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&uri, "uri", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidUriValue, "invalid nw, se or zoom in uri", details)
		}

		points, err := h.baseStationDB.GetClusters(c.Request.Context(), uri.N, uri.W, uri.S, uri.E, uri.Zoom)
		if err != nil {
			logger.Errorf("baseStations.GetCluster failed to cluster", "err", err)
			return handler.NewInternalErrorResponse(fmt.Errorf("Can't obtain clusters"))
		}
		return handler.NewSuccessResponse(http.StatusOK, NewClusterResponse(points))
	})
}

func RouteV1(cfg *config.Config, h *Handler, r *gin.Engine) {
	v1 := r.Group("v1/api")
	v1.Use(middleware.CorsMiddleware(), middleware.RequestIDMiddleware(), middleware.TimeoutMiddleware(cfg.ServerConfig.WriteTimeout))

	baseStationV1 := v1.Group("baseStations")
	baseStationV1.Use()
	{
		baseStationV1.GET("/nw/:n/:w/se/:s/:e/zoom/:zoom", h.GetClusters)
		baseStationV1.GET("/id/:id", h.GetCluster)
	}
}
