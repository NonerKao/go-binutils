package rvgc

import (
	"debug/elf"
	"encoding/binary"
	"fmt"
	"strconv"
)

type RV_INST_TYPE uint32

const (
	RV_INST_NONE   RV_INST_TYPE = 0
	RV_INST_R_TYPE RV_INST_TYPE = 1
	RV_INST_I_TYPE RV_INST_TYPE = 2
	RV_INST_S_TYPE RV_INST_TYPE = 3
	RV_INST_B_TYPE RV_INST_TYPE = 4
	RV_INST_U_TYPE RV_INST_TYPE = 5
	RV_INST_J_TYPE RV_INST_TYPE = 6
	RV_INST_PSEUDO RV_INST_TYPE = 7
)

var mnem2type = map[string]RV_INST_TYPE{
	"add":  RV_INST_R_TYPE,
	"sub":  RV_INST_R_TYPE,
	"sll":  RV_INST_R_TYPE,
	"slt":  RV_INST_R_TYPE,
	"sltu": RV_INST_R_TYPE,
	"xor":  RV_INST_R_TYPE,
	"srl":  RV_INST_R_TYPE,
	"sra":  RV_INST_R_TYPE,
	"or":   RV_INST_R_TYPE,
	"and":  RV_INST_R_TYPE,
	"addw": RV_INST_R_TYPE,
	"subw": RV_INST_R_TYPE,
	"sllw": RV_INST_R_TYPE,
	"srlw": RV_INST_R_TYPE,
	"sraw": RV_INST_R_TYPE,

	"lb":  RV_INST_I_TYPE,
	"lbu": RV_INST_I_TYPE,
	"lh":  RV_INST_I_TYPE,
	"lhu": RV_INST_I_TYPE,
	"lw":  RV_INST_I_TYPE,
	"lwu": RV_INST_I_TYPE,
	"ld":  RV_INST_I_TYPE,

	"sb": RV_INST_S_TYPE,
	"sh": RV_INST_S_TYPE,
	"sw": RV_INST_S_TYPE,
	"sd": RV_INST_S_TYPE,

	"addi":  RV_INST_I_TYPE,
	"slti":  RV_INST_I_TYPE,
	"sltiu": RV_INST_I_TYPE,
	"xori":  RV_INST_I_TYPE,
	"ori":   RV_INST_I_TYPE,
	"andi":  RV_INST_I_TYPE,
	"addiw": RV_INST_I_TYPE,
	"slli":  RV_INST_I_TYPE,
	"srli":  RV_INST_I_TYPE,
	"srai":  RV_INST_I_TYPE,
	"slliw": RV_INST_I_TYPE,
	"srliw": RV_INST_I_TYPE,
	"sraiw": RV_INST_I_TYPE,

	"beq":  RV_INST_B_TYPE,
	"bne":  RV_INST_B_TYPE,
	"blt":  RV_INST_B_TYPE,
	"bgt":  RV_INST_B_TYPE,
	"bltu": RV_INST_B_TYPE,
	"bgtu": RV_INST_B_TYPE,

	"jal":  RV_INST_J_TYPE,
	"jalr": RV_INST_I_TYPE,

	"lui":   RV_INST_U_TYPE,
	"auipc": RV_INST_U_TYPE,

	"ecall": RV_INST_NONE,

	"call": RV_INST_PSEUDO,
}

type RV_OPCODE_TYPE uint32

const (
	RV_OPCODE_LOAD      RV_OPCODE_TYPE = 0x03
	RV_OPCODE_OP_IMM    RV_OPCODE_TYPE = 0x13
	RV_OPCODE_AUIPC     RV_OPCODE_TYPE = 0x17
	RV_OPCODE_OP_IMM_32 RV_OPCODE_TYPE = 0x1b
	RV_OPCODE_STORE     RV_OPCODE_TYPE = 0x23
	RV_OPCODE_OP        RV_OPCODE_TYPE = 0x33
	RV_OPCODE_LUI       RV_OPCODE_TYPE = 0x37
	RV_OPCODE_OP_32     RV_OPCODE_TYPE = 0x3b
	RV_OPCODE_BRANCH    RV_OPCODE_TYPE = 0x63
	RV_OPCODE_JALR      RV_OPCODE_TYPE = 0x67
	RV_OPCODE_JAL       RV_OPCODE_TYPE = 0x6f
	RV_OPCODE_SYSTEM    RV_OPCODE_TYPE = 0x73
)

var mnem2opcode = map[string]RV_OPCODE_TYPE{
	"add":  RV_OPCODE_OP,
	"sub":  RV_OPCODE_OP,
	"sll":  RV_OPCODE_OP,
	"slt":  RV_OPCODE_OP,
	"sltu": RV_OPCODE_OP,
	"xor":  RV_OPCODE_OP,
	"srl":  RV_OPCODE_OP,
	"sra":  RV_OPCODE_OP,
	"or":   RV_OPCODE_OP,
	"and":  RV_OPCODE_OP,
	"addw": RV_OPCODE_OP_32,
	"subw": RV_OPCODE_OP_32,
	"sllw": RV_OPCODE_OP_32,
	"srlw": RV_OPCODE_OP_32,
	"sraw": RV_OPCODE_OP_32,

	"lb":  RV_OPCODE_LOAD,
	"lbu": RV_OPCODE_LOAD,
	"lh":  RV_OPCODE_LOAD,
	"lhu": RV_OPCODE_LOAD,
	"lw":  RV_OPCODE_LOAD,
	"lwu": RV_OPCODE_LOAD,
	"ld":  RV_OPCODE_LOAD,

	"sb": RV_OPCODE_STORE,
	"sh": RV_OPCODE_STORE,
	"sw": RV_OPCODE_STORE,
	"sd": RV_OPCODE_STORE,

	"addi":  RV_OPCODE_OP_IMM,
	"slti":  RV_OPCODE_OP_IMM,
	"sltiu": RV_OPCODE_OP_IMM,
	"xori":  RV_OPCODE_OP_IMM,
	"ori":   RV_OPCODE_OP_IMM,
	"andi":  RV_OPCODE_OP_IMM,
	"slli":  RV_OPCODE_OP_IMM,
	"srli":  RV_OPCODE_OP_IMM,
	"srai":  RV_OPCODE_OP_IMM,
	"addiw": RV_OPCODE_OP_IMM_32,
	"slliw": RV_OPCODE_OP_IMM_32,
	"srliw": RV_OPCODE_OP_IMM_32,
	"sraiw": RV_OPCODE_OP_IMM_32,

	"beq":  RV_OPCODE_BRANCH,
	"bne":  RV_OPCODE_BRANCH,
	"blt":  RV_OPCODE_BRANCH,
	"bgt":  RV_OPCODE_BRANCH,
	"bltu": RV_OPCODE_BRANCH,
	"bgtu": RV_OPCODE_BRANCH,

	"jal":  RV_OPCODE_JAL,
	"jalr": RV_OPCODE_JALR,

	"lui":   RV_OPCODE_LUI,
	"auipc": RV_OPCODE_AUIPC,

	"ecall": RV_OPCODE_SYSTEM,
}

var bits2reg = map[byte]string{
	0x00: "zero",
	0x01: "ra",
	0x02: "sp",
	0x03: "gp",
	0x04: "tp",
	0x05: "t0",
	0x06: "t1",
	0x07: "t2",
	0x08: "fp",
	0x09: "s1",
	0x0a: "a0",
	0x0b: "a1",
	0x0c: "a2",
	0x0d: "a3",
	0x0e: "a4",
	0x0f: "a5",
	0x10: "a6",
	0x11: "a7",
	0x12: "s2",
	0x13: "s3",
	0x14: "s4",
	0x15: "s5",
	0x16: "s6",
	0x17: "s7",
	0x18: "s8",
	0x19: "s9",
	0x1a: "s10",
	0x1b: "s11",
	0x1c: "t3",
	0x1d: "t4",
	0x1e: "t5",
	0x1f: "t6",
}
var reg2bits = map[string]uint32{
	"x0":   0x00,
	"zero": 0x00,
	"x1":   0x01,
	"ra":   0x01,
	"x2":   0x02,
	"sp":   0x02,
	"x3":   0x03,
	"gp":   0x03,
	"x4":   0x04,
	"tp":   0x04,
	"x5":   0x05,
	"t0":   0x05,
	"x6":   0x06,
	"t1":   0x06,
	"x7":   0x07,
	"t2":   0x07,
	"x8":   0x08,
	"s0":   0x08,
	"fp":   0x08,
	"x9":   0x09,
	"s1":   0x09,
	"x10":  0x0a,
	"a0":   0x0a,
	"x11":  0x0b,
	"a1":   0x0b,
	"x12":  0x0c,
	"a2":   0x0c,
	"x13":  0x0d,
	"a3":   0x0d,
	"x14":  0x0e,
	"a4":   0x0e,
	"x15":  0x0f,
	"a5":   0x0f,
	"x16":  0x10,
	"a6":   0x10,
	"x17":  0x11,
	"a7":   0x11,
	"x18":  0x12,
	"s2":   0x12,
	"x19":  0x13,
	"s3":   0x13,
	"x20":  0x14,
	"s4":   0x14,
	"x21":  0x15,
	"s5":   0x15,
	"x22":  0x16,
	"s6":   0x16,
	"x23":  0x17,
	"s7":   0x17,
	"x24":  0x18,
	"s8":   0x18,
	"x25":  0x19,
	"s9":   0x19,
	"x26":  0x1a,
	"s10":  0x1a,
	"x27":  0x1b,
	"s11":  0x1b,
	"x28":  0x1c,
	"t3":   0x1c,
	"x29":  0x1d,
	"t4":   0x1d,
	"x30":  0x1e,
	"t5":   0x1e,
	"x31":  0x1f,
	"t6":   0x1f,
}

type RV64Inst struct {
	Type   RV_INST_TYPE
	Opcode RV_OPCODE_TYPE
	Inst   interface{}
}

type RV64InstR struct {
	funct3 uint8
	funct7 uint8
	rd     uint8
	rs1    uint8
	rs2    uint8
}

type RV64InstI struct {
	funct3 uint8
	rd     uint8
	rs1    uint8
	imm    uint16
}

type RV64InstS struct {
	funct3 uint8
	rs1    uint8
	rs2    uint8
	imm    uint16
}
type RV64InstB RV64InstS

type RV64InstU struct {
	funct3 uint8
	rd     uint8
	imm    uint32
}
type RV64InstJ RV64InstU

var funct3 = map[string]uint32{
	"add":   0x00,
	"addw":  0x00,
	"addi":  0x00,
	"addiw": 0x00,
	"sub":   0x00,
	"subw":  0x00,
	"sll":   0x01,
	"slli":  0x01,
	"slt":   0x02,
	"slti":  0x02,
	"sltu":  0x03,
	"sltiu": 0x03,
	"xor":   0x04,
	"xori":  0x04,
	"srl":   0x05,
	"srlw":  0x05,
	"srli":  0x05,
	"srliw": 0x05,
	"sra":   0x05,
	"sraw":  0x05,
	"srai":  0x05,
	"sraiw": 0x05,
	"or":    0x06,
	"ori":   0x06,
	"and":   0x07,
	"andi":  0x07,

	"lb":   0x00,
	"lbu":  0x04,
	"sb":   0x00,
	"lh":   0x01,
	"lhu":  0x05,
	"sh":   0x01,
	"lw":   0x02,
	"lwu":  0x06,
	"sw":   0x02,
	"ld":   0x03,
	"sd":   0x03,
	"jalr": 0x00,

	"beq":  0x00,
	"bne":  0x01,
	"blt":  0x04,
	"bgt":  0x05,
	"bltu": 0x06,
	"bgtu": 0x07,
}

func BinToInst(bin []byte) string {

	nm, t := instType(bin)
	if nm == "noimp" {
		return "noimp"
	}
	rd := bits2reg[(bin[1]&0x0f)<<1+bin[0]>>7]
	rs1 := bits2reg[(bin[2]&0x0f)<<1+bin[1]>>7]
	rs2 := bits2reg[(bin[2]&0xf0)>>4+(bin[3]&0x01)<<4]

	switch t {
	case RV_INST_R_TYPE:
		return nm + " " + rd + "," + rs1 + "," + rs2
	case RV_INST_I_TYPE:
		imm := bin[2]>>4 + bin[3]<<4
		return nm + " " + rd + "," + rs1 + "," + strconv.FormatUint(uint64(imm), 16)
	case RV_INST_S_TYPE:
		imm := (bin[3]&0xfe)<<5 + (bin[1]&0x0f)<<1 + bin[0]>>7
		return nm + " " + rs2 + "," + rs1 + "," + strconv.FormatUint(uint64(imm), 16)
	case RV_INST_B_TYPE:
		imm := ((bin[3]&0x80)<<4 + (bin[0]&0x80)<<3 + (bin[3]&0x7e)<<3 + (bin[1] & 0x0f)) << 1
		return nm + " " + rs1 + "," + rs2 + "," + strconv.FormatUint(uint64(imm), 16)
	case RV_INST_U_TYPE:
		imm := uint32(bin[3])<<24 + uint32(bin[2])<<16 + uint32(bin[1]&0xf0)<<8
		return nm + " " + rd + "," + strconv.FormatUint(uint64(imm), 16)
	case RV_INST_J_TYPE:
	}

	return nm
}

func instType(bin []byte) (string, RV_INST_TYPE) {
	var nm string
	var t RV_INST_TYPE

	switch RV_OPCODE_TYPE(bin[0] & 0x7f) {
	case RV_OPCODE_LOAD:
		t = RV_INST_R_TYPE
		switch (bin[1] & 0x70) >> 4 {
		case 0:
			nm = "lb"
		case 1:
			nm = "lh"
		case 2:
			nm = "lw"
		case 3:
			nm = "ld"
		case 4:
			nm = "lbu"
		case 5:
			nm = "lhu"
		case 6:
			nm = "lwu"
		}
		return nm, t
	case RV_OPCODE_OP_IMM:
		t = RV_INST_I_TYPE
		switch (bin[1] & 0x70) >> 4 {
		case 0:
			nm = "addi"
		case 1:
			nm = "slli"
		case 2:
			nm = "slti"
		case 3:
			nm = "sltiu"
		case 4:
			nm = "ori"
		case 5:
			if bin[3]&0x40 == 0 {
				nm = "srli"
			} else {
				nm = "srai"
			}
		case 6:
			nm = "xori"
		case 7:
			nm = "andi"
		}
		return nm, t
	case RV_OPCODE_AUIPC:
		return "auipc", RV_INST_U_TYPE
	case RV_OPCODE_OP_IMM_32:
		t = RV_INST_I_TYPE
		switch (bin[1] & 0x70) >> 4 {
		case 0:
			nm = "addiw"
		case 1:
			nm = "slliw"
		case 5:
			if bin[3]&0x40 == 0 {
				nm = "srliw"
			} else {
				nm = "sraiw"
			}
		}
		return nm, t
	case RV_OPCODE_STORE:
		t = RV_INST_S_TYPE
		switch (bin[1] & 0x70) >> 4 {
		case 0:
			nm = "sb"
		case 1:
			nm = "sh"
		case 2:
			nm = "sw"
		case 3:
			nm = "sd"
		}
		return nm, t
	case RV_OPCODE_OP:
		t = RV_INST_R_TYPE
		switch (bin[1] & 0x70) >> 4 {
		case 0:
			if bin[3]&0x40 == 0 {
				nm = "add"
			} else {
				nm = "sub"
			}
		case 1:
			nm = "sll"
		case 2:
			nm = "slt"
		case 3:
			nm = "sltu"
		case 4:
			nm = "or"
		case 5:
			if bin[3]&0x40 == 0 {
				nm = "srl"
			} else {
				nm = "sra"
			}
		case 6:
			nm = "xor"
		case 7:
			nm = "and"
		}
		return nm, t
	case RV_OPCODE_LUI:
		return "lui", RV_INST_U_TYPE
	case RV_OPCODE_OP_32:
		t = RV_INST_R_TYPE
		switch (bin[1] & 0x70) >> 4 {
		case 0:
			if bin[3]&0x40 == 0 {
				nm = "addw"
			} else {
				nm = "subw"
			}
		case 1:
			nm = "sllw"
		case 5:
			if bin[3]&0x40 == 0 {
				nm = "srlw"
			} else {
				nm = "sraw"
			}
		}
		return nm, t
	case RV_OPCODE_BRANCH:
		t = RV_INST_B_TYPE
		switch (bin[1] & 0x70) >> 4 {
		case 0:
			nm = "beq"
		case 1:
			nm = "bne"
		case 4:
			nm = "blt"
		case 5:
			nm = "bgt"
		case 6:
			nm = "bltu"
		case 7:
			nm = "bgtu"
		}
		return nm, t
	case RV_OPCODE_JALR:
		return "jalr", RV_INST_I_TYPE
	case RV_OPCODE_JAL:
		return "jal", RV_INST_J_TYPE
	case RV_OPCODE_SYSTEM:
		return "ecall", RV_INST_NONE
	}
	return "noimp", RV_INST_NONE
}

func InstToBin(inst []string) ([]byte, elf.R_RISCV) {

	t := mnem2type[inst[0]]
	var op RV_OPCODE_TYPE
	if t != RV_INST_PSEUDO {
		op = mnem2opcode[inst[0]]
	}

	var bits uint32
	switch t {
	case RV_INST_R_TYPE:
		rd := reg2bits[inst[1]]
		rs1 := reg2bits[inst[2]]
		rs2 := reg2bits[inst[3]]
		f3 := funct3[inst[0]]

		var f7 uint32
		if inst[0] == "sub" || inst[0] == "sra" || inst[0] == "sraw" {
			f7 = 0x20
		}

		bits = f7<<25 | rs2<<20 | rs1<<15 | f3<<12 | rd<<7 | uint32(op)

	case RV_INST_I_TYPE:
		var isop int
		if (op == RV_OPCODE_OP_IMM) || (op == RV_OPCODE_OP_IMM_32) {
			isop = 1
		}

		f3 := funct3[inst[0]]
		var issh bool
		if f3 == 0x01 || f3 == 0x05 {
			issh = true
		}

		rd := reg2bits[inst[1]]
		rs1 := reg2bits[inst[3-isop]]
		if issh {
			shamt, _ := strconv.ParseUint(inst[2+isop], 16, 6)

			var f6 uint32
			if inst[0] == "srai" || inst[0] == "sraiw" {
				f6 = 0x10
			} else {
				f6 = 0
			}

			bits = f6<<26 | uint32(shamt)<<20 | rs1<<15 | f3<<12 | rd<<7 | uint32(op)
		} else {
			imm, _ := strconv.ParseInt(inst[2+isop], 10, 12)
			bits = uint32(imm)<<20 | rs1<<15 | f3<<12 | rd<<7 | uint32(op)
		}

	case RV_INST_S_TYPE:
		f3 := funct3[inst[0]]
		rs1 := reg2bits[inst[3]]
		imm, _ := strconv.ParseInt(inst[2], 10, 12)
		rs2 := reg2bits[inst[1]]

		bits = uint32(imm) << 20
		bits &= 0xfe000000
		bits |= (rs2<<20 | rs1<<15 | f3<<12 | (uint32(imm)&0x1f)<<7 | uint32(op))

	case RV_INST_B_TYPE:
		f3 := funct3[inst[0]]
		rs1 := reg2bits[inst[1]]
		rs2 := reg2bits[inst[2]]
		imm, _ := strconv.ParseUint(inst[3], 16, 12)

		fmt.Println(imm)
		fmt.Println(imm & 0x800)
		imm12 := uint32((imm & 0x800) >> 11)
		imm11 := uint32((imm & 0x400) >> 10)
		imm10_5 := uint32((imm & 0x3f0) >> 4)
		imm4_1 := uint32(imm & 0x00f)
		fmt.Println(imm12, imm11, imm10_5, imm4_1)

		bits = imm12<<31 | imm10_5<<25 | rs2<<20 | rs1<<15 | f3<<12 | imm4_1<<8 | imm11<<7 | uint32(op)

	case RV_INST_U_TYPE:
		rd := reg2bits[inst[1]]
		imm, _ := strconv.ParseUint(inst[2], 16, 20)
		bits = uint32(imm)<<12 | rd<<7 | uint32(op)

	case RV_INST_J_TYPE:
	case RV_INST_NONE:
		bits |= uint32(op)
	case RV_INST_PSEUDO:
		if inst[0] == "call" {
			ra, _ := InstToBin([]string{"auipc", "ra", "0"})
			rb, _ := InstToBin([]string{"jalr", "ra", "0", "ra"})
			return append(ra, rb...), elf.R_RISCV_CALL
		}
	}

	ret := make([]byte, 4)
	binary.LittleEndian.PutUint32(ret, bits)
	return ret, elf.R_RISCV_NONE
}
