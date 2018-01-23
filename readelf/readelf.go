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

package readelf

import (
	"debug/elf"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"text/tabwriter"

	"github.com/NonerKao/go-binutils/common"
)

type readelfUtil struct {
	file *elf.File
	raw  map[string][]byte
}

func New() *readelfUtil {
	return &readelfUtil{file: nil, raw: make(map[string][]byte)}
}

func (reu *readelfUtil) Init(filename string) error {

	var err error
	reu.file, err = common.Init(filename)
	if err != nil {
		return err
	}

	return nil
}

func (reu *readelfUtil) DefineFlags() map[string]interface{} {

	args := map[string]interface{}{
		"h": flag.Bool("h", false, "Show file header"),
		"l": flag.Bool("l", false, "Show program headers"),
		"S": flag.Bool("S", false, "Show section headers"),
		"r": flag.Bool("r", false, "Show relocation sections"),
	}

	return args
}

func (reu *readelfUtil) Run(args map[string]interface{}) error {

	if *args["h"].(*bool) {
		raw, err := json.Marshal(reu.file.FileHeader)
		if err != nil {
			return err
		}

		reu.raw["h"] = raw
	}

	if *args["l"].(*bool) {
		str := "]"
		for _, p := range reu.file.Progs {
			raw, err := json.Marshal(p)
			if err != nil {
				return err
			}

			str = "," + string(raw) + str
		}
		re, _ := regexp.Compile("^,")
		str = re.ReplaceAllString(str, "[")

		reu.raw["l"] = []byte(str)
	}

	if *args["S"].(*bool) {
		str := "]"
		for _, p := range reu.file.Sections {
			raw, err := json.Marshal(p)
			if err != nil {
				return err
			}

			str = "," + string(raw) + str
		}
		re, _ := regexp.Compile("^,")
		str = re.ReplaceAllString(str, "[")

		reu.raw["S"] = []byte(str)
	}

	if *args["r"].(*bool) {
		var index int
		for i, p := range reu.file.Sections {
			if p.Name == ".rela.text" {
				index = i
				break
			}
		}

		var rela elf.Rela64
		str := "]"
		var end error
		r := reu.file.Sections[index].Open()
		for ; end == nil; end = binary.Read(r, binary.LittleEndian, &rela) {

			raw, err := json.Marshal(rela)
			if err != nil {
				return err
			}

			str = "," + string(raw) + str
		}
		re, _ := regexp.Compile("^,")
		str = re.ReplaceAllString(str, "[")

		reu.raw["r"] = []byte(str)
	}

	return nil
}

func (reu *readelfUtil) Output(args map[string]interface{}) error {

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, ' ', 0)

	if *args["h"].(*bool) {
		var output elf.FileHeader
		/* err := */ json.Unmarshal(reu.raw["h"], &output)
		/*if err != nil {
			return err
		}*/

		fmt.Fprintln(w, "ELF File Header:\t\t")
		fmt.Fprintln(w, "\tClass:\t", output.Class.GoString())
		fmt.Fprintln(w, "\tData:\t", output.Data.GoString())
		fmt.Fprintln(w, "\tOSABI:\t", output.OSABI.GoString())
		fmt.Fprintln(w, "\tABIVersion:\t", output.ABIVersion)
		fmt.Fprintln(w, "\tType:\t", output.Type.GoString())
		fmt.Fprintln(w, "\tMachine:\t", output.Machine.GoString())
		fmt.Fprintln(w)
		w.Flush()
	}

	if *args["l"].(*bool) {
		var output []elf.ProgHeader
		json.Unmarshal(reu.raw["l"], &output)

		fmt.Fprintln(w, "Program Header:\t\t\t\t\t\t\t")
		fmt.Fprintln(w, "Number\tType\tFlags\tOffset\tAddress\tFile Size\tMemory Size\tAlignment")
		for i, p := range output {
			fmt.Fprintf(w, "%d\t%s\t%s\t%x\t%x\t%d\t%d\t0x%x\n",
				i, p.Type.GoString(), p.Flags.GoString(),
				p.Off, p.Vaddr, p.Filesz, p.Memsz, p.Align)
		}
		fmt.Fprintln(w)
		w.Flush()
	}

	if *args["S"].(*bool) {
		var output []elf.SectionHeader
		json.Unmarshal(reu.raw["S"], &output)

		fmt.Fprintln(w, "Section Header:\t\t\t\t\t\t\t\t\t")
		fmt.Fprintln(w, "Number\tName\tType\tFlags\tAddress\tOffset\tSize\tLink\tInfo\tAlignment")
		for i, s := range output {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%x\t%x\t%d\t%d\t%d\t0x%x\n",
				i, s.Name, s.Type.GoString(), s.Flags.GoString(),
				s.Addr, s.Offset, s.Size, s.Link, s.Info, s.Addralign)
		}
		fmt.Fprintln(w)
		w.Flush()
	}

	if *args["r"].(*bool) {
		var output []elf.Rela64
		json.Unmarshal(reu.raw["r"], &output)

		fmt.Fprintln(w, "Relocation:\t\t\t")
		fmt.Fprintln(w, "Offset\tType\tSymbol\tAppend")

		syms, _ := reu.file.Symbols()

		for _, s := range output {
			if elf.R_SYM64(s.Info) == 0 {
				continue
			}

			fmt.Fprintf(w, "%016x\t%s\t%s\t%d\n",
				s.Off, elf.R_RISCV(elf.R_TYPE64(s.Info)).GoString(), syms[elf.R_SYM64(s.Info)-1].Name, s.Addend)
		}
		fmt.Fprintln(w)
		w.Flush()
	}

	return nil
}
