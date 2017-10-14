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

package main

// main.go: The unified entry point of go-binutils project

import (
	"flag"
	"fmt"
	"os"

	"./common"
	"./readelf"
)

var (
	util common.Util
)

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage of go-binutils: %s\n", os.Args[0])
		fmt.Printf("    example7 file1 file2 ...\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	util := readelf.Init()
	util.Run(make(map[string]string))
	util.Output()
}
