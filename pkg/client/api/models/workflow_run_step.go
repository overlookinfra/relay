// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// WorkflowRunStep An individual workflow run step
// swagger:model WorkflowRunStep
type WorkflowRunStep struct {

	// Time at which the step execution ended
	EndedAt string `json:"ended_at,omitempty"`

	// Container image on which step is executed
	// Required: true
	Image *string `json:"image"`

	// A user provided step name. Must be unique within the workflow definition
	// Required: true
	Name *string `json:"name"`

	// JSON representation of the step specification
	// Required: true
	Spec interface{} `json:"spec"`

	// Time at which step execution started
	StartedAt string `json:"started_at,omitempty"`

	// Workflow run step status
	// Required: true
	// Enum: [success failure in-progress pending]
	Status *string `json:"status"`
}

// Validate validates this workflow run step
func (m *WorkflowRunStep) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateImage(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateName(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateSpec(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStatus(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *WorkflowRunStep) validateImage(formats strfmt.Registry) error {

	if err := validate.Required("image", "body", m.Image); err != nil {
		return err
	}

	return nil
}

func (m *WorkflowRunStep) validateName(formats strfmt.Registry) error {

	if err := validate.Required("name", "body", m.Name); err != nil {
		return err
	}

	return nil
}

func (m *WorkflowRunStep) validateSpec(formats strfmt.Registry) error {

	if err := validate.Required("spec", "body", m.Spec); err != nil {
		return err
	}

	return nil
}

var workflowRunStepTypeStatusPropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["success","failure","in-progress","pending"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		workflowRunStepTypeStatusPropEnum = append(workflowRunStepTypeStatusPropEnum, v)
	}
}

const (

	// WorkflowRunStepStatusSuccess captures enum value "success"
	WorkflowRunStepStatusSuccess string = "success"

	// WorkflowRunStepStatusFailure captures enum value "failure"
	WorkflowRunStepStatusFailure string = "failure"

	// WorkflowRunStepStatusInProgress captures enum value "in-progress"
	WorkflowRunStepStatusInProgress string = "in-progress"

	// WorkflowRunStepStatusPending captures enum value "pending"
	WorkflowRunStepStatusPending string = "pending"
)

// prop value enum
func (m *WorkflowRunStep) validateStatusEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, workflowRunStepTypeStatusPropEnum); err != nil {
		return err
	}
	return nil
}

func (m *WorkflowRunStep) validateStatus(formats strfmt.Registry) error {

	if err := validate.Required("status", "body", m.Status); err != nil {
		return err
	}

	// value enum
	if err := m.validateStatusEnum("status", "body", *m.Status); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *WorkflowRunStep) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *WorkflowRunStep) UnmarshalBinary(b []byte) error {
	var res WorkflowRunStep
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}