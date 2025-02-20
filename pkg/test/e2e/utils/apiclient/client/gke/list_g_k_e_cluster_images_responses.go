// Code generated by go-swagger; DO NOT EDIT.

package gke

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"k8c.io/kubermatic/v2/pkg/test/e2e/utils/apiclient/models"
)

// ListGKEClusterImagesReader is a Reader for the ListGKEClusterImages structure.
type ListGKEClusterImagesReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListGKEClusterImagesReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListGKEClusterImagesOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewListGKEClusterImagesUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewListGKEClusterImagesForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		result := NewListGKEClusterImagesDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewListGKEClusterImagesOK creates a ListGKEClusterImagesOK with default headers values
func NewListGKEClusterImagesOK() *ListGKEClusterImagesOK {
	return &ListGKEClusterImagesOK{}
}

/* ListGKEClusterImagesOK describes a response with status code 200, with default header values.

GKEImageList
*/
type ListGKEClusterImagesOK struct {
	Payload models.GKEImageList
}

func (o *ListGKEClusterImagesOK) Error() string {
	return fmt.Sprintf("[GET /api/v2/projects/{project_id}/kubernetes/clusters/{cluster_id}/providers/gke/images][%d] listGKEClusterImagesOK  %+v", 200, o.Payload)
}
func (o *ListGKEClusterImagesOK) GetPayload() models.GKEImageList {
	return o.Payload
}

func (o *ListGKEClusterImagesOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListGKEClusterImagesUnauthorized creates a ListGKEClusterImagesUnauthorized with default headers values
func NewListGKEClusterImagesUnauthorized() *ListGKEClusterImagesUnauthorized {
	return &ListGKEClusterImagesUnauthorized{}
}

/* ListGKEClusterImagesUnauthorized describes a response with status code 401, with default header values.

EmptyResponse is a empty response
*/
type ListGKEClusterImagesUnauthorized struct {
}

func (o *ListGKEClusterImagesUnauthorized) Error() string {
	return fmt.Sprintf("[GET /api/v2/projects/{project_id}/kubernetes/clusters/{cluster_id}/providers/gke/images][%d] listGKEClusterImagesUnauthorized ", 401)
}

func (o *ListGKEClusterImagesUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewListGKEClusterImagesForbidden creates a ListGKEClusterImagesForbidden with default headers values
func NewListGKEClusterImagesForbidden() *ListGKEClusterImagesForbidden {
	return &ListGKEClusterImagesForbidden{}
}

/* ListGKEClusterImagesForbidden describes a response with status code 403, with default header values.

EmptyResponse is a empty response
*/
type ListGKEClusterImagesForbidden struct {
}

func (o *ListGKEClusterImagesForbidden) Error() string {
	return fmt.Sprintf("[GET /api/v2/projects/{project_id}/kubernetes/clusters/{cluster_id}/providers/gke/images][%d] listGKEClusterImagesForbidden ", 403)
}

func (o *ListGKEClusterImagesForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewListGKEClusterImagesDefault creates a ListGKEClusterImagesDefault with default headers values
func NewListGKEClusterImagesDefault(code int) *ListGKEClusterImagesDefault {
	return &ListGKEClusterImagesDefault{
		_statusCode: code,
	}
}

/* ListGKEClusterImagesDefault describes a response with status code -1, with default header values.

errorResponse
*/
type ListGKEClusterImagesDefault struct {
	_statusCode int

	Payload *models.ErrorResponse
}

// Code gets the status code for the list g k e cluster images default response
func (o *ListGKEClusterImagesDefault) Code() int {
	return o._statusCode
}

func (o *ListGKEClusterImagesDefault) Error() string {
	return fmt.Sprintf("[GET /api/v2/projects/{project_id}/kubernetes/clusters/{cluster_id}/providers/gke/images][%d] listGKEClusterImages default  %+v", o._statusCode, o.Payload)
}
func (o *ListGKEClusterImagesDefault) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *ListGKEClusterImagesDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
