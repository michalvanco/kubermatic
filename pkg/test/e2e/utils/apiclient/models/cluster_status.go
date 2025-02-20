// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ClusterStatus ClusterStatus defines the cluster status.
//
// swagger:model ClusterStatus
type ClusterStatus struct {

	// URL specifies the address at which the cluster is available
	URL string `json:"url,omitempty"`

	// external c c m migration
	ExternalCCMMigration ExternalCCMMigrationStatus `json:"externalCCMMigration,omitempty"`

	// version
	Version Semver `json:"version,omitempty"`
}

// Validate validates this cluster status
func (m *ClusterStatus) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateExternalCCMMigration(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateVersion(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ClusterStatus) validateExternalCCMMigration(formats strfmt.Registry) error {
	if swag.IsZero(m.ExternalCCMMigration) { // not required
		return nil
	}

	if err := m.ExternalCCMMigration.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("externalCCMMigration")
		}
		return err
	}

	return nil
}

func (m *ClusterStatus) validateVersion(formats strfmt.Registry) error {
	if swag.IsZero(m.Version) { // not required
		return nil
	}

	if err := m.Version.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("version")
		}
		return err
	}

	return nil
}

// ContextValidate validate this cluster status based on the context it is used
func (m *ClusterStatus) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateExternalCCMMigration(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateVersion(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ClusterStatus) contextValidateExternalCCMMigration(ctx context.Context, formats strfmt.Registry) error {

	if err := m.ExternalCCMMigration.ContextValidate(ctx, formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("externalCCMMigration")
		}
		return err
	}

	return nil
}

func (m *ClusterStatus) contextValidateVersion(ctx context.Context, formats strfmt.Registry) error {

	if err := m.Version.ContextValidate(ctx, formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("version")
		}
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *ClusterStatus) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ClusterStatus) UnmarshalBinary(b []byte) error {
	var res ClusterStatus
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
