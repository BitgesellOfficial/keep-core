// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: pb/wallet.proto

package pb

import (
	bytes "bytes"
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
	reflect "reflect"
	strings "strings"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type Wallet struct {
	PublicKey             []byte   `protobuf:"bytes,1,opt,name=publicKey,proto3" json:"publicKey,omitempty"`
	SigningGroupOperators []string `protobuf:"bytes,2,rep,name=signingGroupOperators,proto3" json:"signingGroupOperators,omitempty"`
}

func (m *Wallet) Reset()      { *m = Wallet{} }
func (*Wallet) ProtoMessage() {}
func (*Wallet) Descriptor() ([]byte, []int) {
	return fileDescriptor_656331f01f421569, []int{0}
}
func (m *Wallet) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Wallet) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Wallet.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Wallet) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Wallet.Merge(m, src)
}
func (m *Wallet) XXX_Size() int {
	return m.Size()
}
func (m *Wallet) XXX_DiscardUnknown() {
	xxx_messageInfo_Wallet.DiscardUnknown(m)
}

var xxx_messageInfo_Wallet proto.InternalMessageInfo

func (m *Wallet) GetPublicKey() []byte {
	if m != nil {
		return m.PublicKey
	}
	return nil
}

func (m *Wallet) GetSigningGroupOperators() []string {
	if m != nil {
		return m.SigningGroupOperators
	}
	return nil
}

type Signer struct {
	Wallet                  []byte `protobuf:"bytes,1,opt,name=wallet,proto3" json:"wallet,omitempty"`
	SigningGroupMemberIndex uint32 `protobuf:"varint,2,opt,name=signingGroupMemberIndex,proto3" json:"signingGroupMemberIndex,omitempty"`
	PrivateKeyShare         []byte `protobuf:"bytes,3,opt,name=privateKeyShare,proto3" json:"privateKeyShare,omitempty"`
}

func (m *Signer) Reset()      { *m = Signer{} }
func (*Signer) ProtoMessage() {}
func (*Signer) Descriptor() ([]byte, []int) {
	return fileDescriptor_656331f01f421569, []int{1}
}
func (m *Signer) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Signer) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Signer.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Signer) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Signer.Merge(m, src)
}
func (m *Signer) XXX_Size() int {
	return m.Size()
}
func (m *Signer) XXX_DiscardUnknown() {
	xxx_messageInfo_Signer.DiscardUnknown(m)
}

var xxx_messageInfo_Signer proto.InternalMessageInfo

func (m *Signer) GetWallet() []byte {
	if m != nil {
		return m.Wallet
	}
	return nil
}

func (m *Signer) GetSigningGroupMemberIndex() uint32 {
	if m != nil {
		return m.SigningGroupMemberIndex
	}
	return 0
}

func (m *Signer) GetPrivateKeyShare() []byte {
	if m != nil {
		return m.PrivateKeyShare
	}
	return nil
}

func init() {
	proto.RegisterType((*Wallet)(nil), "tbtc.Wallet")
	proto.RegisterType((*Signer)(nil), "tbtc.Signer")
}

func init() { proto.RegisterFile("pb/wallet.proto", fileDescriptor_656331f01f421569) }

var fileDescriptor_656331f01f421569 = []byte{
	// 249 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2f, 0x48, 0xd2, 0x2f,
	0x4f, 0xcc, 0xc9, 0x49, 0x2d, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x29, 0x49, 0x2a,
	0x49, 0x56, 0x8a, 0xe1, 0x62, 0x0b, 0x07, 0x8b, 0x0a, 0xc9, 0x70, 0x71, 0x16, 0x94, 0x26, 0xe5,
	0x64, 0x26, 0x7b, 0xa7, 0x56, 0x4a, 0x30, 0x2a, 0x30, 0x6a, 0xf0, 0x04, 0x21, 0x04, 0x84, 0x4c,
	0xb8, 0x44, 0x8b, 0x33, 0xd3, 0xf3, 0x32, 0xf3, 0xd2, 0xdd, 0x8b, 0xf2, 0x4b, 0x0b, 0xfc, 0x0b,
	0x52, 0x8b, 0x12, 0x4b, 0xf2, 0x8b, 0x8a, 0x25, 0x98, 0x14, 0x98, 0x35, 0x38, 0x83, 0xb0, 0x4b,
	0x2a, 0xb5, 0x30, 0x72, 0xb1, 0x05, 0x67, 0xa6, 0xe7, 0xa5, 0x16, 0x09, 0x89, 0x71, 0xb1, 0x41,
	0xac, 0x87, 0x9a, 0x0d, 0xe5, 0x09, 0x59, 0x70, 0x89, 0x23, 0xeb, 0xf5, 0x4d, 0xcd, 0x4d, 0x4a,
	0x2d, 0xf2, 0xcc, 0x4b, 0x49, 0xad, 0x90, 0x60, 0x52, 0x60, 0xd4, 0xe0, 0x0d, 0xc2, 0x25, 0x2d,
	0xa4, 0xc1, 0xc5, 0x5f, 0x50, 0x94, 0x59, 0x96, 0x58, 0x92, 0xea, 0x9d, 0x5a, 0x19, 0x9c, 0x91,
	0x58, 0x94, 0x2a, 0xc1, 0x0c, 0x36, 0x1a, 0x5d, 0xd8, 0xc9, 0xe2, 0xc2, 0x43, 0x39, 0x86, 0x1b,
	0x0f, 0xe5, 0x18, 0x3e, 0x3c, 0x94, 0x63, 0x6c, 0x78, 0x24, 0xc7, 0xb8, 0xe2, 0x91, 0x1c, 0xe3,
	0x89, 0x47, 0x72, 0x8c, 0x17, 0x1e, 0xc9, 0x31, 0x3e, 0x78, 0x24, 0xc7, 0xf8, 0xe2, 0x91, 0x1c,
	0xc3, 0x87, 0x47, 0x72, 0x8c, 0x13, 0x1e, 0xcb, 0x31, 0x5c, 0x78, 0x2c, 0xc7, 0x70, 0xe3, 0xb1,
	0x1c, 0x43, 0x14, 0x53, 0x41, 0x52, 0x12, 0x1b, 0x38, 0xac, 0x8c, 0x01, 0x01, 0x00, 0x00, 0xff,
	0xff, 0x23, 0xd3, 0xa0, 0xbd, 0x3e, 0x01, 0x00, 0x00,
}

func (this *Wallet) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Wallet)
	if !ok {
		that2, ok := that.(Wallet)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !bytes.Equal(this.PublicKey, that1.PublicKey) {
		return false
	}
	if len(this.SigningGroupOperators) != len(that1.SigningGroupOperators) {
		return false
	}
	for i := range this.SigningGroupOperators {
		if this.SigningGroupOperators[i] != that1.SigningGroupOperators[i] {
			return false
		}
	}
	return true
}
func (this *Signer) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Signer)
	if !ok {
		that2, ok := that.(Signer)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !bytes.Equal(this.Wallet, that1.Wallet) {
		return false
	}
	if this.SigningGroupMemberIndex != that1.SigningGroupMemberIndex {
		return false
	}
	if !bytes.Equal(this.PrivateKeyShare, that1.PrivateKeyShare) {
		return false
	}
	return true
}
func (this *Wallet) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 6)
	s = append(s, "&pb.Wallet{")
	s = append(s, "PublicKey: "+fmt.Sprintf("%#v", this.PublicKey)+",\n")
	s = append(s, "SigningGroupOperators: "+fmt.Sprintf("%#v", this.SigningGroupOperators)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *Signer) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 7)
	s = append(s, "&pb.Signer{")
	s = append(s, "Wallet: "+fmt.Sprintf("%#v", this.Wallet)+",\n")
	s = append(s, "SigningGroupMemberIndex: "+fmt.Sprintf("%#v", this.SigningGroupMemberIndex)+",\n")
	s = append(s, "PrivateKeyShare: "+fmt.Sprintf("%#v", this.PrivateKeyShare)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringWallet(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func (m *Wallet) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Wallet) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Wallet) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.SigningGroupOperators) > 0 {
		for iNdEx := len(m.SigningGroupOperators) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.SigningGroupOperators[iNdEx])
			copy(dAtA[i:], m.SigningGroupOperators[iNdEx])
			i = encodeVarintWallet(dAtA, i, uint64(len(m.SigningGroupOperators[iNdEx])))
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.PublicKey) > 0 {
		i -= len(m.PublicKey)
		copy(dAtA[i:], m.PublicKey)
		i = encodeVarintWallet(dAtA, i, uint64(len(m.PublicKey)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Signer) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Signer) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Signer) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.PrivateKeyShare) > 0 {
		i -= len(m.PrivateKeyShare)
		copy(dAtA[i:], m.PrivateKeyShare)
		i = encodeVarintWallet(dAtA, i, uint64(len(m.PrivateKeyShare)))
		i--
		dAtA[i] = 0x1a
	}
	if m.SigningGroupMemberIndex != 0 {
		i = encodeVarintWallet(dAtA, i, uint64(m.SigningGroupMemberIndex))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Wallet) > 0 {
		i -= len(m.Wallet)
		copy(dAtA[i:], m.Wallet)
		i = encodeVarintWallet(dAtA, i, uint64(len(m.Wallet)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintWallet(dAtA []byte, offset int, v uint64) int {
	offset -= sovWallet(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Wallet) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.PublicKey)
	if l > 0 {
		n += 1 + l + sovWallet(uint64(l))
	}
	if len(m.SigningGroupOperators) > 0 {
		for _, s := range m.SigningGroupOperators {
			l = len(s)
			n += 1 + l + sovWallet(uint64(l))
		}
	}
	return n
}

func (m *Signer) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Wallet)
	if l > 0 {
		n += 1 + l + sovWallet(uint64(l))
	}
	if m.SigningGroupMemberIndex != 0 {
		n += 1 + sovWallet(uint64(m.SigningGroupMemberIndex))
	}
	l = len(m.PrivateKeyShare)
	if l > 0 {
		n += 1 + l + sovWallet(uint64(l))
	}
	return n
}

func sovWallet(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozWallet(x uint64) (n int) {
	return sovWallet(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *Wallet) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&Wallet{`,
		`PublicKey:` + fmt.Sprintf("%v", this.PublicKey) + `,`,
		`SigningGroupOperators:` + fmt.Sprintf("%v", this.SigningGroupOperators) + `,`,
		`}`,
	}, "")
	return s
}
func (this *Signer) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&Signer{`,
		`Wallet:` + fmt.Sprintf("%v", this.Wallet) + `,`,
		`SigningGroupMemberIndex:` + fmt.Sprintf("%v", this.SigningGroupMemberIndex) + `,`,
		`PrivateKeyShare:` + fmt.Sprintf("%v", this.PrivateKeyShare) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringWallet(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *Wallet) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowWallet
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
			return fmt.Errorf("proto: Wallet: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Wallet: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PublicKey", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowWallet
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthWallet
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthWallet
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PublicKey = append(m.PublicKey[:0], dAtA[iNdEx:postIndex]...)
			if m.PublicKey == nil {
				m.PublicKey = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SigningGroupOperators", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowWallet
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
				return ErrInvalidLengthWallet
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthWallet
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SigningGroupOperators = append(m.SigningGroupOperators, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipWallet(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthWallet
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
func (m *Signer) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowWallet
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
			return fmt.Errorf("proto: Signer: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Signer: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Wallet", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowWallet
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthWallet
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthWallet
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Wallet = append(m.Wallet[:0], dAtA[iNdEx:postIndex]...)
			if m.Wallet == nil {
				m.Wallet = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SigningGroupMemberIndex", wireType)
			}
			m.SigningGroupMemberIndex = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowWallet
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SigningGroupMemberIndex |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PrivateKeyShare", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowWallet
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthWallet
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthWallet
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PrivateKeyShare = append(m.PrivateKeyShare[:0], dAtA[iNdEx:postIndex]...)
			if m.PrivateKeyShare == nil {
				m.PrivateKeyShare = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipWallet(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthWallet
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
func skipWallet(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowWallet
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
					return 0, ErrIntOverflowWallet
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowWallet
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
				return 0, ErrInvalidLengthWallet
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupWallet
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthWallet
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthWallet        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowWallet          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupWallet = fmt.Errorf("proto: unexpected end of group")
)
