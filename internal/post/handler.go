package post

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"net/http"
	"simpleServer/internal/config"
	"simpleServer/internal/middleware"
	"simpleServer/internal/middleware/handler"
	"simpleServer/internal/post/database"
	"simpleServer/internal/post/model"
	"simpleServer/pkg/logging"
	"simpleServer/pkg/validate"
	"time"
)

type Handler struct {
	postDB database.PostDB
}

func NewHandler(db database.PostDB) *Handler {
	return &Handler{postDB: db}
}

func (h *Handler) GetPosts(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		var posts []model.Post
		var err error
		if posts, err = h.postDB.GetAllPosts(c); err != nil {
			return handler.NewInternalErrorResponse(err)
		}

		if len(posts) != 0 {
			for i, post := range posts {
				gpsData, err := h.postDB.GetPostScanDates(c, post.Id)
				if err != nil {
					return handler.NewInternalErrorResponse(err)
				}
				posts[i].Coordinates = gpsData
			}
		}

		return handler.NewSuccessResponse(http.StatusOK, NewPostDateResponse(posts))
	})
}

func (h *Handler) GetPostPath(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		type RequestUrl struct {
			Id      string `uri:"id"`
			Date    string `uri:"date"`
			Measure int    `uri:"measure"`
		}
		var uri RequestUrl
		if err := c.ShouldBindUri(&uri); err != nil {
			logger.Errorf("post.GetPostPath failed to bind", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&uri, "uri", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidUriValue, "invalid id in uri", details)
		}

		ctx := context.Background()
		post, err := h.postDB.GetPostById(ctx, uuid.FromStringOrNil(uri.Id))
		if err != nil {
			return handler.NewInternalErrorResponse(fmt.Errorf("can't find post with such id"))
		}
		date, err := time.Parse("02.January.2006", uri.Date)
		if err != nil {
			return handler.NewInternalErrorResponse(fmt.Errorf("can't parse provided date"))
		}
		err = h.postDB.GetPostPath(ctx, post, date, uri.Measure)
		if err != nil {
			return handler.NewInternalErrorResponse(fmt.Errorf("can't find post path values"))
		}
		//return handler.NewSuccessResponse(http.StatusOK, NewPostPathResponse(post))
		c.Header("Content-Type", "application/json")
		return handler.NewSuccessResponse(http.StatusOK, NewPostPathResponse(post))
	})
}

func RouteV1(cfg *config.Config, h *Handler, r *gin.Engine) {
	v1 := r.Group("v1/api")
	v1.Use(middleware.CorsMiddleware(), middleware.RequestIDMiddleware(), middleware.TimeoutMiddleware(cfg.ServerConfig.WriteTimeout))

	postsV1 := v1.Group("post")
	postsV1.Use()
	{
		postsV1.GET("/all", h.GetPosts)
		postsV1.GET("id/:id/date/:date/measure/:measure", h.GetPostPath)
	}
}
