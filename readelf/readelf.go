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
	"flag"
	"fmt"

	"github.com/NonerKao/go-binutils/common"
)

type readelfUtil struct {
	file *elf.File
}

func Init(fileName string) (*readelfUtil, error) {

	var reu readelfUtil
	var err error
	reu.file, err = common.Init(fileName)
	if err != nil {
		return nil, err
	}

	return &reu, nil
}

func (reu *readelfUtil) DefineFlags() map[string]interface{} {

	args := map[string]interface{}{
		"h": flag.Bool("h", false, "Show file header"),
		"l": flag.Bool("l", false, "Show program headers"),
		"S": flag.Bool("S", false, "Show section headers"),
	}

	return args
}

func (reu *readelfUtil) Run(args map[string]interface{}) (string, error) {

	fmt.Println(*args["S"].(*bool))
	fmt.Println(*args["h"].(*bool))
	fmt.Println(*args["l"].(*bool))

	return "nothing yet", nil
}
