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
	// CassandraOpsRequestTypeUpdateVersion is a CassandraOpsRequestType of type UpdateVersion.
	CassandraOpsRequestTypeUpdateVersion CassandraOpsRequestType = "UpdateVersion"
	// CassandraOpsRequestTypeVerticalScaling is a CassandraOpsRequestType of type VerticalScaling.
	CassandraOpsRequestTypeVerticalScaling CassandraOpsRequestType = "VerticalScaling"
	// CassandraOpsRequestTypeRestart is a CassandraOpsRequestType of type Restart.
	CassandraOpsRequestTypeRestart CassandraOpsRequestType = "Restart"
)

var ErrInvalidCassandraOpsRequestType = fmt.Errorf("not a valid CassandraOpsRequestType, try [%s]", strings.Join(_CassandraOpsRequestTypeNames, ", "))

var _CassandraOpsRequestTypeNames = []string{
	string(CassandraOpsRequestTypeUpdateVersion),
	string(CassandraOpsRequestTypeVerticalScaling),
	string(CassandraOpsRequestTypeRestart),
}

// CassandraOpsRequestTypeNames returns a list of possible string values of CassandraOpsRequestType.
func CassandraOpsRequestTypeNames() []string {
	tmp := make([]string, len(_CassandraOpsRequestTypeNames))
	copy(tmp, _CassandraOpsRequestTypeNames)
	return tmp
}

// CassandraOpsRequestTypeValues returns a list of the values for CassandraOpsRequestType
func CassandraOpsRequestTypeValues() []CassandraOpsRequestType {
	return []CassandraOpsRequestType{
		CassandraOpsRequestTypeUpdateVersion,
		CassandraOpsRequestTypeVerticalScaling,
		CassandraOpsRequestTypeRestart,
	}
}

// String implements the Stringer interface.
func (x CassandraOpsRequestType) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x CassandraOpsRequestType) IsValid() bool {
	_, err := ParseCassandraOpsRequestType(string(x))
	return err == nil
}

var _CassandraOpsRequestTypeValue = map[string]CassandraOpsRequestType{
	"UpdateVersion":   CassandraOpsRequestTypeUpdateVersion,
	"VerticalScaling": CassandraOpsRequestTypeVerticalScaling,
	"Restart":         CassandraOpsRequestTypeRestart,
}

// ParseCassandraOpsRequestType attempts to convert a string to a CassandraOpsRequestType.
func ParseCassandraOpsRequestType(name string) (CassandraOpsRequestType, error) {
	if x, ok := _CassandraOpsRequestTypeValue[name]; ok {
		return x, nil
	}
	return CassandraOpsRequestType(""), fmt.Errorf("%s is %w", name, ErrInvalidCassandraOpsRequestType)
}

// MustParseCassandraOpsRequestType converts a string to a CassandraOpsRequestType, and panics if is not valid.
func MustParseCassandraOpsRequestType(name string) CassandraOpsRequestType {
	val, err := ParseCassandraOpsRequestType(name)
	if err != nil {
		panic(err)
	}
	return val
}
