// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: restaking_protocol/restaking/coordinator/v1/tx.proto

package types

import (
	fmt "fmt"
	_ "github.com/cometbft/cometbft/proto/tendermint/crypto"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	io "io"
	math "math"
	math_bits "math/bits"
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

type MsgRegisterOperator struct {
	ConsumerChainIDs           []string `protobuf:"bytes,1,rep,name=ConsumerChainIDs,proto3" json:"ConsumerChainIDs,omitempty"`
	ConsumerValidatorAddresses []string `protobuf:"bytes,2,rep,name=consumer_validator_addresses,json=consumerValidatorAddresses,proto3" json:"consumer_validator_addresses,omitempty"`
	RestakingDenom             string   `protobuf:"bytes,3,opt,name=restaking_denom,json=restakingDenom,proto3" json:"restaking_denom,omitempty"`
	Sender                     string   `protobuf:"bytes,4,opt,name=sender,proto3" json:"sender,omitempty"`
}

func (m *MsgRegisterOperator) Reset()         { *m = MsgRegisterOperator{} }
func (m *MsgRegisterOperator) String() string { return proto.CompactTextString(m) }
func (*MsgRegisterOperator) ProtoMessage()    {}
func (*MsgRegisterOperator) Descriptor() ([]byte, []int) {
	return fileDescriptor_5f608792a1167783, []int{0}
}
func (m *MsgRegisterOperator) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRegisterOperator) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRegisterOperator.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRegisterOperator) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRegisterOperator.Merge(m, src)
}
func (m *MsgRegisterOperator) XXX_Size() int {
	return m.Size()
}
func (m *MsgRegisterOperator) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRegisterOperator.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRegisterOperator proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgRegisterOperator)(nil), "restaking_protocol.restaking.coordinator.v1.MsgRegisterOperator")
}

func init() {
	proto.RegisterFile("restaking_protocol/restaking/coordinator/v1/tx.proto", fileDescriptor_5f608792a1167783)
}

var fileDescriptor_5f608792a1167783 = []byte{
	// 332 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x32, 0x29, 0x4a, 0x2d, 0x2e,
	0x49, 0xcc, 0xce, 0xcc, 0x4b, 0x8f, 0x2f, 0x28, 0xca, 0x2f, 0xc9, 0x4f, 0xce, 0xcf, 0xd1, 0x87,
	0x0b, 0xe9, 0x27, 0xe7, 0xe7, 0x17, 0xa5, 0x64, 0xe6, 0x25, 0x96, 0xe4, 0x17, 0xe9, 0x97, 0x19,
	0xea, 0x97, 0x54, 0xe8, 0x81, 0x15, 0x09, 0x69, 0x63, 0xea, 0xd2, 0x83, 0x0b, 0xe9, 0x21, 0xe9,
	0xd2, 0x2b, 0x33, 0x94, 0x12, 0x49, 0xcf, 0x4f, 0xcf, 0x07, 0x2b, 0xd3, 0x07, 0xb1, 0x20, 0x46,
	0x48, 0x49, 0x26, 0xe7, 0x17, 0xe7, 0xe6, 0x17, 0x43, 0xf4, 0xeb, 0x43, 0x38, 0x50, 0x29, 0x99,
	0x92, 0xd4, 0xbc, 0x94, 0xd4, 0xa2, 0xdc, 0xcc, 0xbc, 0x12, 0xfd, 0xe4, 0xa2, 0xca, 0x82, 0x92,
	0x7c, 0xfd, 0xec, 0xd4, 0x4a, 0xa8, 0xac, 0xd2, 0x3f, 0x46, 0x2e, 0x61, 0xdf, 0xe2, 0xf4, 0xa0,
	0xd4, 0xf4, 0xcc, 0xe2, 0x92, 0xd4, 0x22, 0xff, 0x82, 0xd4, 0x22, 0x90, 0x4d, 0x42, 0x5a, 0x5c,
	0x02, 0xce, 0xf9, 0x79, 0xc5, 0xa5, 0xb9, 0xa9, 0x45, 0xce, 0x19, 0x89, 0x99, 0x79, 0x9e, 0x2e,
	0xc5, 0x12, 0x8c, 0x0a, 0xcc, 0x1a, 0x9c, 0x41, 0x18, 0xe2, 0x42, 0x51, 0x5c, 0x32, 0xc9, 0x50,
	0xb1, 0xf8, 0xb2, 0xc4, 0x9c, 0xcc, 0x14, 0x90, 0x09, 0xf1, 0x89, 0x29, 0x29, 0x45, 0xa9, 0xc5,
	0xc5, 0xa9, 0xc5, 0x12, 0x4c, 0x20, 0x7d, 0x4e, 0x12, 0x97, 0xb6, 0xe8, 0x8a, 0x40, 0x5d, 0xe6,
	0x08, 0x91, 0x0b, 0x2e, 0x29, 0xca, 0xcc, 0x4b, 0x0f, 0x92, 0x82, 0xe9, 0x0e, 0x83, 0x69, 0x76,
	0x84, 0xe9, 0x15, 0x52, 0xe7, 0xe2, 0x47, 0x84, 0x4e, 0x4a, 0x6a, 0x5e, 0x7e, 0xae, 0x04, 0xb3,
	0x02, 0xa3, 0x06, 0x67, 0x10, 0x1f, 0x5c, 0xd8, 0x05, 0x24, 0x2a, 0x64, 0xc0, 0xc5, 0x56, 0x0c,
	0xf6, 0xa8, 0x04, 0x0b, 0x48, 0x1e, 0x8f, 0x75, 0x50, 0x75, 0x56, 0x2c, 0x1d, 0x0b, 0xe4, 0x19,
	0x9c, 0xbc, 0x4e, 0x3c, 0x92, 0x63, 0xbc, 0xf0, 0x48, 0x8e, 0xf1, 0xc1, 0x23, 0x39, 0xc6, 0x09,
	0x8f, 0xe5, 0x18, 0x2e, 0x3c, 0x96, 0x63, 0xb8, 0xf1, 0x58, 0x8e, 0x21, 0xca, 0x00, 0x4b, 0x64,
	0x56, 0xe0, 0x88, 0xce, 0x92, 0xca, 0x82, 0xd4, 0xe2, 0x24, 0x36, 0xb0, 0x3a, 0x63, 0x40, 0x00,
	0x00, 0x00, 0xff, 0xff, 0xf3, 0x38, 0x63, 0x14, 0x07, 0x02, 0x00, 0x00,
}

func (m *MsgRegisterOperator) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRegisterOperator) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRegisterOperator) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Sender) > 0 {
		i -= len(m.Sender)
		copy(dAtA[i:], m.Sender)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Sender)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.RestakingDenom) > 0 {
		i -= len(m.RestakingDenom)
		copy(dAtA[i:], m.RestakingDenom)
		i = encodeVarintTx(dAtA, i, uint64(len(m.RestakingDenom)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.ConsumerValidatorAddresses) > 0 {
		for iNdEx := len(m.ConsumerValidatorAddresses) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.ConsumerValidatorAddresses[iNdEx])
			copy(dAtA[i:], m.ConsumerValidatorAddresses[iNdEx])
			i = encodeVarintTx(dAtA, i, uint64(len(m.ConsumerValidatorAddresses[iNdEx])))
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.ConsumerChainIDs) > 0 {
		for iNdEx := len(m.ConsumerChainIDs) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.ConsumerChainIDs[iNdEx])
			copy(dAtA[i:], m.ConsumerChainIDs[iNdEx])
			i = encodeVarintTx(dAtA, i, uint64(len(m.ConsumerChainIDs[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgRegisterOperator) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.ConsumerChainIDs) > 0 {
		for _, s := range m.ConsumerChainIDs {
			l = len(s)
			n += 1 + l + sovTx(uint64(l))
		}
	}
	if len(m.ConsumerValidatorAddresses) > 0 {
		for _, s := range m.ConsumerValidatorAddresses {
			l = len(s)
			n += 1 + l + sovTx(uint64(l))
		}
	}
	l = len(m.RestakingDenom)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgRegisterOperator) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgRegisterOperator: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRegisterOperator: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ConsumerChainIDs", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ConsumerChainIDs = append(m.ConsumerChainIDs, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ConsumerValidatorAddresses", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ConsumerValidatorAddresses = append(m.ConsumerValidatorAddresses, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RestakingDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RestakingDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sender", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sender = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
