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

package main

// main.go: The unified entry point

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/NonerKao/go-binutils/as"
	"github.com/NonerKao/go-binutils/common"
	"github.com/NonerKao/go-binutils/nm"
	"github.com/NonerKao/go-binutils/objdump"
	"github.com/NonerKao/go-binutils/readelf"
	"github.com/NonerKao/go-binutils/size"
)

func main() {

	util, err := route()
	if err != nil {
		fmt.Println(err.Error())
		printUsage()
		return
	}

	args := util.DefineFlags()
	flag.Usage = printUsage
	flag.Parse()

	tail := flag.Args()
	err1 := util.Init(tail[len(tail)-1])
	if err1 != nil {
		fmt.Println(err1.Error())
		return
	}

	err2 := util.Run(args)
	if err2 != nil {
		fmt.Println(err2.Error())
		return
	}

	err3 := util.Output(args)
	if err3 != nil {
		fmt.Println(err3.Error())
		return
	}

}

func route() (common.Util, error) {

	var util common.Util
	switch {
	case strings.HasSuffix(os.Args[0], "as"):
		util = as.New()
	case strings.HasSuffix(os.Args[0], "nm"):
		util = nm.New()
	case strings.HasSuffix(os.Args[0], "objdump"):
		util = objdump.New()
	case strings.HasSuffix(os.Args[0], "readelf"):
		util = readelf.New()
	case strings.HasSuffix(os.Args[0], "size"):
		util = size.New()
	default:
		return nil, errors.New("No such usage!")
	}
	return util, nil
}

func printUsage() {
	fmt.Printf("Usage of go-binutils: %s\n", os.Args[0])
	fmt.Printf("    example7 file1 file2 ...\n")
	flag.PrintDefaults()
}
