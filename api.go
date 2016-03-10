package gj

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/yoru9zine/gj/pkg/id"
)

var (
	respBadRequest    = APIResponseModel{Msg: "bad request"}
	respNotFound      = APIResponseModel{Msg: "not found"}
	respDuplicated    = APIResponseModel{Msg: "duplicated"}
	respInternalError = APIResponseModel{Msg: "internal error"}
	respOK            = APIResponseModel{Msg: "ok"}
)

type APIServer struct {
	logDir string
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
	b, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("failed to read body: %s", err)
		c.JSON(http.StatusInternalServerError, respInternalError)
	}
	var pvm ProcessViewModel
	if err := json.Unmarshal(b, &pvm); err != nil {
		log.Printf("failed to parse json: %s\n`%s`", err, b)
		c.JSON(http.StatusBadRequest, respBadRequest)
		return
	}
	id := id.New()
	pvm.ID = id
	a.Procs[pvm.ID] = pvm.Process()
	c.IndentedJSON(http.StatusOK, APIResponseCreateProc{respOK, id})
	return
}

func (a *APIServer) ShowProc(c *gin.Context) {
	pid := c.Param("pid")
	proc, apierr := a.findProcess(pid)
	if apierr != nil {
		c.IndentedJSON(apierr.Status, apierr.Model)
		return
	}
	c.IndentedJSON(http.StatusOK, APIResponseShowProc{respOK, proc.ViewModel()})
	return
}

func (a *APIServer) StartProc(c *gin.Context) {
	pid := c.Param("pid")
	proc, apierr := a.findProcess(pid)
	if apierr != nil {
		c.IndentedJSON(apierr.Status, apierr.Model)
		return
	}
	logDir := fmt.Sprintf("%s/%s", a.logDir, pid)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("failed to create logdir `%s`: %s", logDir, err)
		c.IndentedJSON(http.StatusInternalServerError, respInternalError)
	}
	if err := proc.Start(logDir); err != nil {
		log.Printf("error at starting process: %s", err)
		c.String(http.StatusInternalServerError, "ng")
	}
	c.String(http.StatusOK, "ok")
}

func (a *APIServer) ShowProcLog(c *gin.Context) {
	pid := c.Param("pid")
	proc, apierr := a.findProcess(pid)
	if apierr != nil {
		c.IndentedJSON(apierr.Status, apierr.Model)
		return
	}
	c.String(http.StatusOK, proc.output.String())
}

func (a *APIServer) findProcess(pid string) (*Process, *APIError) {
	proc, err := a.Procs.Find(pid)
	if err != nil {
		switch err {
		case ErrNotUniq:
			return nil, &APIError{http.StatusConflict, respDuplicated}
		case ErrProcessNotFound:
			return nil, &APIError{http.StatusNotFound, respNotFound}
		default:
			log.Printf("failed to find process for %s: %s\n", pid, err)
			return nil, &APIError{http.StatusInternalServerError, respInternalError}
		}
	}
	return proc, nil
}

func NewAPIServer(logDir string) *APIServer {
	s := &APIServer{
		logDir: strings.TrimSuffix(logDir, "/"),
		Engine: gin.Default(),
		Procs:  map[string]*Process{},
	}
	s.Setup()
	return s
}

type APIResponseModel struct {
	Msg string `json:"message"`
}
type APIResponseCreateProc struct {
	APIResponseModel
	PID string `json:"pid"`
}
type APIResponseShowProc struct {
	APIResponseModel
	Proc *ProcessViewModel `json:"proc"`
}
type APIResponseShowProcs struct {
	APIResponseModel
	Procs map[string]*ProcessViewModel `json:"procs"`
}

type APIError struct {
	Status int
	Model  APIResponseModel
}

func (a APIError) Error() string {
	return fmt.Sprintf("status: %d, data: %+v", a.Status, a.Model)
}
