package post

import (
	"fmt"
	"github.com/gofrs/uuid"
	"simpleServer/internal/post/model"
	"time"
)

type Post struct {
	PostId   uuid.UUID              `json:"postId"`
	Name     string                 `json:"name"`
	Children map[string]interface{} `json:"children"`
}

type TimeNode struct {
	Name      string    `json:"name"`
	RowNumber int       `json:"rowNumber"`
	Date      string    `json:"date"`
	PostId    uuid.UUID `json:"postId"`
	IsTime    bool      `json:"isTime"`
}

type Waypoint struct {
	Coordinates []float64 `json:"coordinates"`
	Timestamp   int       `json:"timestamp"`
}

func (node TimeNode) GetMapNode() map[string]interface{} {
	return map[string]interface{}{
		"name":      node.Name,
		"rowNumber": node.RowNumber,
		"date":      node.Date,
		"postId":    node.PostId,
		"isTime":    node.IsTime,
	}
}

type JsonNodeDetour struct {
	Graph      map[string]interface{}
	DetourNode string
}

func (detour *JsonNodeDetour) JsonDetour() {
	detourQueue := []map[string]interface{}{detour.Graph}
	iterator := 1

	for len(detourQueue) > 0 {
		detourNode := detourQueue[0]
		detourQueue = detourQueue[1:]
		if _, ok := detourNode[detour.DetourNode]; ok {
			detourNode["id"] = iterator
			iterator++
		}
		for _, value := range detourNode {
			switch value.(type) {
			case []map[string]interface{}:
				for i := range value.([]map[string]interface{}) {
					detourQueue = append(detourQueue, value.([]map[string]interface{})[i])
				}
			case map[string]interface{}:
				detourQueue = append(detourQueue, value.(map[string]interface{}))
			}
		}
	}
}

func NewPostDateResponse(postsDb []model.Post) map[string]interface{} {
	postData := make(map[string]interface{})
	postData["title"] = "Posts"

	posts := make([]map[string]interface{}, 0)
	for _, postDb := range postsDb {
		post := Post{
			PostId: postDb.Id,
			Name:   postDb.Name,
		}

		scanDates := make([]map[string]interface{}, 0)
		if len(postDb.Coordinates) != 0 {
			scanDate := make(map[string]interface{})
			scanDate["name"] = postDb.Coordinates[0].Time.Format("2.January.2006")
			lastTimeNode := TimeNode{
				Name:      postDb.Coordinates[0].Time.Format("15:04"),
				RowNumber: 0,
				Date:      scanDate["name"].(string),
				PostId:    postDb.Id,
				IsTime:    true,
			}
			scanDate["children"] = []map[string]interface{}{
				{
					"title":   "Scans time",
					"content": []map[string]interface{}{lastTimeNode.GetMapNode()},
				},
			}
			scanDates = append(scanDates, scanDate)
			for i := 1; i < len(postDb.Coordinates); i++ {
				currDate := postDb.Coordinates[i].Time
				if scanDate["name"].(string) == currDate.Format("2.January.2006") {
					first, _ := time.Parse("15:04", lastTimeNode.Name)
					second, _ := time.Parse("15:04", fmt.Sprintf("%d:%d", currDate.Hour(), currDate.Minute()))
					timeDiff := second.Sub(first)
					if timeDiff.Seconds() >= 120 {
						lastTimeNode = TimeNode{
							Name:      currDate.Format("15:04"),
							RowNumber: lastTimeNode.RowNumber + 1,
							Date:      scanDate["name"].(string),
							PostId:    postDb.Id,
							IsTime:    true,
						}
						scanDate["children"].([]map[string]interface{})[0]["content"] = append(scanDate["children"].([]map[string]interface{})[0]["content"].([]map[string]interface{}), lastTimeNode.GetMapNode())
					}
				} else {
					scanDate = make(map[string]interface{})
					scanDate["name"] = currDate.Format("2.January.2006")
					lastTimeNode = TimeNode{
						Name:      currDate.Format("15:04"),
						RowNumber: 0,
						Date:      scanDate["name"].(string),
						PostId:    postDb.Id,
						IsTime:    true,
					}
					scanDate["children"] = []map[string]interface{}{
						{
							"title":   "Scans time",
							"content": []map[string]interface{}{lastTimeNode.GetMapNode()},
						},
					}
					scanDates = append(scanDates, scanDate)
				}
			}
		}

		posts = append(posts, map[string]interface{}{
			"postId": post.PostId,
			"name":   post.Name,
			"children": []map[string]interface{}{{
				"title":   "Scan Dates",
				"content": scanDates,
			}},
		})
	}
	postData["content"] = posts

	detourQueue := JsonNodeDetour{
		Graph:      postData,
		DetourNode: "name",
	}

	detourQueue.JsonDetour()

	return postData
}

func NewPostPathResponse(post *model.Post) []map[string]interface{} {
	path := make([]interface{}, 0)
	timestampStart := 1554772579000
	for i := range post.Coordinates {
		path = append(path, Waypoint{
			Coordinates: []float64{post.Coordinates[i].Coordinates.X(), post.Coordinates[i].Coordinates.Y()},
			Timestamp:   timestampStart + i*10,
		})
	}

	result := make(map[string]interface{})
	result["waypoints"] = path
	trueResult := make([]map[string]interface{}, 0)
	trueResult = append(trueResult, result)

	return trueResult
}
