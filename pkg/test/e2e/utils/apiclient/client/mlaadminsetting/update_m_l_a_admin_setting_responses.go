// Code generated by go-swagger; DO NOT EDIT.

package mlaadminsetting

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"k8c.io/kubermatic/v2/pkg/test/e2e/utils/apiclient/models"
)

// UpdateMLAAdminSettingReader is a Reader for the UpdateMLAAdminSetting structure.
type UpdateMLAAdminSettingReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *UpdateMLAAdminSettingReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewUpdateMLAAdminSettingOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewUpdateMLAAdminSettingUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewUpdateMLAAdminSettingForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		result := NewUpdateMLAAdminSettingDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewUpdateMLAAdminSettingOK creates a UpdateMLAAdminSettingOK with default headers values
func NewUpdateMLAAdminSettingOK() *UpdateMLAAdminSettingOK {
	return &UpdateMLAAdminSettingOK{}
}

/* UpdateMLAAdminSettingOK describes a response with status code 200, with default header values.

MLAAdminSetting
*/
type UpdateMLAAdminSettingOK struct {
	Payload *models.MLAAdminSetting
}

func (o *UpdateMLAAdminSettingOK) Error() string {
	return fmt.Sprintf("[PUT /api/v2/projects/{project_id}/clusters/{cluster_id}/mlaadminsetting][%d] updateMLAAdminSettingOK  %+v", 200, o.Payload)
}
func (o *UpdateMLAAdminSettingOK) GetPayload() *models.MLAAdminSetting {
	return o.Payload
}

func (o *UpdateMLAAdminSettingOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.MLAAdminSetting)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUpdateMLAAdminSettingUnauthorized creates a UpdateMLAAdminSettingUnauthorized with default headers values
func NewUpdateMLAAdminSettingUnauthorized() *UpdateMLAAdminSettingUnauthorized {
	return &UpdateMLAAdminSettingUnauthorized{}
}

/* UpdateMLAAdminSettingUnauthorized describes a response with status code 401, with default header values.

EmptyResponse is a empty response
*/
type UpdateMLAAdminSettingUnauthorized struct {
}

func (o *UpdateMLAAdminSettingUnauthorized) Error() string {
	return fmt.Sprintf("[PUT /api/v2/projects/{project_id}/clusters/{cluster_id}/mlaadminsetting][%d] updateMLAAdminSettingUnauthorized ", 401)
}

func (o *UpdateMLAAdminSettingUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewUpdateMLAAdminSettingForbidden creates a UpdateMLAAdminSettingForbidden with default headers values
func NewUpdateMLAAdminSettingForbidden() *UpdateMLAAdminSettingForbidden {
	return &UpdateMLAAdminSettingForbidden{}
}

/* UpdateMLAAdminSettingForbidden describes a response with status code 403, with default header values.

EmptyResponse is a empty response
*/
type UpdateMLAAdminSettingForbidden struct {
}

func (o *UpdateMLAAdminSettingForbidden) Error() string {
	return fmt.Sprintf("[PUT /api/v2/projects/{project_id}/clusters/{cluster_id}/mlaadminsetting][%d] updateMLAAdminSettingForbidden ", 403)
}

func (o *UpdateMLAAdminSettingForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewUpdateMLAAdminSettingDefault creates a UpdateMLAAdminSettingDefault with default headers values
func NewUpdateMLAAdminSettingDefault(code int) *UpdateMLAAdminSettingDefault {
	return &UpdateMLAAdminSettingDefault{
		_statusCode: code,
	}
}

/* UpdateMLAAdminSettingDefault describes a response with status code -1, with default header values.

errorResponse
*/
type UpdateMLAAdminSettingDefault struct {
	_statusCode int

	Payload *models.ErrorResponse
}

// Code gets the status code for the update m l a admin setting default response
func (o *UpdateMLAAdminSettingDefault) Code() int {
	return o._statusCode
}

func (o *UpdateMLAAdminSettingDefault) Error() string {
	return fmt.Sprintf("[PUT /api/v2/projects/{project_id}/clusters/{cluster_id}/mlaadminsetting][%d] updateMLAAdminSetting default  %+v", o._statusCode, o.Payload)
}
func (o *UpdateMLAAdminSettingDefault) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *UpdateMLAAdminSettingDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
