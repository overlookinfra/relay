package client

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/puppetlabs/leg/encoding/transfer"
	"github.com/puppetlabs/relay/pkg/debug"
	"github.com/puppetlabs/relay/pkg/errors"
	"github.com/puppetlabs/relay/pkg/model"
)

type ListWorkflowSecretsResponse struct {
	WorkflowSecrets []model.WorkflowSecretSummary `json:"secrets"`
}

func (c *Client) ListWorkflowSecrets(workflow string) (*ListWorkflowSecretsResponse, errors.Error) {
	resp := &ListWorkflowSecretsResponse{}

	if err := c.Request(
		WithPath(fmt.Sprintf("/api/workflows/%v/secrets", workflow)),
		WithResponseInto(&resp)); err != nil {
		return nil, err
	}

	return resp, nil
}

type CreateWorkflowSecretParameters struct {
	Name  string                 `json:"name"`
	Value transfer.JSONInterface `json:"value"`
}

func (c *Client) CreateWorkflowSecret(workflow, secret, value string) (*model.WorkflowSecretEntity, errors.Error) {
	params := &CreateWorkflowSecretParameters{
		Name:  secret,
		Value: transfer.JSONInterface{Data: value},
	}

	response := &model.WorkflowSecretEntity{}

	if err := c.Request(
		WithMethod(http.MethodPost),
		WithPath(fmt.Sprintf("/api/workflows/%v/secrets", workflow)),
		WithBody(params),
		WithResponseInto(response),
	); err != nil {
		return nil, err
	}

	return response, nil
}

type UpdateWorkflowSecretParameters struct {
	Value transfer.JSONInterface `json:"value"`
}

func (c *Client) UpdateWorkflowSecret(workflow, secret, value string) (*model.WorkflowSecretEntity, errors.Error) {
	params := &UpdateWorkflowSecretParameters{
		Value: transfer.JSONInterface{Data: value},
	}

	response := &model.WorkflowSecretEntity{}

	if err := c.Request(
		WithMethod(http.MethodPut),
		WithPath(fmt.Sprintf("/api/workflows/%v/secrets/%v", workflow, secret)),
		WithBody(params),
		WithResponseInto(response),
	); err != nil {
		return nil, err
	}

	return response, nil
}

type DeleteWorkflowSecretResponse struct {
	Success    bool   `json:"success"`
	ResourceId string `json:"resource_id"`
}

func (c *Client) DeleteWorkflowSecret(workflow, secret string) (*DeleteWorkflowSecretResponse, errors.Error) {
	response := &DeleteWorkflowSecretResponse{}

	if err := c.Request(
		WithMethod(http.MethodDelete),
		WithPath(fmt.Sprintf("/api/workflows/%v/secrets/%v", workflow, secret)),
		WithResponseInto(response),
	); err != nil {
		return nil, err
	}

	return response, nil
}

type CreateWorkflowParameters struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *Client) CreateWorkflow(name string) (*model.WorkflowEntity, errors.Error) {
	params := &CreateWorkflowParameters{
		Name:        name,
		Description: "",
	}

	response := &model.WorkflowEntity{}

	if err := c.Request(
		WithMethod(http.MethodPost),
		WithPath("/api/workflows"),
		WithBody(params),
		WithResponseInto(response),
	); err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) GetWorkflow(name string) (*model.WorkflowEntity, errors.Error) {
	response := &model.WorkflowEntity{}

	if err := c.Request(
		WithPath(fmt.Sprintf("/api/workflows/%v", name)),
		WithResponseInto(response),
	); err != nil {
		return nil, err
	}

	return response, nil
}

type DeleteWorkflowResponse struct {
	Success    bool   `json:"success"`
	ResourceId string `json:"resource_id"`
}

func (c *Client) DeleteWorkflow(name string) (*DeleteWorkflowResponse, errors.Error) {
	response := &DeleteWorkflowResponse{}

	if err := c.Request(
		WithMethod(http.MethodDelete),
		WithPath(fmt.Sprintf("/api/workflows/%v", name)),
		WithResponseInto(response),
	); err != nil {
		return nil, err
	}

	return response, nil
}

type RunWorkflowParameterValueRequest struct {
	Value string `json:"value"`
}

type RunWorkflowRequest struct {
	Parameters map[string]RunWorkflowParameterValueRequest `json:"parameters"`
}

type RunWorkflowWorkflowResponse struct {
	Name string `json:"name"`
}

type RunWorkflowParameterValueResponse struct {
	Value string `json:"value"`
}

type RunWorkflowRevisionResponse struct {
	Id string `json:"id"`
}

type RunWorkflowStateResponse struct {
	Status    string     `json:"status"`
	StartedAt *time.Time `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`

	// TODO: Add steps here, in case we really care about that.
}

type RunWorkflowRunResponse struct {
	CreatedAt  time.Time                                    `json:"created_at"`
	RunNumber  int                                          `json:"run_number"`
	Revision   RunWorkflowRevisionResponse                  `json:"revision"`
	State      RunWorkflowStateResponse                     `json:"state"`
	Parameters map[string]RunWorkflowParameterValueResponse `json:"parameters"`
	Workflow   RunWorkflowWorkflowResponse                  `json:"workflow"`
}

type RunWorkflowResponse struct {
	Run RunWorkflowRunResponse `json:"run"`
}

func setupParams(params map[string]string) map[string]RunWorkflowParameterValueRequest {
	res := make(map[string]RunWorkflowParameterValueRequest, len(params))

	for key, val := range params {
		res[key] = RunWorkflowParameterValueRequest{val}
	}

	return res
}

func (c *Client) RunWorkflow(name string, params map[string]string) (*RunWorkflowResponse, errors.Error) {
	req := &RunWorkflowRequest{
		Parameters: setupParams(params),
	}

	resp := &RunWorkflowResponse{}

	if err := c.Request(
		WithMethod(http.MethodPost),
		WithPath(fmt.Sprintf("/api/workflows/%v/runs", name)),
		WithBody(req),
		WithResponseInto(resp),
	); err != nil {
		return nil, err
	}

	return resp, nil
}

// DownloadWorkflow gets the latest configuration (as a YAML string) for a
// given workflow name.
func (c *Client) DownloadWorkflow(name string) (string, errors.Error) {
	rev, err := c.GetLatestRevision(name)
	if err != nil {
		return "", err
	}

	dec, berr := base64.StdEncoding.DecodeString(rev.Revision.Raw)

	if berr != nil {
		debug.Logf("the workflow body was in the wrong format. %s", berr.Error())
		return "", errors.NewClientUnknownError().WithCause(berr)
	}

	return string(dec), nil
}
