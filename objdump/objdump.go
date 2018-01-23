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

package objdump

import (
	"debug/elf"
	"flag"
	"fmt"
	"sort"

	"github.com/NonerKao/go-binutils/common"
	"github.com/NonerKao/go-binutils/rvgc"
)

type label struct {
	addr uint64
	name string
}

type objdumpUtil struct {
	file   *elf.File
	raw    []string
	labels []label
}

func New() *objdumpUtil {
	return &objdumpUtil{file: nil, raw: make([]string, 0), labels: make([]label, 0)}
}

func (obu *objdumpUtil) Init(filename string) error {

	var err error
	obu.file, err = common.Init(filename)
	if err != nil {
		return err
	}

	return nil
}

func (obu *objdumpUtil) DefineFlags() map[string]interface{} {

	args := map[string]interface{}{
		"d": flag.Bool("d", false, "disassemble text section"),
	}

	return args
}

func (obu *objdumpUtil) Run(args map[string]interface{}) error {

	if *args["d"].(*bool) {
		var text int

		symtab, _ := obu.file.Symbols()
		for _, s := range symtab {
			if int(s.Section) >= len(obu.file.Sections) {
				continue
			}
			if obu.file.Sections[s.Section].Name == ".text" {
				text = int(s.Section)
				var l label
				l.name = s.Name
				l.addr = s.Value
				if l.name != "" {
					obu.labels = append(obu.labels, l)
				}
			}
		}
		sort.Slice(obu.labels, func(i, j int) bool {
			return obu.labels[i].addr < obu.labels[j].addr
		})
		for _, s := range obu.labels {
			fmt.Println(s)
		}

		bin, _ := obu.file.Sections[text].Data()
		for len(bin) > 0 {
			obu.raw = append(obu.raw, rvgc.BinToInst(bin))
			bin = bin[4:]
		}
	}

	return nil
}

func (obu *objdumpUtil) Output(args map[string]interface{}) error {

	//w := new(tabwriter.Writer)
	//w.Init(os.Stdout, 0, 8, 1, ' ', 0)

	if *args["d"].(*bool) {
		for _, s := range obu.raw {
			fmt.Println(s)
		}
	}

	return nil
}
