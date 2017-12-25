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
	"debug/elf"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
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
}

func New() *asUtil {
	return &asUtil{
		src:     nil,
		objFile: nil,
		obj: &elf64{
			sections: make([]elf.Section64, 0, 10),
		},
		raw: make(map[string][]string),
	}
}

var currentOffsetShStr = 0
var currentSection = ""

func (reu *asUtil) addSection(sec string) {
	if reu.obj.header.Shnum < 6 {
		reu.raw[".shstrtab"][reu.obj.header.Shnum] = sec
	} else {
		append(reu.raw[".shstrtab"], sec)
	}

	reu.obj.sections[reu.obj.header.Shnum].Name = currentOffsetShStr
	switch sec {
	case ".shstrtab":
		reu.obj.sections[reu.obj.header.Shnum].Type = elf.SHT_STRTAB
		reu.obj.sections[reu.obj.header.Shnum].Addralign = 0x1
	case ".strtab":
		reu.obj.sections[reu.obj.header.Shnum].Type = elf.SHT_STRTAB
		reu.obj.sections[reu.obj.header.Shnum].Addralign = 0x1
	case ".symtab":
		reu.obj.sections[reu.obj.header.Shnum].Type = elf.SHT_SYMTAB
		reu.obj.sections[reu.obj.header.Shnum].Addralign = 0x8
	case ".bss":
		reu.obj.sections[reu.obj.header.Shnum].Type = elf.SHT_NOBITS
		reu.obj.sections[reu.obj.header.Shnum].Flags = elf.SHF_ALLOC | elf.SHF_WRITE
		reu.obj.sections[reu.obj.header.Shnum].Addralign = 0x1
	case ".data":
		reu.obj.sections[reu.obj.header.Shnum].Type = elf.SHT_PROGBITS
		reu.obj.sections[reu.obj.header.Shnum].Flags = elf.SHF_ALLOC | elf.SHF_WRITE
		reu.obj.sections[reu.obj.header.Shnum].Addralign = 0x1
	case ".text":
		reu.obj.sections[reu.obj.header.Shnum].Type = elf.SHT_PROGBITS
		reu.obj.sections[reu.obj.header.Shnum].Flags = elf.SHF_ALLOC | elf.SHF_EXECINSTR
		reu.obj.sections[reu.obj.header.Shnum].Addralign = 0x4
	}

	currentOffsetShStr += len(sec) + 1
	currentSection = sec
	reu.obj.header.Shnum += 1
}

func (reu *asUtil) Init(filename string) error {

	var err error
	reu.src, err = os.Open(filename)
	if err != nil {
		return err
	}

	reu.obj.header.Ident[1] = '\x7f'
	reu.obj.header.Ident[1] = 'E'
	reu.obj.header.Ident[2] = 'L'
	reu.obj.header.Ident[3] = 'F'
	reu.obj.header.Ident[4] = elf.ELFCLASS64
	reu.obj.header.Ident[5] = elf.ELFDATA2LSB
	reu.obj.header.Ident[6] = elf.EV_CURRENT
	reu.obj.header.Type = elf.ET_REL
	reu.obj.header.Machine = elf.EM_RISCV
	reu.obj.header.Version = elf.EV_CURRENT
	reu.obj.header.Ehsize = 64

	reu.obj.header.Shentsize = 64
	reu.obj.header.Shnum = 0
	reu.obj.header.Shstrndx = 0

	reu.raw[".shstrtab"] = make([]string, 6)
	addSection(".shstrtab")
	addSection(".strtab")
	addSection(".symtab")
	addSection(".data")
	addSection(".bss")
	addSection(".text")

	return nil
}

func (reu *asUtil) DefineFlags() map[string]interface{} {

	args := map[string]interface{}{
		"o": flag.String("o", "a.out", "Output file name"),
	}

	return args
}

func (reu *asUtil) Run(args map[string]interface{}) error {

	r := bufio.NewReaderSize(reu.src, 1024)
	line, _, err := r.ReadLine()
	end := false
	for err == nil {
		sa := strings.Split(string(line), " ")

		if sa[0][0] == '.' {
			end, err = dire(sa)
			if end {
				break
			}
		} else if sa[0][len(sa[0])-1] == ':' {
			label(sa[0][0 : len(sa[0])-1])
		} else {
			inst(sa)
		}

		line, _, err = r.ReadLine()
	}

	if end == false && err != io.EOF {
		return err
	}

	reu.src.Close()
	return nil
}

var internalSection = []string{
	".symtab",
	".strtab",
	".shstrtab",
}

func (reu *asUtil) dire(d []string) (bool, error) {
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
		setLabel(d[1])

	case ".end":
		return true, nil
	}
	return false, nil
}

func (reu *asUtil) label(d string) {
	fmt.Println("label", d)
}

func (reu *asUtil) inst(d []string) {
	fmt.Println("inst:", d[0])
}

func (reu *asUtil) Output(args map[string]interface{}) error {

	var err error
	reu.obj, err = os.Open(*args["o"].(*string))
	if err != nil {
		return err
	}

	reu.obj.Close()
	return nil
}
