package router

import (
	"encoding/json"
	"fmt"
	"github.com/mgorunuch/test-task-affise/contracts"
	"github.com/mgorunuch/test-task-affise/models"
	"github.com/mgorunuch/test-task-affise/service"
	"github.com/mgorunuch/test-task-affise/utils"
	"net/http"
)

func New(r *utils.RateLimiter, d service.UrlDownloader) ServerHandler {
	return ServerHandler{
		rateLimiter:   r,
		urlDownloader: d,
	}
}

type ServerHandler struct {
	rateLimiter   *utils.RateLimiter
	urlDownloader service.UrlDownloader
}

func sendError(resp http.ResponseWriter, errMsg string, code int) {
	data := contracts.UrlDownloadResponse{
		Error: &errMsg,
	}

	dataBody, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("failed to send error message %s", errMsg)

		return
	}

	sendJSON(resp, dataBody, code)
}

func sendSuccess(resp http.ResponseWriter, responses models.UrlResponses, code int) {
	data := contracts.UrlDownloadResponse{
		Responses: &responses,
	}

	dataBody, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("failed to send success message %s", responses)
		sendError(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	sendJSON(resp, dataBody, code)
}

func sendJSON(resp http.ResponseWriter, body []byte, code int) {
	resp.Header().Add("content-type", "application/json")

	resp.WriteHeader(code)

	_, _ = resp.Write(body)
}

func (s ServerHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method != http.MethodPost {
		sendError(resp, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	body := contracts.UrlDownloadRequest{}

	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		sendError(resp, "unprocessable entity", http.StatusUnprocessableEntity)

		return
	}

	urlsCount := len(body)
	if urlsCount > 20 || urlsCount == 0 {
		sendError(resp, "urls count must be in range of 1 and 20", http.StatusUnprocessableEntity)

		return
	}

	ok := s.rateLimiter.RegisterConnection()
	if !ok {
		sendError(resp, "max number of requests gained", http.StatusTeapot)

		return
	}

	data, err := s.urlDownloader.Download(ctx, body)
	if err != nil {
		sendError(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	respData := models.UrlResponses{}
	for _, d := range data {
		respData[d.URL] = d.Response
	}

	sendSuccess(resp, respData, http.StatusOK)

	defer func() {
		s.rateLimiter.FinishConnection()
	}()
}
