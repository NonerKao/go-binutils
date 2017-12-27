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
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/NonerKao/go-binutils/rvgc"
)

type sec64 struct {
	header  elf.Section64
	content []string
}

type elf64 struct {
	header   elf.Header64
	sections map[string]*sec64
}

type asUtil struct {
	src     *os.File
	objFile *os.File
	obj     *elf64
	symtab  map[string]*elf.Sym64
}

func New() *asUtil {
	return &asUtil{
		src:     nil,
		objFile: nil,
		obj: &elf64{
			sections: make(map[string]*sec64),
		},
		symtab: make(map[string]*elf.Sym64),
	}
}

var currentOffsetShStr uint32
var currentOffsetStr uint32
var currentSection string

func (asu *asUtil) addSection(sec string) error {
	if asu.obj.sections[sec] != nil {
		return errors.New("Section " + sec + " already exists!")
	} else {
		asu.obj.sections[sec] = new(sec64)
		asu.obj.sections[sec].content = make([]string, 0)
	}
	asu.obj.sections[".shstrtab"].content = append(asu.obj.sections[".shstrtab"].content, sec)
	thisSec := asu.obj.sections[sec]

	thisSec.header.Name = currentOffsetShStr
	switch sec {
	case ".shstrtab":
		thisSec.header.Type = uint32(elf.SHT_STRTAB)
		thisSec.header.Addralign = 0x1
	case ".strtab":
		thisSec.header.Type = uint32(elf.SHT_STRTAB)
		thisSec.header.Addralign = 0x1
	case ".symtab":
		thisSec.header.Type = uint32(elf.SHT_SYMTAB)
		thisSec.header.Addralign = 0x8
		thisSec.header.Entsize = 0x18
	case ".bss":
		thisSec.header.Type = uint32(elf.SHT_NOBITS)
		thisSec.header.Flags = uint64(elf.SHF_ALLOC | elf.SHF_WRITE)
		thisSec.header.Addralign = 0x1
	case ".data":
		thisSec.header.Type = uint32(elf.SHT_PROGBITS)
		thisSec.header.Flags = uint64(elf.SHF_ALLOC | elf.SHF_WRITE)
		thisSec.header.Addralign = 0x1
	case ".text":
		thisSec.header.Type = uint32(elf.SHT_PROGBITS)
		thisSec.header.Flags = uint64(elf.SHF_ALLOC | elf.SHF_EXECINSTR)
		thisSec.header.Addralign = 0x4
	}

	currentOffsetShStr += uint32(len(sec) + 1)
	currentSection = sec
	asu.obj.header.Shnum += 1

	return nil
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

	for _, sec := range internalSection {
		err := asu.addSection(sec)
		if err != nil {
			return err
		}
	}

	return nil
}

func (asu *asUtil) DefineFlags() map[string]interface{} {

	args := map[string]interface{}{
		"o": flag.String("o", "a.out", "Output file name"),
	}

	return args
}

func (asu *asUtil) addLabel(lab string) {
	asu.obj.sections[".strtab"].content = append(asu.obj.sections[".strtab"].content, lab)
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
	var err error
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
		err = asu.addSection(d[1])
		return false, err

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
	asu.obj.sections[currentSection].content = append(asu.obj.sections[currentSection].content, string(rvgc.Cmd2Hex(d)))
}

func (asu *asUtil) write(secname string, align uint64) uint64 {
	var size uint64
	switch secname {
	case ".shstrtab", ".strtab":
		for _, str := range asu.obj.sections[secname].content {
			temp, _ := asu.objFile.WriteString(str)
			asu.objFile.WriteString(" ")
			size = size + uint64(temp) + 1
		}
		fmt.Println("write ", secname, size)
	case ".symtab":
		for _, syment := range asu.symtab {
			var binbuf bytes.Buffer
			binary.Write(&binbuf, binary.LittleEndian, syment)
			temp, _ := asu.objFile.Write(binbuf.Bytes())
			size = size + uint64(temp) + 1
		}
		fmt.Println("write .symtab ", size)
	case ".data":
		for _, data := range asu.obj.sections[secname].content {
			temp, _ := asu.objFile.WriteString(data)
			sizeAlign := uint64(temp) / align * align
			asu.objFile.WriteString(strings.Repeat(" ", int(sizeAlign-size)))
			size += sizeAlign
		}
	case ".text":
		for _, text := range asu.obj.sections[secname].content {
			temp, _ := asu.objFile.WriteString(text)
			size += uint64(temp)
		}
		fmt.Println("write .text ", size)
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
	var binbuf bytes.Buffer
	binary.Write(&binbuf, binary.LittleEndian, &asu.obj.header)
	asu.objFile.Write(binbuf.Bytes())

	var contentOffset uint64 = uint64(asu.obj.header.Shnum*asu.obj.header.Shentsize + asu.obj.header.Ehsize)
	var headerOffset uint64 = uint64(asu.obj.header.Ehsize)
	for name, sec := range asu.obj.sections {
		asu.objFile.Seek(int64(contentOffset), 0)
		sec.header.Size = asu.write(name, sec.header.Addralign)
		contentOffset += sec.header.Size

		asu.objFile.Seek(int64(headerOffset), 0)
		binary.Write(&binbuf, binary.LittleEndian, &sec.header)
		asu.objFile.Write(binbuf.Bytes())
		headerOffset += uint64(asu.obj.header.Shentsize)
	}

	asu.objFile.Close()
	return nil
}
