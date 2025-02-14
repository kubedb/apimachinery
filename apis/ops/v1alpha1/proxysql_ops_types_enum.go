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
	// ProxySQLOpsRequestTypeUpdateVersion is a ProxySQLOpsRequestType of type UpdateVersion.
	ProxySQLOpsRequestTypeUpdateVersion ProxySQLOpsRequestType = "UpdateVersion"
	// ProxySQLOpsRequestTypeHorizontalScaling is a ProxySQLOpsRequestType of type HorizontalScaling.
	ProxySQLOpsRequestTypeHorizontalScaling ProxySQLOpsRequestType = "HorizontalScaling"
	// ProxySQLOpsRequestTypeVerticalScaling is a ProxySQLOpsRequestType of type VerticalScaling.
	ProxySQLOpsRequestTypeVerticalScaling ProxySQLOpsRequestType = "VerticalScaling"
	// ProxySQLOpsRequestTypeRestart is a ProxySQLOpsRequestType of type Restart.
	ProxySQLOpsRequestTypeRestart ProxySQLOpsRequestType = "Restart"
	// ProxySQLOpsRequestTypeReconfigure is a ProxySQLOpsRequestType of type Reconfigure.
	ProxySQLOpsRequestTypeReconfigure ProxySQLOpsRequestType = "Reconfigure"
	// ProxySQLOpsRequestTypeReconfigureTLS is a ProxySQLOpsRequestType of type ReconfigureTLS.
	ProxySQLOpsRequestTypeReconfigureTLS ProxySQLOpsRequestType = "ReconfigureTLS"
	// ProxySQLOpsRequestTypeRotateAuth is a ProxySQLOpsRequestType of type RotateAuth.
	ProxySQLOpsRequestTypeRotateAuth ProxySQLOpsRequestType = "RotateAuth"
)

var ErrInvalidProxySQLOpsRequestType = fmt.Errorf("not a valid ProxySQLOpsRequestType, try [%s]", strings.Join(_ProxySQLOpsRequestTypeNames, ", "))

var _ProxySQLOpsRequestTypeNames = []string{
	string(ProxySQLOpsRequestTypeUpdateVersion),
	string(ProxySQLOpsRequestTypeHorizontalScaling),
	string(ProxySQLOpsRequestTypeVerticalScaling),
	string(ProxySQLOpsRequestTypeRestart),
	string(ProxySQLOpsRequestTypeReconfigure),
	string(ProxySQLOpsRequestTypeReconfigureTLS),
	string(ProxySQLOpsRequestTypeRotateAuth),
}

// ProxySQLOpsRequestTypeNames returns a list of possible string values of ProxySQLOpsRequestType.
func ProxySQLOpsRequestTypeNames() []string {
	tmp := make([]string, len(_ProxySQLOpsRequestTypeNames))
	copy(tmp, _ProxySQLOpsRequestTypeNames)
	return tmp
}

// ProxySQLOpsRequestTypeValues returns a list of the values for ProxySQLOpsRequestType
func ProxySQLOpsRequestTypeValues() []ProxySQLOpsRequestType {
	return []ProxySQLOpsRequestType{
		ProxySQLOpsRequestTypeUpdateVersion,
		ProxySQLOpsRequestTypeHorizontalScaling,
		ProxySQLOpsRequestTypeVerticalScaling,
		ProxySQLOpsRequestTypeRestart,
		ProxySQLOpsRequestTypeReconfigure,
		ProxySQLOpsRequestTypeReconfigureTLS,
		ProxySQLOpsRequestTypeRotateAuth,
	}
}

// String implements the Stringer interface.
func (x ProxySQLOpsRequestType) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x ProxySQLOpsRequestType) IsValid() bool {
	_, err := ParseProxySQLOpsRequestType(string(x))
	return err == nil
}

var _ProxySQLOpsRequestTypeValue = map[string]ProxySQLOpsRequestType{
	"UpdateVersion":     ProxySQLOpsRequestTypeUpdateVersion,
	"HorizontalScaling": ProxySQLOpsRequestTypeHorizontalScaling,
	"VerticalScaling":   ProxySQLOpsRequestTypeVerticalScaling,
	"Restart":           ProxySQLOpsRequestTypeRestart,
	"Reconfigure":       ProxySQLOpsRequestTypeReconfigure,
	"ReconfigureTLS":    ProxySQLOpsRequestTypeReconfigureTLS,
	"RotateAuth":        ProxySQLOpsRequestTypeRotateAuth,
}

// ParseProxySQLOpsRequestType attempts to convert a string to a ProxySQLOpsRequestType.
func ParseProxySQLOpsRequestType(name string) (ProxySQLOpsRequestType, error) {
	if x, ok := _ProxySQLOpsRequestTypeValue[name]; ok {
		return x, nil
	}
	return ProxySQLOpsRequestType(""), fmt.Errorf("%s is %w", name, ErrInvalidProxySQLOpsRequestType)
}

// MustParseProxySQLOpsRequestType converts a string to a ProxySQLOpsRequestType, and panics if is not valid.
func MustParseProxySQLOpsRequestType(name string) ProxySQLOpsRequestType {
	val, err := ParseProxySQLOpsRequestType(name)
	if err != nil {
		panic(err)
	}
	return val
}
