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

package nm

import (
	"debug/elf"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"text/tabwriter"

	"github.com/NonerKao/go-binutils/common"
)

type nmUtil struct {
	file *elf.File
	raw  []byte
}

func New() *nmUtil {
	return &nmUtil{file: nil, raw: make([]byte, 0)}
}

func (nmu *nmUtil) Init(filename string) error {

	var err error
	nmu.file, err = common.Init(filename)
	if err != nil {
		return err
	}

	return nil
}

func (nmu *nmUtil) DefineFlags() map[string]interface{} {

	return nil
}

func (nmu *nmUtil) Run(args map[string]interface{}) error {

	symtab, _ := nmu.file.Symbols()
	dynsym, _ := nmu.file.DynamicSymbols()
	syms := append(symtab, dynsym...)

	str := "]"
	for _, d := range syms {
		raw, err := json.Marshal(d)
		if err != nil {
			return err
		}

		str = "," + string(raw) + str
	}
	re, _ := regexp.Compile("^,")
	str = re.ReplaceAllString(str, "[")

	nmu.raw = []byte(str)
	return nil
}

func (nmu *nmUtil) Output(args map[string]interface{}) error {

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, ' ', 0)

	var output []elf.Symbol
	json.Unmarshal(nmu.raw, &output)

	fmt.Fprintln(w, "Offset/Address\tType\tName")
	for _, s := range output {
		fmt.Fprintf(w, "%016x\t%s\t%s\n", s.Value,
			elf.SymType(elf.ST_TYPE(s.Info)).GoString(), s.Name)
	}
	fmt.Fprintln(w)
	w.Flush()

	return nil
}
