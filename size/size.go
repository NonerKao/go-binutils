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

package size

import (
	"debug/elf"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"text/tabwriter"

	"github.com/NonerKao/go-binutils/common"
)

type sizeUtil struct {
	file *elf.File
	raw  []byte
}

func New() *sizeUtil {
	return &sizeUtil{file: nil, raw: make([]byte, 0)}
}

func (siu *sizeUtil) Init(filename string) error {

	var err error
	siu.file, err = common.Init(filename)
	if err != nil {
		return err
	}

	return nil
}

func (siu *sizeUtil) DefineFlags() map[string]interface{} {

	return nil
}

func (siu *sizeUtil) Run(args map[string]interface{}) error {

	str := "]"
	for _, p := range siu.file.Sections {
		raw, err := json.Marshal(p)
		if err != nil {
			return err
		}

		str = "," + string(raw) + str
	}
	re, _ := regexp.Compile("^,")
	str = re.ReplaceAllString(str, "[")

	siu.raw = []byte(str)
	return nil
}

func (siu *sizeUtil) Output(args map[string]interface{}) error {

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, ' ', tabwriter.AlignRight)

	var output []elf.SectionHeader
	json.Unmarshal(siu.raw, &output)

	fmt.Fprintln(w, "Name\tSize\tAddress\tOffset\t")
	for _, s := range output {
		fmt.Fprintf(w, "%s\t%d(%x)\t%x\t%x\t\n",
			s.Name, s.Size, s.Size, s.Addr, s.Offset)
	}
	fmt.Fprintln(w)
	w.Flush()

	return nil
}
