/*
 Copyright (C) 2022 alsritter

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as
 published by the Free Software Foundation, either version 3 of the
 License, or (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/caseprovider"
	"github.com/alsritter/middlebaby/pkg/protomanager"
	"github.com/alsritter/middlebaby/pkg/targetprocess"
	"github.com/alsritter/middlebaby/pkg/taskserver"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/handlers"
)

type status string

const (
	statusSuccess status = "success"
	statusError   status = "error"
)

type errorType string

const (
	errorNone        errorType = ""
	errorTimeout     errorType = "timeout"
	errorCanceled    errorType = "canceled"
	errorExec        errorType = "execution"
	errorBadData     errorType = "bad_data"
	errorInternal    errorType = "internal"
	errorUnavailable errorType = "unavailable"
	errorNotFound    errorType = "not_found"
)

type apiFuncResult struct {
	data      interface{}
	err       *apiError
	finalizer func()
}

type apiFunc func(r *http.Request) apiFuncResult

type apiError struct {
	typ errorType
	err error
}

func (e *apiError) Error() string {
	return fmt.Sprintf("%s: %s", e.typ, e.err)
}

type response struct {
	Status    status      `json:"status"`
	Data      interface{} `json:"data,omitempty"`
	ErrorType errorType   `json:"errorType,omitempty"`
	Error     string      `json:"error,omitempty"`
	Warnings  []string    `json:"warnings,omitempty"`
}

type API struct {
	logger.Logger

	apiProvider  apimanager.Provider
	caseProvider caseprovider.Provider
	protoManager protomanager.Provider
	taskService  taskserver.Provider
	target       targetprocess.Provider
}

// TODO: here cannot need all services providers.
func NewAPI(
	log logger.Logger,
	apiProvider apimanager.Provider,
	caseProvider caseprovider.Provider,
	protoManager protomanager.Provider,
	taskService taskserver.Provider,
	target targetprocess.Provider,
) *API {
	return &API{
		Logger:       log.NewLogger("v1"),
		apiProvider:  apiProvider,
		caseProvider: caseProvider,
		protoManager: protoManager,
		taskService:  taskService,
		target:       target,
	}
}

func (a *API) Register(r *gin.Engine) {
	wrap := func(f apiFunc) gin.HandlerFunc {
		hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			result := f(r)
			if result.finalizer != nil {
				defer result.finalizer()
			}

			if result.err != nil {
				a.respondError(w, result.err, result.data)
				return
			}

			if result.data != nil {
				a.respond(w, result.data)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})

		return func(c *gin.Context) {
			handlers.CompressHandler(hf).ServeHTTP(c.Writer, c.Request)
		}
	}
	v1 := r.Group("/v1")
	{
		v1.GET("/getCaseList", wrap(a.getCaseList))
		v1.POST("/runSingleCase", wrap(a.runSingleCase))
	}
}

func (a *API) getCaseList(r *http.Request) (result apiFuncResult) {
	allCase := a.caseProvider.GetAllItfWithFileInfo()
	return apiFuncResult{allCase, nil, nil}
}

func (a *API) runSingleCase(r *http.Request) (result apiFuncResult) {
	itfName := r.FormValue("itfName")
	if itfName == "" {
		return apiFuncResult{nil, &apiError{errorBadData, errors.New("itfName is required")}, nil}
	}
	caseName := r.FormValue("caseName")
	if caseName == "" {
		return apiFuncResult{nil, &apiError{errorBadData, errors.New("caseName is required")}, nil}
	}

	res, err := a.taskService.RunSingleTaskCase(r.Context(), itfName, caseName)
	if err != nil {
		return apiFuncResult{nil, &apiError{errorBadData, err}, nil}
	}

	return apiFuncResult{res, nil, nil}
}

func (api *API) respond(w http.ResponseWriter, data interface{}) {
	statusMessage := statusSuccess
	b, err := json.Marshal(&response{
		Status: statusMessage,
		Data:   data,
	})
	if err != nil {
		api.Error(nil, "error marshaling json response [%v]", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if n, err := w.Write(b); err != nil {
		api.Error(nil, "error writing response err [%v], bytesWritten [%d]", err, n)
	}
}

func (api *API) respondError(w http.ResponseWriter, apiErr *apiError, data interface{}) {
	b, err := json.Marshal(&response{
		Status:    statusError,
		ErrorType: apiErr.typ,
		Error:     apiErr.err.Error(),
		Data:      data,
	})

	if err != nil {
		api.Error(nil, "error marshaling json response err [%v]", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var code int
	switch apiErr.typ {
	case errorBadData:
		code = http.StatusBadRequest
	case errorExec:
		code = http.StatusUnprocessableEntity
	case errorCanceled, errorTimeout:
		code = http.StatusServiceUnavailable
	case errorInternal:
		code = http.StatusInternalServerError
	case errorNotFound:
		code = http.StatusNotFound
	default:
		code = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if n, err := w.Write(b); err != nil {
		api.Error(nil, "error writing response err [%v], bytesWritten [%d]", err, n)
	}
}
