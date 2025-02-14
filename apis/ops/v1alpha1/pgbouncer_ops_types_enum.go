// Code generated by go-enum DO NOT EDIT.
// Version:
// Revision:
// Build Date:
// Built By:

package v1alpha1

import (
	"fmt"
	"strings"
)

const (
	// PgBouncerOpsRequestTypeHorizontalScaling is a PgBouncerOpsRequestType of type HorizontalScaling.
	PgBouncerOpsRequestTypeHorizontalScaling PgBouncerOpsRequestType = "HorizontalScaling"
	// PgBouncerOpsRequestTypeVerticalScaling is a PgBouncerOpsRequestType of type VerticalScaling.
	PgBouncerOpsRequestTypeVerticalScaling PgBouncerOpsRequestType = "VerticalScaling"
	// PgBouncerOpsRequestTypeUpdateVersion is a PgBouncerOpsRequestType of type UpdateVersion.
	PgBouncerOpsRequestTypeUpdateVersion PgBouncerOpsRequestType = "UpdateVersion"
	// PgBouncerOpsRequestTypeReconfigure is a PgBouncerOpsRequestType of type Reconfigure.
	PgBouncerOpsRequestTypeReconfigure PgBouncerOpsRequestType = "Reconfigure"
	// PgBouncerOpsRequestTypeRotateAuth is a PgBouncerOpsRequestType of type RotateAuth.
	PgBouncerOpsRequestTypeRotateAuth PgBouncerOpsRequestType = "RotateAuth"
	// PgBouncerOpsRequestTypeRestart is a PgBouncerOpsRequestType of type Restart.
	PgBouncerOpsRequestTypeRestart PgBouncerOpsRequestType = "Restart"
	// PgBouncerOpsRequestTypeReconfigureTLS is a PgBouncerOpsRequestType of type ReconfigureTLS.
	PgBouncerOpsRequestTypeReconfigureTLS PgBouncerOpsRequestType = "ReconfigureTLS"
)

var ErrInvalidPgBouncerOpsRequestType = fmt.Errorf("not a valid PgBouncerOpsRequestType, try [%s]", strings.Join(_PgBouncerOpsRequestTypeNames, ", "))

var _PgBouncerOpsRequestTypeNames = []string{
	string(PgBouncerOpsRequestTypeHorizontalScaling),
	string(PgBouncerOpsRequestTypeVerticalScaling),
	string(PgBouncerOpsRequestTypeUpdateVersion),
	string(PgBouncerOpsRequestTypeReconfigure),
	string(PgBouncerOpsRequestTypeRotateAuth),
	string(PgBouncerOpsRequestTypeRestart),
	string(PgBouncerOpsRequestTypeReconfigureTLS),
}

// PgBouncerOpsRequestTypeNames returns a list of possible string values of PgBouncerOpsRequestType.
func PgBouncerOpsRequestTypeNames() []string {
	tmp := make([]string, len(_PgBouncerOpsRequestTypeNames))
	copy(tmp, _PgBouncerOpsRequestTypeNames)
	return tmp
}

// PgBouncerOpsRequestTypeValues returns a list of the values for PgBouncerOpsRequestType
func PgBouncerOpsRequestTypeValues() []PgBouncerOpsRequestType {
	return []PgBouncerOpsRequestType{
		PgBouncerOpsRequestTypeHorizontalScaling,
		PgBouncerOpsRequestTypeVerticalScaling,
		PgBouncerOpsRequestTypeUpdateVersion,
		PgBouncerOpsRequestTypeReconfigure,
		PgBouncerOpsRequestTypeRotateAuth,
		PgBouncerOpsRequestTypeRestart,
		PgBouncerOpsRequestTypeReconfigureTLS,
	}
}

// String implements the Stringer interface.
func (x PgBouncerOpsRequestType) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x PgBouncerOpsRequestType) IsValid() bool {
	_, err := ParsePgBouncerOpsRequestType(string(x))
	return err == nil
}

var _PgBouncerOpsRequestTypeValue = map[string]PgBouncerOpsRequestType{
	"HorizontalScaling": PgBouncerOpsRequestTypeHorizontalScaling,
	"VerticalScaling":   PgBouncerOpsRequestTypeVerticalScaling,
	"UpdateVersion":     PgBouncerOpsRequestTypeUpdateVersion,
	"Reconfigure":       PgBouncerOpsRequestTypeReconfigure,
	"RotateAuth":        PgBouncerOpsRequestTypeRotateAuth,
	"Restart":           PgBouncerOpsRequestTypeRestart,
	"ReconfigureTLS":    PgBouncerOpsRequestTypeReconfigureTLS,
}

// ParsePgBouncerOpsRequestType attempts to convert a string to a PgBouncerOpsRequestType.
func ParsePgBouncerOpsRequestType(name string) (PgBouncerOpsRequestType, error) {
	if x, ok := _PgBouncerOpsRequestTypeValue[name]; ok {
		return x, nil
	}
	return PgBouncerOpsRequestType(""), fmt.Errorf("%s is %w", name, ErrInvalidPgBouncerOpsRequestType)
}

// MustParsePgBouncerOpsRequestType converts a string to a PgBouncerOpsRequestType, and panics if is not valid.
func MustParsePgBouncerOpsRequestType(name string) PgBouncerOpsRequestType {
	val, err := ParsePgBouncerOpsRequestType(name)
	if err != nil {
		panic(err)
	}
	return val
}
