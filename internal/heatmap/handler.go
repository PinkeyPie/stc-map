package heatmap

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"simpleServer/internal/config"
	"simpleServer/internal/heatmap/database"
	"simpleServer/internal/heatmap/model"
	"simpleServer/internal/middleware"
	"simpleServer/internal/middleware/handler"
	"simpleServer/pkg/logging"
	"simpleServer/pkg/validate"
)

type Handler struct {
	heatmapDB database.HeatmapDB
}

func NewHandler(db database.HeatmapDB) *Handler { return &Handler{heatmapDB: db} }

func (h *Handler) GetHeatMapPointsInBbox(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		type RequestUri struct {
			N float64 `uri:"n"`
			W float64 `uri:"w"`
			S float64 `uri:"s"`
			E float64 `uri:"e"`
		}
		var uri RequestUri
		if err := c.ShouldBindUri(&uri); err != nil {
			logger.Errorf("heatmap uri parse error: %v", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&uri, "uri", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidUriValue, "invalid nw, se", details)
		}
		var points []model.HeatmapPoint
		var err error
		if points, err = h.heatmapDB.GetAllHeatmapPointsInBbox(c, uri.N, uri.W, uri.S, uri.W); err != nil {
			logger.Errorf("GetHeatMapPointsInBbox err: %v", err)
			return handler.NewInternalErrorResponse(err)
		}

		return handler.NewSuccessResponse(http.StatusOK, NewHeatmapPointsInBboxResponse(points))
	})
}

func (h *Handler) GetHeatMapPointsByBsId(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		type RequestUri struct {
			Id int `uri:"id"`
		}
		var uri RequestUri
		if err := c.ShouldBindUri(&uri); err != nil {
			logger.Errorf("heatmap uri parse error: %v", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&uri, "uri", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidUriValue, "invalid bs id", details)
		}
		var points []model.HeatmapPoint
		var err error
		if points, err = h.heatmapDB.GetHeatmapPointsByIdDB(c, uri.Id); err != nil {
			logger.Errorf("GetHeatMapPointsInBbox err: %v", err)
			return handler.NewInternalErrorResponse(err)
		}

		return handler.NewSuccessResponse(http.StatusOK, NewHeatmapPointsInBboxResponse(points))
	})
}

func (h *Handler) GetHeatMapPointsByCoords(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		type RequestUri struct {
			Lat float64 `uri:"lat"`
			Lng float64 `uri:"lng"`
		}
		var uri RequestUri
		if err := c.ShouldBindUri(&uri); err != nil {
			logger.Errorf("heatmap uri parse error: %v", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&uri, "uri", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidUriValue, "invalid bs id", details)
		}
		var points []model.HeatmapPoint
		var err error
		if points, err = h.heatmapDB.GetAllHeatmapPointsByCoordsDB(c, uri.Lat, uri.Lng); err != nil {
			logger.Errorf("GetHeatmapPointsByCoordsDB err: %v", err)
			return handler.NewInternalErrorResponse(err)
		}

		return handler.NewSuccessResponse(http.StatusOK, NewHeatmapPointsInBboxResponse(points))
	})
}

func RouteV1(cfg *config.Config, h *Handler, r *gin.Engine) {
	v1 := r.Group("v1/api")
	v1.Use(middleware.CorsMiddleware(), middleware.RequestIDMiddleware(), middleware.TimeoutMiddleware(cfg.ServerConfig.WriteTimeout))

	heatmapV1 := v1.Group("heatmap")
	heatmapV1.Use()
	{
		heatmapV1.GET("/nw/:n/:w/se/:s/:e", h.GetHeatMapPointsInBbox)
		heatmapV1.GET("/lat/:lat/lng/:lng", h.GetHeatMapPointsByCoords)
		// Not work for now but maybe need later
		//heatmapV1.GET("/id/:id", h.GetHeatMapPointsByBsId)
	}

}
