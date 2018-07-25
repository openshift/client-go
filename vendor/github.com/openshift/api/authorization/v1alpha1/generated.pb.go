// Code generated by protoc-gen-gogo.
// source: github.com/openshift/api/authorization/v1alpha1/generated.proto
// DO NOT EDIT!

/*
	Package v1alpha1 is a generated protocol buffer package.

	It is generated from these files:
		github.com/openshift/api/authorization/v1alpha1/generated.proto

	It has these top-level messages:
		AccessRestriction
		AccessRestrictionList
		AccessRestrictionSpec
		SubjectMatcher
*/
package v1alpha1

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

import github_com_openshift_api_authorization_v1 "github.com/openshift/api/authorization/v1"
import k8s_io_api_rbac_v1 "k8s.io/api/rbac/v1"

import strings "strings"
import reflect "reflect"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

func (m *AccessRestriction) Reset()                    { *m = AccessRestriction{} }
func (*AccessRestriction) ProtoMessage()               {}
func (*AccessRestriction) Descriptor() ([]byte, []int) { return fileDescriptorGenerated, []int{0} }

func (m *AccessRestrictionList) Reset()                    { *m = AccessRestrictionList{} }
func (*AccessRestrictionList) ProtoMessage()               {}
func (*AccessRestrictionList) Descriptor() ([]byte, []int) { return fileDescriptorGenerated, []int{1} }

func (m *AccessRestrictionSpec) Reset()                    { *m = AccessRestrictionSpec{} }
func (*AccessRestrictionSpec) ProtoMessage()               {}
func (*AccessRestrictionSpec) Descriptor() ([]byte, []int) { return fileDescriptorGenerated, []int{2} }

func (m *SubjectMatcher) Reset()                    { *m = SubjectMatcher{} }
func (*SubjectMatcher) ProtoMessage()               {}
func (*SubjectMatcher) Descriptor() ([]byte, []int) { return fileDescriptorGenerated, []int{3} }

func init() {
	proto.RegisterType((*AccessRestriction)(nil), "github.com.openshift.api.authorization.v1alpha1.AccessRestriction")
	proto.RegisterType((*AccessRestrictionList)(nil), "github.com.openshift.api.authorization.v1alpha1.AccessRestrictionList")
	proto.RegisterType((*AccessRestrictionSpec)(nil), "github.com.openshift.api.authorization.v1alpha1.AccessRestrictionSpec")
	proto.RegisterType((*SubjectMatcher)(nil), "github.com.openshift.api.authorization.v1alpha1.SubjectMatcher")
}
func (m *AccessRestriction) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AccessRestriction) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintGenerated(dAtA, i, uint64(m.ObjectMeta.Size()))
	n1, err := m.ObjectMeta.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	dAtA[i] = 0x12
	i++
	i = encodeVarintGenerated(dAtA, i, uint64(m.Spec.Size()))
	n2, err := m.Spec.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n2
	return i, nil
}

func (m *AccessRestrictionList) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AccessRestrictionList) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintGenerated(dAtA, i, uint64(m.ListMeta.Size()))
	n3, err := m.ListMeta.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n3
	if len(m.Items) > 0 {
		for _, msg := range m.Items {
			dAtA[i] = 0x12
			i++
			i = encodeVarintGenerated(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *AccessRestrictionSpec) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AccessRestrictionSpec) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.MatchAttributes) > 0 {
		for _, msg := range m.MatchAttributes {
			dAtA[i] = 0xa
			i++
			i = encodeVarintGenerated(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.AllowedSubjects) > 0 {
		for _, msg := range m.AllowedSubjects {
			dAtA[i] = 0x12
			i++
			i = encodeVarintGenerated(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.DeniedSubjects) > 0 {
		for _, msg := range m.DeniedSubjects {
			dAtA[i] = 0x1a
			i++
			i = encodeVarintGenerated(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *SubjectMatcher) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SubjectMatcher) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.UserRestriction != nil {
		dAtA[i] = 0xa
		i++
		i = encodeVarintGenerated(dAtA, i, uint64(m.UserRestriction.Size()))
		n4, err := m.UserRestriction.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n4
	}
	if m.GroupRestriction != nil {
		dAtA[i] = 0x12
		i++
		i = encodeVarintGenerated(dAtA, i, uint64(m.GroupRestriction.Size()))
		n5, err := m.GroupRestriction.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n5
	}
	return i, nil
}

func encodeFixed64Generated(dAtA []byte, offset int, v uint64) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	dAtA[offset+4] = uint8(v >> 32)
	dAtA[offset+5] = uint8(v >> 40)
	dAtA[offset+6] = uint8(v >> 48)
	dAtA[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Generated(dAtA []byte, offset int, v uint32) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintGenerated(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *AccessRestriction) Size() (n int) {
	var l int
	_ = l
	l = m.ObjectMeta.Size()
	n += 1 + l + sovGenerated(uint64(l))
	l = m.Spec.Size()
	n += 1 + l + sovGenerated(uint64(l))
	return n
}

func (m *AccessRestrictionList) Size() (n int) {
	var l int
	_ = l
	l = m.ListMeta.Size()
	n += 1 + l + sovGenerated(uint64(l))
	if len(m.Items) > 0 {
		for _, e := range m.Items {
			l = e.Size()
			n += 1 + l + sovGenerated(uint64(l))
		}
	}
	return n
}

func (m *AccessRestrictionSpec) Size() (n int) {
	var l int
	_ = l
	if len(m.MatchAttributes) > 0 {
		for _, e := range m.MatchAttributes {
			l = e.Size()
			n += 1 + l + sovGenerated(uint64(l))
		}
	}
	if len(m.AllowedSubjects) > 0 {
		for _, e := range m.AllowedSubjects {
			l = e.Size()
			n += 1 + l + sovGenerated(uint64(l))
		}
	}
	if len(m.DeniedSubjects) > 0 {
		for _, e := range m.DeniedSubjects {
			l = e.Size()
			n += 1 + l + sovGenerated(uint64(l))
		}
	}
	return n
}

func (m *SubjectMatcher) Size() (n int) {
	var l int
	_ = l
	if m.UserRestriction != nil {
		l = m.UserRestriction.Size()
		n += 1 + l + sovGenerated(uint64(l))
	}
	if m.GroupRestriction != nil {
		l = m.GroupRestriction.Size()
		n += 1 + l + sovGenerated(uint64(l))
	}
	return n
}

func sovGenerated(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozGenerated(x uint64) (n int) {
	return sovGenerated(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *AccessRestriction) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&AccessRestriction{`,
		`ObjectMeta:` + strings.Replace(strings.Replace(this.ObjectMeta.String(), "ObjectMeta", "k8s_io_apimachinery_pkg_apis_meta_v1.ObjectMeta", 1), `&`, ``, 1) + `,`,
		`Spec:` + strings.Replace(strings.Replace(this.Spec.String(), "AccessRestrictionSpec", "AccessRestrictionSpec", 1), `&`, ``, 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *AccessRestrictionList) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&AccessRestrictionList{`,
		`ListMeta:` + strings.Replace(strings.Replace(this.ListMeta.String(), "ListMeta", "k8s_io_apimachinery_pkg_apis_meta_v1.ListMeta", 1), `&`, ``, 1) + `,`,
		`Items:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.Items), "AccessRestriction", "AccessRestriction", 1), `&`, ``, 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *AccessRestrictionSpec) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&AccessRestrictionSpec{`,
		`MatchAttributes:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.MatchAttributes), "PolicyRule", "k8s_io_api_rbac_v1.PolicyRule", 1), `&`, ``, 1) + `,`,
		`AllowedSubjects:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.AllowedSubjects), "SubjectMatcher", "SubjectMatcher", 1), `&`, ``, 1) + `,`,
		`DeniedSubjects:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.DeniedSubjects), "SubjectMatcher", "SubjectMatcher", 1), `&`, ``, 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *SubjectMatcher) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&SubjectMatcher{`,
		`UserRestriction:` + strings.Replace(fmt.Sprintf("%v", this.UserRestriction), "UserRestriction", "github_com_openshift_api_authorization_v1.UserRestriction", 1) + `,`,
		`GroupRestriction:` + strings.Replace(fmt.Sprintf("%v", this.GroupRestriction), "GroupRestriction", "github_com_openshift_api_authorization_v1.GroupRestriction", 1) + `,`,
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
func (m *AccessRestriction) Unmarshal(dAtA []byte) error {
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
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: AccessRestriction: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AccessRestriction: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ObjectMeta", wireType)
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
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ObjectMeta.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Spec", wireType)
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
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Spec.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
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
func (m *AccessRestrictionList) Unmarshal(dAtA []byte) error {
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
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: AccessRestrictionList: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AccessRestrictionList: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ListMeta", wireType)
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
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ListMeta.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Items", wireType)
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
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Items = append(m.Items, AccessRestriction{})
			if err := m.Items[len(m.Items)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
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
func (m *AccessRestrictionSpec) Unmarshal(dAtA []byte) error {
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
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: AccessRestrictionSpec: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AccessRestrictionSpec: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MatchAttributes", wireType)
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
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MatchAttributes = append(m.MatchAttributes, k8s_io_api_rbac_v1.PolicyRule{})
			if err := m.MatchAttributes[len(m.MatchAttributes)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AllowedSubjects", wireType)
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
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AllowedSubjects = append(m.AllowedSubjects, SubjectMatcher{})
			if err := m.AllowedSubjects[len(m.AllowedSubjects)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DeniedSubjects", wireType)
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
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DeniedSubjects = append(m.DeniedSubjects, SubjectMatcher{})
			if err := m.DeniedSubjects[len(m.DeniedSubjects)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
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
func (m *SubjectMatcher) Unmarshal(dAtA []byte) error {
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
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SubjectMatcher: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SubjectMatcher: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field UserRestriction", wireType)
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
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.UserRestriction == nil {
				m.UserRestriction = &github_com_openshift_api_authorization_v1.UserRestriction{}
			}
			if err := m.UserRestriction.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field GroupRestriction", wireType)
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
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.GroupRestriction == nil {
				m.GroupRestriction = &github_com_openshift_api_authorization_v1.GroupRestriction{}
			}
			if err := m.GroupRestriction.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
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
			iNdEx += length
			if length < 0 {
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

func init() {
	proto.RegisterFile("github.com/openshift/api/authorization/v1alpha1/generated.proto", fileDescriptorGenerated)
}

var fileDescriptorGenerated = []byte{
	// 588 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x93, 0x4d, 0x6f, 0xd3, 0x30,
	0x1c, 0xc6, 0x9b, 0xbd, 0x48, 0x93, 0x07, 0x6d, 0x09, 0x6f, 0x55, 0x0f, 0xe9, 0xd4, 0xd3, 0x2e,
	0x38, 0x74, 0x20, 0xc4, 0xcb, 0x61, 0x6a, 0x84, 0x40, 0x48, 0x4c, 0xa0, 0x4c, 0x5c, 0x10, 0x07,
	0x5c, 0xd7, 0x4b, 0x4c, 0x93, 0x38, 0xb2, 0x9d, 0xa1, 0x22, 0x21, 0x26, 0x8e, 0x9c, 0xf8, 0x58,
	0x3d, 0xee, 0xb8, 0x53, 0xa1, 0xe1, 0x8b, 0x20, 0xbb, 0xe9, 0x9a, 0xa4, 0x9d, 0xd4, 0x0a, 0x6e,
	0xf5, 0x5f, 0xff, 0xe7, 0xf7, 0x3c, 0x8f, 0x9d, 0x82, 0x43, 0x8f, 0x4a, 0x3f, 0xe9, 0x41, 0xcc,
	0x42, 0x9b, 0xc5, 0x24, 0x12, 0x3e, 0x3d, 0x91, 0x36, 0x8a, 0xa9, 0x8d, 0x12, 0xe9, 0x33, 0x4e,
	0xbf, 0x20, 0x49, 0x59, 0x64, 0x9f, 0x76, 0x50, 0x10, 0xfb, 0xa8, 0x63, 0x7b, 0x24, 0x22, 0x1c,
	0x49, 0xd2, 0x87, 0x31, 0x67, 0x92, 0x99, 0xf6, 0x1c, 0x00, 0x2f, 0x01, 0x10, 0xc5, 0x14, 0x16,
	0x00, 0x70, 0x06, 0x68, 0xde, 0xcb, 0x39, 0x7a, 0xcc, 0x63, 0xb6, 0xe6, 0xf4, 0x92, 0x13, 0x7d,
	0xd2, 0x07, 0xfd, 0x6b, 0xca, 0x6f, 0x3e, 0x59, 0x39, 0x60, 0x39, 0x5a, 0xb3, 0x3d, 0x78, 0x2c,
	0x20, 0x65, 0x7a, 0x99, 0xf7, 0x10, 0x5e, 0xb6, 0xf3, 0x70, 0xbe, 0x13, 0x22, 0xec, 0xd3, 0x88,
	0xf0, 0xa1, 0x1d, 0x0f, 0x3c, 0x35, 0x10, 0x76, 0x48, 0x24, 0x5a, 0xa6, 0x7a, 0x74, 0x95, 0x8a,
	0x27, 0x91, 0xa4, 0x21, 0xb1, 0x05, 0xf6, 0x49, 0x88, 0x16, 0x74, 0x0f, 0xae, 0xd2, 0x25, 0x92,
	0x06, 0x36, 0x8d, 0xa4, 0x90, 0xbc, 0x2c, 0x6a, 0xff, 0x32, 0xc0, 0x8d, 0x2e, 0xc6, 0x44, 0x08,
	0x97, 0x08, 0xc9, 0x29, 0x56, 0x7d, 0xcd, 0x8f, 0x60, 0x47, 0xa5, 0xeb, 0x23, 0x89, 0x1a, 0xc6,
	0x9e, 0xb1, 0xbf, 0x7b, 0x70, 0x1f, 0x4e, 0xe9, 0x30, 0x4f, 0x87, 0xf1, 0xc0, 0x53, 0x03, 0x01,
	0xd5, 0x36, 0x3c, 0xed, 0xc0, 0x37, 0xbd, 0x4f, 0x04, 0xcb, 0x23, 0x22, 0x91, 0x63, 0x8e, 0xc6,
	0xad, 0x4a, 0x3a, 0x6e, 0x81, 0xf9, 0xcc, 0xbd, 0xa4, 0x9a, 0x3e, 0xd8, 0x12, 0x31, 0xc1, 0x8d,
	0x0d, 0x4d, 0x7f, 0x01, 0xd7, 0x7c, 0x68, 0xb8, 0x90, 0xf9, 0x38, 0x26, 0xd8, 0xb9, 0x96, 0x79,
	0x6e, 0xa9, 0x93, 0xab, 0x1d, 0xda, 0x63, 0x03, 0xdc, 0x5e, 0xd8, 0x7e, 0x4d, 0x85, 0x34, 0x3f,
	0x2c, 0xb4, 0x84, 0xab, 0xb5, 0x54, 0x6a, 0xdd, 0xb1, 0x9e, 0xf9, 0xed, 0xcc, 0x26, 0xb9, 0x86,
	0x1e, 0xd8, 0xa6, 0x92, 0x84, 0xa2, 0xb1, 0xb1, 0xb7, 0xb9, 0xbf, 0x7b, 0xe0, 0xfc, 0x7b, 0x45,
	0xe7, 0x7a, 0x66, 0xb7, 0xfd, 0x4a, 0x81, 0xdd, 0x29, 0xbf, 0x7d, 0xb6, 0xb9, 0xa4, 0xa0, 0xba,
	0x00, 0x13, 0x81, 0x5a, 0x88, 0x24, 0xf6, 0xbb, 0x52, 0x72, 0xda, 0x4b, 0x24, 0x11, 0x0d, 0x43,
	0x87, 0xb1, 0x72, 0x3d, 0xa1, 0xfa, 0x7a, 0x55, 0xab, 0xb7, 0x2c, 0xa0, 0x78, 0xe8, 0x26, 0x01,
	0x71, 0xee, 0x66, 0x46, 0xb5, 0xa3, 0xa2, 0xdc, 0x2d, 0xf3, 0xcc, 0xef, 0x06, 0xa8, 0xa1, 0x20,
	0x60, 0x9f, 0x49, 0xff, 0x38, 0xd1, 0x0f, 0x3d, 0x2b, 0x7c, 0xb8, 0x76, 0xe1, 0x0c, 0xa0, 0x3d,
	0x09, 0x9f, 0x87, 0xe8, 0x16, 0xf9, 0x6e, 0xd9, 0xd0, 0xfc, 0x06, 0xaa, 0x7d, 0x12, 0xd1, 0x5c,
	0x84, 0xcd, 0xff, 0x13, 0xe1, 0x4e, 0x16, 0xa1, 0xfa, 0xbc, 0x80, 0x77, 0x4b, 0x76, 0xed, 0x1f,
	0x1b, 0xa0, 0x5a, 0x94, 0x9a, 0x43, 0x50, 0x4b, 0x04, 0xe1, 0xb9, 0x27, 0xc9, 0xbe, 0xb1, 0xa7,
	0xab, 0x87, 0x82, 0xef, 0x8a, 0x04, 0xe7, 0xa6, 0xba, 0x8e, 0xd2, 0xd0, 0x2d, 0xfb, 0x98, 0x5f,
	0x41, 0xdd, 0xe3, 0x2c, 0x89, 0xf3, 0xde, 0xd3, 0xff, 0xd9, 0xb3, 0x35, 0xbc, 0x5f, 0x96, 0x10,
	0xce, 0xad, 0x74, 0xdc, 0xaa, 0x97, 0xa7, 0xee, 0x82, 0x95, 0x03, 0x47, 0x13, 0xab, 0x72, 0x3e,
	0xb1, 0x2a, 0x17, 0x13, 0xab, 0x72, 0x96, 0x5a, 0xc6, 0x28, 0xb5, 0x8c, 0xf3, 0xd4, 0x32, 0x2e,
	0x52, 0xcb, 0xf8, 0x9d, 0x5a, 0xc6, 0xcf, 0x3f, 0x56, 0xe5, 0xfd, 0xce, 0xec, 0xce, 0xff, 0x06,
	0x00, 0x00, 0xff, 0xff, 0x51, 0x2a, 0x0e, 0xdf, 0x26, 0x06, 0x00, 0x00,
}