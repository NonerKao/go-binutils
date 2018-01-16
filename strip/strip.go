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

package strip

import (
	"bufio"
	"bytes"
	"debug/elf"
	"encoding/binary"
	"errors"
	"flag"
	"io"
	"os"
	"regexp"
	"strings"
)

type stripUtil struct {
	src     *elf.FILE
	objFile *os.File
	obj     *elf64
	symtab  []*elf.Sym64
	shOrder []string
}

func New() *asUtil {
	return &asUtil{
		src:     nil,
		objFile: nil,
		obj: &elf64{
			sections: make(map[string]*sec64),
		},
		symtab:  make([]*elf.Sym64, 0),
		shOrder: make([]string, 0),
	}
}

var currentOffsetShStr uint32
var currentOffsetStr uint32
var currentSection string

var internalSection = []string{
	"",
	".shstrtab",
	".strtab",
	".symtab",
}

func (asu *asUtil) addSection(sec string) error {
	if asu.obj.sections[sec] != nil {
		return errors.New("Section " + sec + " already exists!")
	} else {
		asu.obj.sections[sec] = new(sec64)
		asu.obj.sections[sec].content = make([]string, 0)
	}

	asu.shOrder = append(asu.shOrder, sec)

	thisSec := asu.obj.sections[sec]
	thisSec.header.Name = currentOffsetShStr
	switch sec {
	case "":
		thisSec.header.Type = uint32(elf.SHT_NULL)
		asu.symtab = append(asu.symtab, &elf.Sym64{
			Name:  0,
			Info:  elf.ST_INFO(elf.STB_LOCAL, elf.STT_NOTYPE),
			Shndx: 0,
		})
	case ".shstrtab":
		thisSec.header.Type = uint32(elf.SHT_STRTAB)
		thisSec.header.Addralign = 0x1
		thisSec.content = append(thisSec.content, "")
		currentOffsetShStr = 1
	case ".strtab":
		thisSec.header.Type = uint32(elf.SHT_STRTAB)
		thisSec.header.Addralign = 0x1
		thisSec.content = append(thisSec.content, "")
		currentOffsetStr = 1
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
		asu.symtab = append(asu.symtab, &elf.Sym64{
			Info:  elf.ST_INFO(elf.STB_LOCAL, elf.STT_SECTION),
			Shndx: asu.obj.header.Shnum,
		})
	}

	if sec != "" {
		asu.obj.sections[".shstrtab"].content = append(asu.obj.sections[".shstrtab"].content, sec)
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

	asu.obj.header.Ident[0] = '\x7f'
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
	asu.obj.header.Flags = uint32(0x0005)

	asu.obj.header.Shentsize = 64
	asu.obj.header.Shnum = 0
	asu.obj.header.Shstrndx = 1

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
	asu.symtab = append(asu.symtab, &elf.Sym64{
		Name:  currentOffsetStr,
		Info:  elf.ST_INFO(elf.STB_GLOBAL, elf.STT_FUNC),
		Shndx: asu.obj.header.Shnum - 1,
	})

	currentOffsetStr += uint32(len(lab) + 1)
}

func preProcessLine(line string) []string {

	rePunc := regexp.MustCompile(`[,()]`)
	reSpace := regexp.MustCompile(`[[:space:]]+`)
	line = rePunc.ReplaceAllString(line, " ")
	line = reSpace.ReplaceAllString(line, " ")
	sa := strings.Split(strings.TrimSpace(string(line)), " ")

	return sa
}

func (asu *asUtil) Run(args map[string]interface{}) error {

	r := bufio.NewReaderSize(asu.src, 1024)
	line, _, err := r.ReadLine()
	end := false
	for err == nil {
		sa := preProcessLine(string(line))

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
		//asu.symtab["add"].Info = byte(elf.ST_INFO(elf.STB_GLOBAL, elf.STT_NOTYPE))

	case ".end":
		return true, nil
	}
	return false, nil
}

func (asu *asUtil) inst(d []string) {
	asu.obj.sections[currentSection].content = append(asu.obj.sections[currentSection].content, string(rvgc.InstToBin(d)))
}

func (asu *asUtil) write(secname string, align uint64) uint64 {
	var size uint64
	switch secname {
	case ".shstrtab", ".strtab":
		for _, str := range asu.obj.sections[secname].content {
			temp, _ := asu.objFile.WriteString(str)
			asu.objFile.WriteString("\x00")
			size = size + uint64(temp) + 1
		}
	case ".symtab":
		for _, syment := range asu.symtab {
			var binbuf bytes.Buffer
			binary.Write(&binbuf, binary.LittleEndian, syment)
			temp, _ := asu.objFile.Write(binbuf.Bytes())
			size = size + uint64(temp)
		}
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
	for i, name := range asu.shOrder {
		sec := asu.obj.sections[name]

		asu.objFile.Seek(int64(contentOffset), 0)
		sec.header.Size += asu.write(name, sec.header.Addralign)
		sec.header.Off = uint64(contentOffset)
		contentOffset += sec.header.Size

		if name == ".strtab" {
			asu.obj.sections[".symtab"].header.Link = uint32(i)
			asu.obj.sections[".symtab"].header.Info = 2
		}

		asu.objFile.Seek(int64(headerOffset), 0)
		var binbuf bytes.Buffer
		binary.Write(&binbuf, binary.LittleEndian, &sec.header)
		asu.objFile.Write(binbuf.Bytes())
		headerOffset += uint64(asu.obj.header.Shentsize)
	}

	asu.objFile.Close()
	return nil
}
