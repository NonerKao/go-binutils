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
	"encoding/json"
	"flag"
	"fmt"
	"os"
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

func (reu *readelfUtil) Init(fileName string) error {

	var err error
	reu.file, err = common.Init(fileName)
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

	return nil
}

func (reu *readelfUtil) Output(args map[string]interface{}) error {

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)

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

	return nil
}
