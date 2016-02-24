package gj

import (
	"net/http"

	"github.com/coreos/go-log/log"
	"github.com/gin-gonic/gin"
)

var (
	respBadRequest = APIResponseModel{Msg: "bad request"}
	respOK         = APIResponseModel{Msg: "ok"}
)

type APIServer struct {
	*gin.Engine
	Jobs Jobs
}

func (a *APIServer) Setup() {
	a.GET("/api/v1/jobs", a.ShowJobs)
	a.POST("/api/v1/jobs", a.CreateJob)
}

func (a *APIServer) ShowJobs(c *gin.Context) {
	models := a.Jobs.ViewModels()
	resp := APIResponseShowJobs{respOK, models}
	c.IndentedJSON(http.StatusOK, resp)
}

func (a *APIServer) CreateJob(c *gin.Context) {
	var jvm JobViewModel
	if err := c.BindJSON(&jvm); err != nil {
		log.Errorf("failed to parse json: %s", err)
		c.JSON(http.StatusBadRequest, respBadRequest)
		return
	}
	a.Jobs["id"] = jvm.Job()
	c.IndentedJSON(http.StatusOK, respOK)
	return
}

func NewAPIServer() *APIServer {
	s := &APIServer{
		Engine: gin.Default(),
		Jobs:   map[string]*Job{},
	}
	s.Setup()
	return s
}

type APIResponseModel struct {
	Msg string `json:"message"`
}

type APIResponseShowJobs struct {
	APIResponseModel
	Jobs map[string]*JobViewModel `json:"jobs"`
}

type JobViewModel struct {
	Name string `json:"name"`
}

func (j *JobViewModel) Job() *Job {
	return &Job{
		Name: j.Name,
	}
}
