//
// Copyright (C) 2017  Quey-Liang Kao  s101062801@m101.nthu.edu.tw
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
	"fmt"
)

type ReadELFUtil struct {
	Data ReadELFData
}

type ReadELFData struct {
	Section string
}

func Init() *ReadELFUtil {
	fmt.Println("Init")
	return &ReadELFUtil{Data: ReadELFData{Section: "Test 1"}}
}

func (re *ReadELFUtil) Run(args map[string]*string) error {
	fmt.Println(args)
	fmt.Println(args["l"])
	fmt.Println(*args["h"])
	fmt.Println(*args["l"])
	return nil
}

func (re *ReadELFUtil) Output() error {
	fmt.Println("Out")
	return nil
}
