// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.5
// source: pkg/tbtc/gen/pb/wallet.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Wallet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PublicKey             []byte   `protobuf:"bytes,1,opt,name=publicKey,proto3" json:"publicKey,omitempty"`
	SigningGroupOperators []string `protobuf:"bytes,2,rep,name=signingGroupOperators,proto3" json:"signingGroupOperators,omitempty"`
}

func (x *Wallet) Reset() {
	*x = Wallet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_tbtc_gen_pb_wallet_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Wallet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Wallet) ProtoMessage() {}

func (x *Wallet) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_tbtc_gen_pb_wallet_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Wallet.ProtoReflect.Descriptor instead.
func (*Wallet) Descriptor() ([]byte, []int) {
	return file_pkg_tbtc_gen_pb_wallet_proto_rawDescGZIP(), []int{0}
}

func (x *Wallet) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

func (x *Wallet) GetSigningGroupOperators() []string {
	if x != nil {
		return x.SigningGroupOperators
	}
	return nil
}

type Signer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Wallet                  *Wallet `protobuf:"bytes,1,opt,name=wallet,proto3" json:"wallet,omitempty"`
	SigningGroupMemberIndex uint32  `protobuf:"varint,2,opt,name=signingGroupMemberIndex,proto3" json:"signingGroupMemberIndex,omitempty"`
	PrivateKeyShare         []byte  `protobuf:"bytes,3,opt,name=privateKeyShare,proto3" json:"privateKeyShare,omitempty"`
}

func (x *Signer) Reset() {
	*x = Signer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_tbtc_gen_pb_wallet_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Signer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Signer) ProtoMessage() {}

func (x *Signer) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_tbtc_gen_pb_wallet_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Signer.ProtoReflect.Descriptor instead.
func (*Signer) Descriptor() ([]byte, []int) {
	return file_pkg_tbtc_gen_pb_wallet_proto_rawDescGZIP(), []int{1}
}

func (x *Signer) GetWallet() *Wallet {
	if x != nil {
		return x.Wallet
	}
	return nil
}

func (x *Signer) GetSigningGroupMemberIndex() uint32 {
	if x != nil {
		return x.SigningGroupMemberIndex
	}
	return 0
}

func (x *Signer) GetPrivateKeyShare() []byte {
	if x != nil {
		return x.PrivateKeyShare
	}
	return nil
}

var File_pkg_tbtc_gen_pb_wallet_proto protoreflect.FileDescriptor

var file_pkg_tbtc_gen_pb_wallet_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x62, 0x74, 0x63, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70,
	0x62, 0x2f, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04,
	0x74, 0x62, 0x74, 0x63, 0x22, 0x5c, 0x0a, 0x06, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x12, 0x1c,
	0x0a, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x12, 0x34, 0x0a, 0x15,
	0x73, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4f, 0x70, 0x65, 0x72,
	0x61, 0x74, 0x6f, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x15, 0x73, 0x69, 0x67,
	0x6e, 0x69, 0x6e, 0x67, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f,
	0x72, 0x73, 0x22, 0x92, 0x01, 0x0a, 0x06, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x72, 0x12, 0x24, 0x0a,
	0x06, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e,
	0x74, 0x62, 0x74, 0x63, 0x2e, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x52, 0x06, 0x77, 0x61, 0x6c,
	0x6c, 0x65, 0x74, 0x12, 0x38, 0x0a, 0x17, 0x73, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x47, 0x72,
	0x6f, 0x75, 0x70, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x17, 0x73, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x47, 0x72, 0x6f,
	0x75, 0x70, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x28, 0x0a,
	0x0f, 0x70, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x53, 0x68, 0x61, 0x72, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0f, 0x70, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x4b,
	0x65, 0x79, 0x53, 0x68, 0x61, 0x72, 0x65, 0x42, 0x06, 0x5a, 0x04, 0x2e, 0x2f, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_tbtc_gen_pb_wallet_proto_rawDescOnce sync.Once
	file_pkg_tbtc_gen_pb_wallet_proto_rawDescData = file_pkg_tbtc_gen_pb_wallet_proto_rawDesc
)

func file_pkg_tbtc_gen_pb_wallet_proto_rawDescGZIP() []byte {
	file_pkg_tbtc_gen_pb_wallet_proto_rawDescOnce.Do(func() {
		file_pkg_tbtc_gen_pb_wallet_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_tbtc_gen_pb_wallet_proto_rawDescData)
	})
	return file_pkg_tbtc_gen_pb_wallet_proto_rawDescData
}

var file_pkg_tbtc_gen_pb_wallet_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_pkg_tbtc_gen_pb_wallet_proto_goTypes = []interface{}{
	(*Wallet)(nil), // 0: tbtc.Wallet
	(*Signer)(nil), // 1: tbtc.Signer
}
var file_pkg_tbtc_gen_pb_wallet_proto_depIdxs = []int32{
	0, // 0: tbtc.Signer.wallet:type_name -> tbtc.Wallet
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_pkg_tbtc_gen_pb_wallet_proto_init() }
func file_pkg_tbtc_gen_pb_wallet_proto_init() {
	if File_pkg_tbtc_gen_pb_wallet_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_tbtc_gen_pb_wallet_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Wallet); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_tbtc_gen_pb_wallet_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Signer); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_tbtc_gen_pb_wallet_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_tbtc_gen_pb_wallet_proto_goTypes,
		DependencyIndexes: file_pkg_tbtc_gen_pb_wallet_proto_depIdxs,
		MessageInfos:      file_pkg_tbtc_gen_pb_wallet_proto_msgTypes,
	}.Build()
	File_pkg_tbtc_gen_pb_wallet_proto = out.File
	file_pkg_tbtc_gen_pb_wallet_proto_rawDesc = nil
	file_pkg_tbtc_gen_pb_wallet_proto_goTypes = nil
	file_pkg_tbtc_gen_pb_wallet_proto_depIdxs = nil
}
