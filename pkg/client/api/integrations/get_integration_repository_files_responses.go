// Code generated by go-swagger; DO NOT EDIT.

package integrations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/puppetlabs/nebula-cli/pkg/client/api/models"
)

// GetIntegrationRepositoryFilesReader is a Reader for the GetIntegrationRepositoryFiles structure.
type GetIntegrationRepositoryFilesReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetIntegrationRepositoryFilesReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetIntegrationRepositoryFilesOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewGetIntegrationRepositoryFilesDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetIntegrationRepositoryFilesOK creates a GetIntegrationRepositoryFilesOK with default headers values
func NewGetIntegrationRepositoryFilesOK() *GetIntegrationRepositoryFilesOK {
	return &GetIntegrationRepositoryFilesOK{}
}

/*GetIntegrationRepositoryFilesOK handles this case with default header values.

A list of the files matching the given criteria
*/
type GetIntegrationRepositoryFilesOK struct {
	Payload *GetIntegrationRepositoryFilesOKBody
}

func (o *GetIntegrationRepositoryFilesOK) Error() string {
	return fmt.Sprintf("[GET /api/integrations/{integrationId}/repositories/{integrationRepositoryOwner}/{integrationRepositoryName}/branches/{integrationRepositoryBranch}/files][%d] getIntegrationRepositoryFilesOK  %+v", 200, o.Payload)
}

func (o *GetIntegrationRepositoryFilesOK) GetPayload() *GetIntegrationRepositoryFilesOKBody {
	return o.Payload
}

func (o *GetIntegrationRepositoryFilesOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(GetIntegrationRepositoryFilesOKBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetIntegrationRepositoryFilesDefault creates a GetIntegrationRepositoryFilesDefault with default headers values
func NewGetIntegrationRepositoryFilesDefault(code int) *GetIntegrationRepositoryFilesDefault {
	return &GetIntegrationRepositoryFilesDefault{
		_statusCode: code,
	}
}

/*GetIntegrationRepositoryFilesDefault handles this case with default header values.

An error occurred
*/
type GetIntegrationRepositoryFilesDefault struct {
	_statusCode int

	Payload *GetIntegrationRepositoryFilesDefaultBody
}

// Code gets the status code for the get integration repository files default response
func (o *GetIntegrationRepositoryFilesDefault) Code() int {
	return o._statusCode
}

func (o *GetIntegrationRepositoryFilesDefault) Error() string {
	return fmt.Sprintf("[GET /api/integrations/{integrationId}/repositories/{integrationRepositoryOwner}/{integrationRepositoryName}/branches/{integrationRepositoryBranch}/files][%d] getIntegrationRepositoryFiles default  %+v", o._statusCode, o.Payload)
}

func (o *GetIntegrationRepositoryFilesDefault) GetPayload() *GetIntegrationRepositoryFilesDefaultBody {
	return o.Payload
}

func (o *GetIntegrationRepositoryFilesDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(GetIntegrationRepositoryFilesDefaultBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

/*GetIntegrationRepositoryFilesDefaultBody Error response
swagger:model GetIntegrationRepositoryFilesDefaultBody
*/
type GetIntegrationRepositoryFilesDefaultBody struct {

	// error
	Error *models.Error `json:"error,omitempty"`
}

// Validate validates this get integration repository files default body
func (o *GetIntegrationRepositoryFilesDefaultBody) Validate(formats strfmt.Registry) error {
	var res []error

	if err := o.validateError(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *GetIntegrationRepositoryFilesDefaultBody) validateError(formats strfmt.Registry) error {

	if swag.IsZero(o.Error) { // not required
		return nil
	}

	if o.Error != nil {
		if err := o.Error.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("getIntegrationRepositoryFiles default" + "." + "error")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (o *GetIntegrationRepositoryFilesDefaultBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *GetIntegrationRepositoryFilesDefaultBody) UnmarshalBinary(b []byte) error {
	var res GetIntegrationRepositoryFilesDefaultBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*GetIntegrationRepositoryFilesOKBody The response type for listing files
swagger:model GetIntegrationRepositoryFilesOKBody
*/
type GetIntegrationRepositoryFilesOKBody struct {

	// A list of files and directories
	Files []*models.RepositoryFile `json:"files"`
}

// Validate validates this get integration repository files o k body
func (o *GetIntegrationRepositoryFilesOKBody) Validate(formats strfmt.Registry) error {
	var res []error

	if err := o.validateFiles(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *GetIntegrationRepositoryFilesOKBody) validateFiles(formats strfmt.Registry) error {

	if swag.IsZero(o.Files) { // not required
		return nil
	}

	for i := 0; i < len(o.Files); i++ {
		if swag.IsZero(o.Files[i]) { // not required
			continue
		}

		if o.Files[i] != nil {
			if err := o.Files[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("getIntegrationRepositoryFilesOK" + "." + "files" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (o *GetIntegrationRepositoryFilesOKBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *GetIntegrationRepositoryFilesOKBody) UnmarshalBinary(b []byte) error {
	var res GetIntegrationRepositoryFilesOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}