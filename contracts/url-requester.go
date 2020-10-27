package contracts

import "github.com/mgorunuch/test-task-affise/models"

type UrlDownloadRequest []string

type UrlDownloadResponse struct {
	Responses *models.UrlResponses `json:"responses,omitempty"`
	Error     *string              `json:"error,omitempty"`
}
