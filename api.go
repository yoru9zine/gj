package gj

import (
	"net/http"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/pborman/uuid"
)

var (
	respBadRequest = APIResponseModel{Msg: "bad request"}
	respOK         = APIResponseModel{Msg: "ok"}
)

type APIServer struct {
	*gin.Engine
	Procs Processes
}

func (a *APIServer) Setup() {
	a.GET("/api/v1/procs", a.ShowProcs)
	a.POST("/api/v1/procs", a.CreateProc)
	a.GET("/api/v1/procs/:pid", a.ShowProc)
	a.GET("/api/v1/procs/:pid/start", a.StartProc)
	a.GET("/api/v1/procs/:pid/log", a.ShowProcLog)
}

func (a *APIServer) ShowProcs(c *gin.Context) {
	models := a.Procs.ViewModels()
	resp := APIResponseShowProcs{respOK, models}
	c.IndentedJSON(http.StatusOK, resp)
}

func (a *APIServer) CreateProc(c *gin.Context) {
	var pvm ProcessViewModel
	if err := c.BindJSON(&pvm); err != nil {
		log.Printf("failed to parse json: %s", err)
		c.JSON(http.StatusBadRequest, respBadRequest)
		return
	}
	id := uuid.NewUUID()
	pvm.ID = id.String()
	a.Procs[pvm.ID] = pvm.Process()
	c.IndentedJSON(http.StatusOK, respOK)
	return
}

func (a *APIServer) ShowProc(c *gin.Context) {
	pid := c.Param("pid")
	proc := a.Procs[pid]
	c.IndentedJSON(http.StatusOK, APIResponseShowProc{respOK, proc.ViewModel()})
	return
}

func (a *APIServer) StartProc(c *gin.Context) {
	pid := c.Param("pid")
	err := a.Procs[pid].Start()
	if err != nil {
		log.Printf("error at starting process: %s", err)
		c.String(http.StatusInternalServerError, "ng")
	}
	c.String(http.StatusOK, "ok")
}

func (a *APIServer) ShowProcLog(c *gin.Context) {
	pid := c.Param("pid")
	proc := a.Procs[pid]
	c.String(http.StatusOK, proc.output.String())
}

func NewAPIServer() *APIServer {
	s := &APIServer{
		Engine: gin.Default(),
		Procs:  map[string]*Process{},
	}
	s.Setup()
	return s
}

type APIResponseModel struct {
	Msg string `json:"message"`
}

type APIResponseShowProc struct {
	APIResponseModel
	Proc *ProcessViewModel `json:"proc"`
}
type APIResponseShowProcs struct {
	APIResponseModel
	Procs map[string]*ProcessViewModel `json:"procs"`
}
