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
	// SolrOpsRequestTypeUpdateVersion is a SolrOpsRequestType of type UpdateVersion.
	SolrOpsRequestTypeUpdateVersion SolrOpsRequestType = "UpdateVersion"
	// SolrOpsRequestTypeVerticalScaling is a SolrOpsRequestType of type VerticalScaling.
	SolrOpsRequestTypeVerticalScaling SolrOpsRequestType = "VerticalScaling"
	// SolrOpsRequestTypeVolumeExpansion is a SolrOpsRequestType of type VolumeExpansion.
	SolrOpsRequestTypeVolumeExpansion SolrOpsRequestType = "VolumeExpansion"
	// SolrOpsRequestTypeRestart is a SolrOpsRequestType of type Restart.
	SolrOpsRequestTypeRestart SolrOpsRequestType = "Restart"
	// SolrOpsRequestTypeReconfigure is a SolrOpsRequestType of type Reconfigure.
	SolrOpsRequestTypeReconfigure SolrOpsRequestType = "Reconfigure"
)

var ErrInvalidSolrOpsRequestType = fmt.Errorf("not a valid SolrOpsRequestType, try [%s]", strings.Join(_SolrOpsRequestTypeNames, ", "))

var _SolrOpsRequestTypeNames = []string{
	string(SolrOpsRequestTypeUpdateVersion),
	string(SolrOpsRequestTypeVerticalScaling),
	string(SolrOpsRequestTypeVolumeExpansion),
	string(SolrOpsRequestTypeRestart),
	string(SolrOpsRequestTypeReconfigure),
}

// SolrOpsRequestTypeNames returns a list of possible string values of SolrOpsRequestType.
func SolrOpsRequestTypeNames() []string {
	tmp := make([]string, len(_SolrOpsRequestTypeNames))
	copy(tmp, _SolrOpsRequestTypeNames)
	return tmp
}

// SolrOpsRequestTypeValues returns a list of the values for SolrOpsRequestType
func SolrOpsRequestTypeValues() []SolrOpsRequestType {
	return []SolrOpsRequestType{
		SolrOpsRequestTypeUpdateVersion,
		SolrOpsRequestTypeVerticalScaling,
		SolrOpsRequestTypeVolumeExpansion,
		SolrOpsRequestTypeRestart,
		SolrOpsRequestTypeReconfigure,
	}
}

// String implements the Stringer interface.
func (x SolrOpsRequestType) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x SolrOpsRequestType) IsValid() bool {
	_, err := ParseSolrOpsRequestType(string(x))
	return err == nil
}

var _SolrOpsRequestTypeValue = map[string]SolrOpsRequestType{
	"UpdateVersion":   SolrOpsRequestTypeUpdateVersion,
	"VerticalScaling": SolrOpsRequestTypeVerticalScaling,
	"VolumeExpansion": SolrOpsRequestTypeVolumeExpansion,
	"Restart":         SolrOpsRequestTypeRestart,
	"Reconfigure":     SolrOpsRequestTypeReconfigure,
}

// ParseSolrOpsRequestType attempts to convert a string to a SolrOpsRequestType.
func ParseSolrOpsRequestType(name string) (SolrOpsRequestType, error) {
	if x, ok := _SolrOpsRequestTypeValue[name]; ok {
		return x, nil
	}
	return SolrOpsRequestType(""), fmt.Errorf("%s is %w", name, ErrInvalidSolrOpsRequestType)
}

// MustParseSolrOpsRequestType converts a string to a SolrOpsRequestType, and panics if is not valid.
func MustParseSolrOpsRequestType(name string) SolrOpsRequestType {
	val, err := ParseSolrOpsRequestType(name)
	if err != nil {
		panic(err)
	}
	return val
}
