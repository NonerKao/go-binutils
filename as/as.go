//
// Copyright (C) 2017  Alan (Quey-Liang) Kao  alankao@andestech.com
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

package as

import (
	"bufio"
	"bytes"
	"debug/elf"
	"errors"
	"flag"
	//"fmt"
	"io"
	"os"
	"strings"

	"github.com/NonerKao/go-binutils/rvgc"
)

type elf64 struct {
	header   elf.Header64
	sections []elf.Section64
}

type asUtil struct {
	src     *os.File
	objFile *os.File
	obj     *elf64
	raw     map[string][]string
	symtab  map[string]*elf.Sym64
}

func New() *asUtil {
	return &asUtil{
		src:     nil,
		objFile: nil,
		obj: &elf64{
			sections: make([]elf.Section64, 0, 10),
		},
		raw:    make(map[string][]string),
		symtab: make(map[string]*elf.Sym64),
	}
}

var currentOffsetShStr uint32
var currentOffsetStr uint32
var currentSection string

func (asu *asUtil) addSection(sec string) {
	if asu.raw[sec] == nil {
		asu.raw[sec] = make([]string, 0)
	}
	asu.raw[".shstrtab"] = append(asu.raw[".shstrtab"], sec)

	asu.obj.sections[asu.obj.header.Shnum].Name = currentOffsetShStr
	switch sec {
	case ".shstrtab":
		asu.obj.sections[asu.obj.header.Shnum].Type = uint32(elf.SHT_STRTAB)
		asu.obj.sections[asu.obj.header.Shnum].Addralign = 0x1
	case ".strtab":
		asu.obj.sections[asu.obj.header.Shnum].Type = uint32(elf.SHT_STRTAB)
		asu.obj.sections[asu.obj.header.Shnum].Addralign = 0x1
	case ".symtab":
		asu.obj.sections[asu.obj.header.Shnum].Type = uint32(elf.SHT_SYMTAB)
		asu.obj.sections[asu.obj.header.Shnum].Addralign = 0x8
		asu.obj.sections[asu.obj.header.Shnum].Entsize = 0x18
	case ".bss":
		asu.obj.sections[asu.obj.header.Shnum].Type = uint32(elf.SHT_NOBITS)
		asu.obj.sections[asu.obj.header.Shnum].Flags = uint64(elf.SHF_ALLOC | elf.SHF_WRITE)
		asu.obj.sections[asu.obj.header.Shnum].Addralign = 0x1
	case ".data":
		asu.obj.sections[asu.obj.header.Shnum].Type = uint32(elf.SHT_PROGBITS)
		asu.obj.sections[asu.obj.header.Shnum].Flags = uint64(elf.SHF_ALLOC | elf.SHF_WRITE)
		asu.obj.sections[asu.obj.header.Shnum].Addralign = 0x1
	case ".text":
		asu.obj.sections[asu.obj.header.Shnum].Type = uint32(elf.SHT_PROGBITS)
		asu.obj.sections[asu.obj.header.Shnum].Flags = uint64(elf.SHF_ALLOC | elf.SHF_EXECINSTR)
		asu.obj.sections[asu.obj.header.Shnum].Addralign = 0x4
	}

	currentOffsetShStr += uint32(len(sec) + 1)
	currentSection = sec
	asu.obj.header.Shnum += 1
}

func (asu *asUtil) Init(filename string) error {

	var err error
	asu.src, err = os.Open(filename)
	if err != nil {
		return err
	}

	asu.obj.header.Ident[1] = '\x7f'
	asu.obj.header.Ident[1] = 'E'
	asu.obj.header.Ident[2] = 'L'
	asu.obj.header.Ident[3] = 'F'
	asu.obj.header.Ident[4] = byte(elf.ELFCLASS64)
	asu.obj.header.Ident[5] = byte(elf.ELFDATA2LSB)
	asu.obj.header.Ident[6] = byte(elf.EV_CURRENT)
	asu.obj.header.Type = uint16(elf.ET_REL)
	asu.obj.header.Machine = 243 //elf.EM_RISCV
	asu.obj.header.Version = uint32(elf.EV_CURRENT)
	asu.obj.header.Ehsize = 64

	asu.obj.header.Shentsize = 64
	asu.obj.header.Shnum = 0
	asu.obj.header.Shstrndx = 0

	asu.raw[".shstrtab"] = make([]string, 6)
	asu.addSection(".shstrtab")
	asu.addSection(".strtab")
	asu.addSection(".symtab")

	return nil
}

func (asu *asUtil) DefineFlags() map[string]interface{} {

	args := map[string]interface{}{
		"o": flag.String("o", "a.out", "Output file name"),
	}

	return args
}

func (asu *asUtil) addLabel(lab string) {
	asu.raw[".strtab"] = append(asu.raw[".strtab"], lab)
	asu.symtab[lab] = &elf.Sym64{
		Name:  currentOffsetStr,
		Info:  elf.ST_INFO(elf.STB_LOCAL, elf.STT_FUNC),
		Shndx: asu.obj.header.Shnum,
	}

	currentOffsetStr += uint32(len(lab) + 1)
}

func (asu *asUtil) Run(args map[string]interface{}) error {

	r := bufio.NewReaderSize(asu.src, 1024)
	line, _, err := r.ReadLine()
	end := false
	for err == nil {
		sa := strings.Split(string(line), " ")

		if sa[0][0] == '.' {
			end, err = asu.dire(sa)
			if end {
				break
			}
		} else if sa[0][len(sa[0])-1] == ':' {
			asu.addLabel(sa[0][0 : len(sa[0])-1])
		} else {
			asu.inst(sa)
		}

		line, _, err = r.ReadLine()
	}

	if end == false && err != io.EOF {
		return err
	}

	asu.src.Close()
	return nil
}

var internalSection = []string{
	".symtab",
	".strtab",
	".shstrtab",
}

func (asu *asUtil) dire(d []string) (bool, error) {
	switch d[0] {
	case ".section":
		if len(d) != 2 {
			return false, errors.New("Syntax error: section not specified!")
		}
		for _, sec := range internalSection {
			if d[1] == sec {
				return false, errors.New("Syntax error: not allowed section " + d[1])
			}
		}
		//addSection(d[1])

	case ".global":
		if len(d) != 2 {
			return false, errors.New("Syntax error: label not specified!")
		}
		asu.symtab["add"].Info = byte(elf.ST_INFO(elf.STB_GLOBAL, elf.STT_FUNC))

	case ".end":
		return true, nil
	}
	return false, nil
}

func (asu *asUtil) inst(d []string) {
	asu.raw[currentSection] = append(asu.raw[currentSection], string(rvgc.Cmd2Hex(d)))
}

func (asu *asUtil) write(secname string, align uint64) uint64 {
	var size uint64
	switch secname {
	case ".shstrtab", ".strtab":
		for _, str := range asu.raw[secname] {
			temp, _ := asu.objFile.WriteString(str)
			asu.objFile.WriteString(" ")
			size = size + uint64(temp) + 1
		}
	case ".symtab":
		for _, syment := range asu.symtab {
			temp, _ := asu.objFile.Write(syment)
			size = size + uint64(temp) + 1
		}
	case ".data":
		for _, data := range asu.raw[secname] {
			temp, _ := asu.objFile.WriteString(data)
			sizeAlign := uint64(temp) / align * align
			asu.objFile.WriteString(strings.Repeat(" ", int(sizeAlign-size)))
			size += sizeAlign
		}
	case ".text":
		for _, text := range asu.raw[secname] {
			temp, _ := asu.objFile.WriteString(text)
			size += uint64(temp)
		}
	}
	return size
}

func (asu *asUtil) Output(args map[string]interface{}) error {

	var err error
	asu.objFile, err = os.Create(*args["o"].(*string))
	if err != nil {
		return err
	}

	asu.obj.header.Shoff = uint64(asu.obj.header.Ehsize)
	asu.objFile.Write(asu.obj.header)

	var payloadOffset uint64 = uint64(uint16(len(asu.obj.sections))*asu.obj.header.Shentsize + asu.obj.header.Ehsize)
	var headerOffset uint64 = uint64(asu.obj.header.Ehsize)
	for i, sec := range asu.obj.sections {
		asu.objFile.Seek(int64(payloadOffset), 0)
		tail := asu.raw[".shstrtab"][sec.Name:]
		end := bytes.IndexByte([]byte(tail), 0)
		name := tail[:end]
		sec.Size = asu.write(name, sec.Addralign)
		payloadOffset += sec.Size

		asu.objFile.Seek(int64(headerOffset), 0)
		asu.objFile.Write(sec)
		headerOffset += uint64(asu.obj.header.Shentsize)
	}

	asu.objFile.Close()
	return nil
}
