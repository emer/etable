// Copyright (c) 2021, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
)

var (
	Output    string
	OutFile   *os.File
	OutWriter *bufio.Writer
	LF        = []byte("\n")
	Delete    bool
	LogPrec   = 4
)

func main() {
	var help bool
	var avg bool
	flag.BoolVar(&help, "help", false, "if true, report usage info")
	flag.BoolVar(&avg, "avg", false, "if true, files must have same cols and rows, outputs average of any float-type columns across files")
	flag.StringVar(&Output, "output", "", "name of output file -- stdout if not specified")
	flag.StringVar(&Output, "o", "", "name of output file -- stdout if not specified")
	flag.BoolVar(&Delete, "delete", false, "if true, delete the source files after cat -- careful!")
	flag.BoolVar(&Delete, "d", false, "if true, delete the source files after cat -- careful!")
	flag.IntVar(&LogPrec, "prec", 4, "precision for number output -- defaults to 4")
	flag.Parse()

	files := flag.Args()

	sort.StringSlice(files).Sort()

	if !avg && Output != "" {
		OutFile, err := os.Create(Output)
		if err != nil {
			fmt.Println("Error creating output file:", err)
			os.Exit(1)
		}
		defer OutFile.Close()
		OutWriter = bufio.NewWriter(OutFile)
	} else {
		OutWriter = bufio.NewWriter(os.Stdout)
	}

	switch {
	case help || len(files) == 0:
		fmt.Printf("\netcat is a data table concatenation utility\n\tassumes all files have header lines, and only retains the header for the first file\n\t(otherwise just use regular cat)\n")
		flag.PrintDefaults()
	case avg:
		AvgCat(files)
	default:
		RawCat(files)
	}
	OutWriter.Flush()
}

func RawCat(files []string) {
	for fi, fn := range files {
		fp, err := os.Open(fn)
		if err != nil {
			fmt.Println("Error opening file: ", err)
			continue
		}
		scan := bufio.NewScanner(fp)
		li := 0
		for {
			if !scan.Scan() {
				break
			}
			ln := scan.Bytes()
			if li == 0 {
				if fi == 0 {
					OutWriter.Write(ln)
					OutWriter.Write(LF)
				}
			} else {
				OutWriter.Write(ln)
				OutWriter.Write(LF)
			}
			li++
		}
		fp.Close()
		if Delete {
			os.Remove(fn)
		}
	}
}

func AvgCat(files []string) {
	dts := make([]*etable.Table, 0, len(files))
	for _, fn := range files {
		dt := &etable.Table{}
		err := dt.OpenCSV(gi.FileName(fn), etable.Tab)
		if err != nil {
			fmt.Println("Error opening file: ", err)
			continue
		}
		dts = append(dts, dt)
	}
	nt := len(dts)
	if nt == 0 {
		return
	}
	ot := dts[0].Clone()
	ot.SetMetaData("precision", strconv.Itoa(LogPrec))
	nr := ot.Rows
	for ci, cl := range ot.Cols {
		if cl.DataType() != etensor.FLOAT32 && cl.DataType() != etensor.FLOAT64 {
			continue
		}
		for di, dt := range dts {
			if di == 0 {
				continue
			}
			dc := dt.Cols[ci]
			for ri := 0; ri < nr; ri++ {
				cv := cl.FloatVal1D(ri)
				cv += dc.FloatVal1D(ri)
				cl.SetFloat1D(ri, cv)
			}
		}
		for ri := 0; ri < nr; ri++ {
			cv := cl.FloatVal1D(ri)
			cv /= float64(nt)
			cl.SetFloat1D(ri, cv)
		}
	}
	ot.SaveCSV(gi.FileName(Output), etable.Tab, etable.Headers)
}
