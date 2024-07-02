package post

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"simpleServer/internal/config"
	"simpleServer/internal/middleware"
	"simpleServer/internal/middleware/handler"
	"simpleServer/internal/post/database"
	"simpleServer/internal/post/model"
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

}

func RouteV1(cfg *config.Config, h *Handler, r *gin.Engine) {
	v1 := r.Group("v1/api")
	v1.Use(middleware.CorsMiddleware(), middleware.RequestIDMiddleware(), middleware.TimeoutMiddleware(cfg.ServerConfig.WriteTimeout))

	postsV1 := v1.Group("post")
	postsV1.Use()
	{
		postsV1.GET("/all", h.GetPosts)
		postsV1.GET("/:id", h.GetPostPath)
	}
}
