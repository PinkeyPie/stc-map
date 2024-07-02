package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"mime"
	"net/http"
	"regexp"
	"simpleServer/internal/taskstore"
	"strconv"
	"time"
)

type TaskServer struct {
	store *taskstore.TaskStore
}

func NewTaskServer() *TaskServer {
	store := taskstore.New()
	return &TaskServer{store}
}

func (ts *TaskServer) CreateTaskHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Handling task create at %s\n", req.URL.Path)

	type RequestTask struct {
		Text string    `json:"text"`
		Tags []string  `json:"tags"`
		Due  time.Time `json:"due"`
	}

	type ResponseId struct {
		Id int `json:"id"`
	}

	contentType := req.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediaType != "application/json" {
		http.Error(w, "expected json", http.StatusBadRequest)
		return
	}
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var rt RequestTask
	if err := dec.Decode(&rt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := ts.store.CreateTask(rt.Text, rt.Tags, rt.Due)
	renderJSON(w, req, id)
}

func PathValue(path string, key string) string {
	pattern := fmt.Sprintf("[?,&]%s=(\\s+)[&,\\b]")
	expr := regexp.MustCompile(pattern)
	if expr.MatchString(path) {
		match := expr.FindStringSubmatch(path)
		return match[0]
	} else {
		return ""
	}
}

func (ts *TaskServer) GetTaskHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling get task at %s\n", req.URL.Path)

	id, err := strconv.Atoi(PathValue(req.URL.Path, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	task, err := ts.store.GetTask(id)
	if err != nil {
		http.Error(w, "can't find task", http.StatusBadRequest)
		return
	}
	renderJSON(w, req, task)
}

func (ts *TaskServer) GetAllTasksHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling get all tasks at %s\n", req.URL.Path)

	allTasks := ts.store.GetAllTasks()
	renderJSON(w, req, allTasks)
}

func renderJSON(w http.ResponseWriter, req *http.Request, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (ts *TaskServer) DeleteTaskHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling delete task at %s\n", req.URL.Path)

	id, _ := strconv.Atoi(mux.Vars(req)["id"])
	err := ts.store.DeleteTask(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	type Response struct {
		Status string `json:"status"`
		Error  string `json:"error"`
	}
	response := &Response{"ok", ""}
	renderJSON(w, req, response)
}

func (ts *TaskServer) DeleteAllTasksHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("handling delete all tasks at %s\n", req.URL.Path)
	ts.store.DeleteAllTasks()
}

func (ts *TaskServer) TagHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling tasks by tag at %s\n", req.URL.Path)

	tag := mux.Vars(req)["tag"]
	tasks := ts.store.GetTaskByTag(tag)
	renderJSON(w, req, tasks)
}

func (ts *TaskServer) DueHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling tasks by due at %s\n", req.URL.Path)

	vars := mux.Vars(req)
	badRequestError := func() {
		http.Error(w, fmt.Sprintf("expect /due/<year>/<month>/<day>, got %v", req.URL.Path), http.StatusBadRequest)
	}

	year, _ := strconv.Atoi(vars["year"])
	month, _ := strconv.Atoi(vars["month"])
	if month < int(time.January) || month > int(time.December) {
		badRequestError()
		return
	}
	day, _ := strconv.Atoi(vars["day"])

	tasks := ts.store.GetTaskByDueDate(year, time.Month(month), day)
	renderJSON(w, req, tasks)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		next.ServeHTTP(writer, request)
		log.Println("%s %s %s", request.Method, request.RequestURI, time.Since(start))
	})
}

func main() {
	router := mux.NewRouter()
	router.StrictSlash(true)
	server := NewTaskServer()

	router.HandleFunc("/task/", server.CreateTaskHandler).Methods("POST")
	router.HandleFunc("/task/", server.GetAllTasksHandler).Methods("GET")
	router.HandleFunc("/task/", server.DeleteAllTasksHandler).Methods("DELETE")
	router.HandleFunc("/task/{id:[0-9]+}/", server.GetTaskHandler).Methods("GET")
	router.HandleFunc("/task/{id:[0-9]+}/", server.DeleteTaskHandler).Methods("DELETE")
	router.HandleFunc("/tag/{tag}/", server.TagHandler).Methods("GET")
	router.HandleFunc("/due/{year:[0-9]+}/{month:[0-9]+}/{day:[0-9]+}/", server.DueHandler).Methods("GET")

	log.Fatal(http.ListenAndServe("localhost:8080", router))
}
