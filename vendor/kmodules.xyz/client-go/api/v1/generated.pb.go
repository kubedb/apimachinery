/*
Copyright The Kmodules Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: kmodules.xyz/client-go/api/v1/generated.proto

package v1

import (
	fmt "fmt"
	io "io"
	math "math"
	math_bits "math/bits"
	reflect "reflect"
	strings "strings"

	proto "github.com/gogo/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

func (m *Condition) Reset()      { *m = Condition{} }
func (*Condition) ProtoMessage() {}
func (*Condition) Descriptor() ([]byte, []int) {
	return fileDescriptor_af8e7a11c7a1ccd9, []int{0}
}
func (m *Condition) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Condition) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (m *Condition) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Condition.Merge(m, src)
}
func (m *Condition) XXX_Size() int {
	return m.Size()
}
func (m *Condition) XXX_DiscardUnknown() {
	xxx_messageInfo_Condition.DiscardUnknown(m)
}

var xxx_messageInfo_Condition proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Condition)(nil), "kmodules.xyz.client_go.api.v1.Condition")
}

func init() {
	proto.RegisterFile("kmodules.xyz/client-go/api/v1/generated.proto", fileDescriptor_af8e7a11c7a1ccd9)
}

var fileDescriptor_af8e7a11c7a1ccd9 = []byte{
	// 395 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x90, 0x31, 0x8f, 0xd3, 0x30,
	0x1c, 0xc5, 0xe3, 0x5e, 0xc9, 0xe9, 0x0c, 0xe2, 0x24, 0x4f, 0x51, 0x25, 0xdc, 0xc0, 0x80, 0x02,
	0x52, 0x6d, 0x05, 0x31, 0xc0, 0x5a, 0x06, 0x24, 0x04, 0x42, 0x0a, 0x9d, 0x58, 0x90, 0xdb, 0x18,
	0x9f, 0x95, 0xc6, 0x8e, 0x62, 0x27, 0x22, 0x4c, 0x7c, 0x04, 0x3e, 0x56, 0xc7, 0x1b, 0x3b, 0x55,
	0x34, 0x7c, 0x0b, 0x16, 0x50, 0x9c, 0x94, 0xab, 0xd4, 0x6e, 0xf6, 0xfb, 0xff, 0xde, 0x7b, 0xf6,
	0x1f, 0xce, 0xb2, 0x5c, 0xa7, 0xd5, 0x9a, 0x1b, 0xf2, 0xad, 0xf9, 0x4e, 0x57, 0x6b, 0xc9, 0x95,
	0x9d, 0x09, 0x4d, 0x59, 0x21, 0x69, 0x1d, 0x53, 0xc1, 0x15, 0x2f, 0x99, 0xe5, 0x29, 0x29, 0x4a,
	0x6d, 0x35, 0x7a, 0x74, 0x8c, 0x93, 0x1e, 0xff, 0x22, 0x34, 0x61, 0x85, 0x24, 0x75, 0x3c, 0x99,
	0x09, 0x69, 0x6f, 0xaa, 0x25, 0x59, 0xe9, 0x9c, 0x0a, 0x2d, 0x34, 0x75, 0xae, 0x65, 0xf5, 0xd5,
	0xdd, 0xdc, 0xc5, 0x9d, 0xfa, 0xb4, 0xc9, 0xcb, 0xec, 0x95, 0x21, 0xd2, 0x95, 0xe5, 0x6c, 0x75,
	0x23, 0x15, 0x2f, 0x1b, 0x5a, 0x64, 0xa2, 0x13, 0x0c, 0xcd, 0xb9, 0x65, 0x67, 0xde, 0xf0, 0xe4,
	0xef, 0x08, 0x5e, 0xbd, 0xd1, 0x2a, 0x95, 0x56, 0x6a, 0x85, 0x42, 0x38, 0xb6, 0x4d, 0xc1, 0x03,
	0x10, 0x82, 0xe8, 0x6a, 0xfe, 0x60, 0xb3, 0x9b, 0x7a, 0xed, 0x6e, 0x3a, 0x5e, 0x34, 0x05, 0x4f,
	0xdc, 0x04, 0xbd, 0x86, 0xbe, 0xb1, 0xcc, 0x56, 0x26, 0x18, 0x39, 0xe6, 0xf1, 0xc0, 0xf8, 0x9f,
	0x9c, 0xfa, 0x67, 0x37, 0xbd, 0xfe, 0x1f, 0xd7, 0x4b, 0xc9, 0x60, 0x40, 0xef, 0x20, 0xd2, 0x4b,
	0xc3, 0xcb, 0x9a, 0xa7, 0x6f, 0xfb, 0x57, 0x48, 0xad, 0x82, 0x8b, 0x10, 0x44, 0x17, 0xf3, 0xc9,
	0x10, 0x83, 0x3e, 0x9e, 0x10, 0xc9, 0x19, 0x17, 0xaa, 0x21, 0x5a, 0x33, 0x63, 0x17, 0x25, 0x53,
	0xc6, 0x75, 0x2d, 0x64, 0xce, 0x83, 0x71, 0x08, 0xa2, 0xfb, 0x2f, 0x9e, 0x93, 0x7e, 0x13, 0xe4,
	0x78, 0x13, 0xa4, 0xc8, 0x44, 0x27, 0x18, 0xd2, 0x6d, 0x82, 0xd4, 0x31, 0xe9, 0x1c, 0x77, 0xbd,
	0xef, 0x4f, 0xd2, 0x92, 0x33, 0x0d, 0xe8, 0x29, 0xf4, 0x4b, 0xce, 0x8c, 0x56, 0xc1, 0x3d, 0xf7,
	0xfd, 0x87, 0x87, 0xef, 0x27, 0x4e, 0x4d, 0x86, 0x29, 0x7a, 0x06, 0x2f, 0x73, 0x6e, 0x0c, 0x13,
	0x3c, 0xf0, 0x1d, 0x78, 0x3d, 0x80, 0x97, 0x1f, 0x7a, 0x39, 0x39, 0xcc, 0xe7, 0xd1, 0x66, 0x8f,
	0xbd, 0xdb, 0x3d, 0xf6, 0xb6, 0x7b, 0xec, 0xfd, 0x68, 0x31, 0xd8, 0xb4, 0x18, 0xdc, 0xb6, 0x18,
	0x6c, 0x5b, 0x0c, 0x7e, 0xb5, 0x18, 0xfc, 0xfc, 0x8d, 0xbd, 0xcf, 0xa3, 0x3a, 0xfe, 0x17, 0x00,
	0x00, 0xff, 0xff, 0xa4, 0x3d, 0x14, 0x21, 0x5f, 0x02, 0x00, 0x00,
}

func (m *Condition) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Condition) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Condition) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	i -= len(m.Message)
	copy(dAtA[i:], m.Message)
	i = encodeVarintGenerated(dAtA, i, uint64(len(m.Message)))
	i--
	dAtA[i] = 0x32
	i -= len(m.Reason)
	copy(dAtA[i:], m.Reason)
	i = encodeVarintGenerated(dAtA, i, uint64(len(m.Reason)))
	i--
	dAtA[i] = 0x2a
	{
		size, err := m.LastTransitionTime.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenerated(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
	i = encodeVarintGenerated(dAtA, i, uint64(m.ObservedGeneration))
	i--
	dAtA[i] = 0x18
	i -= len(m.Status)
	copy(dAtA[i:], m.Status)
	i = encodeVarintGenerated(dAtA, i, uint64(len(m.Status)))
	i--
	dAtA[i] = 0x12
	i -= len(m.Type)
	copy(dAtA[i:], m.Type)
	i = encodeVarintGenerated(dAtA, i, uint64(len(m.Type)))
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintGenerated(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenerated(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Condition) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Type)
	n += 1 + l + sovGenerated(uint64(l))
	l = len(m.Status)
	n += 1 + l + sovGenerated(uint64(l))
	n += 1 + sovGenerated(uint64(m.ObservedGeneration))
	l = m.LastTransitionTime.Size()
	n += 1 + l + sovGenerated(uint64(l))
	l = len(m.Reason)
	n += 1 + l + sovGenerated(uint64(l))
	l = len(m.Message)
	n += 1 + l + sovGenerated(uint64(l))
	return n
}

func sovGenerated(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenerated(x uint64) (n int) {
	return sovGenerated(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *Condition) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&Condition{`,
		`Type:` + fmt.Sprintf("%v", this.Type) + `,`,
		`Status:` + fmt.Sprintf("%v", this.Status) + `,`,
		`ObservedGeneration:` + fmt.Sprintf("%v", this.ObservedGeneration) + `,`,
		`LastTransitionTime:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.LastTransitionTime), "Time", "v1.Time", 1), `&`, ``, 1) + `,`,
		`Reason:` + fmt.Sprintf("%v", this.Reason) + `,`,
		`Message:` + fmt.Sprintf("%v", this.Message) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringGenerated(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *Condition) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenerated
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Condition: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Condition: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Type = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Status", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Status = ConditionStatus(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ObservedGeneration", wireType)
			}
			m.ObservedGeneration = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ObservedGeneration |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastTransitionTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.LastTransitionTime.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Reason", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Reason = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Message", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Message = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenerated(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGenerated
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthGenerated
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipGenerated(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenerated
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthGenerated
			}
			iNdEx += length
			if iNdEx < 0 {
				return 0, ErrInvalidLengthGenerated
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowGenerated
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipGenerated(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
				if iNdEx < 0 {
					return 0, ErrInvalidLengthGenerated
				}
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthGenerated = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenerated   = fmt.Errorf("proto: integer overflow")
)
