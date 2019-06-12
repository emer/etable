// Copyright (c) 2019, The eTable Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etable

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
)

const (
	// Tab is the tab rune delimiter, for TSV tab separated values
	Tab rune = '\t'

	// Comma is the comma rune delimiter, for CSV comma separated values
	Comma rune = ','

	// Space is the space rune delimiter, for SSV space separated value
	Space rune = ' '
)

// SaveCSV writes a table to a comma-separated-values (CSV) file (where comma = any delimiter,
// specified in the delim arg).
// If headers = true then generate C++ emergent-tyle column headers and add _H: to the header line
// and _D: to the data lines.  These headers have full configuration information for the tensor
// columns.  Otherwise, only the data is written.
func (dt *Table) SaveCSV(filename gi.FileName, delim rune, headers bool) error {
	fp, err := os.Create(string(filename))
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	dt.WriteCSV(fp, delim, headers)
	return nil
}

// OpenCSV reads a table from a comma-separated-values (CSV) file (where comma = any delimiter,
// specified in the delim arg), using the Go standard encoding/csv reader conforming
// to the official CSV standard.
// If the table does not currently have any columns, the first row of the file is assumed to be
// headers, and columns are constructed therefrom.  We parse the C++ emergent column
// headers, if the first line starts with _H: -- these have full configuration information for tensor
// dimensionality, and are also supported for writing using WriteCSV.
// If the table DOES have existing columns, then those are used robustly for whatever information
// fits from each row of the file.
func (dt *Table) OpenCSV(filename gi.FileName, delim rune) error {
	fp, err := os.Open(string(filename))
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	return dt.ReadCSV(fp, delim)
}

// ReadCSV reads a table from a comma-separated-values (CSV) file (where comma = any delimiter,
// specified in the delim arg), using the Go standard encoding/csv reader conforming
// to the official CSV standard.
// If the table does not currently have any columns, the first row of the file is assumed to be
// headers, and columns are constructed therefrom.  We parse the C++ emergent column
// headers, if the first line starts with _H: -- these have full configuration information for tensor
// dimensionality, and are also supported for writing using WriteCSV.
// For non-emergent headers, string-valued columns are constructed first and then
// If the table DOES have existing columns, then those are used robustly for whatever information
// fits from each row of the file.
func (dt *Table) ReadCSV(r io.Reader, delim rune) error {
	cr := csv.NewReader(r)
	if delim != 0 {
		cr.Comma = delim
	}
	rec, err := cr.ReadAll() // todo: lazy, avoid resizing
	if err != nil || len(rec) == 0 {
		return err
	}
	rows := len(rec)
	// cols := len(rec[0])
	strow := 0
	if dt.NumCols() == 0 || rec[0][0] == "_H:" {
		sc, err := SchemaFromHeaders(rec[0], rec)
		if err != nil {
			log.Println(err.Error())
			return err
		}
		strow++
		rows--
		dt.SetFromSchema(sc, rows)
	}
	tc := dt.NumCols()
	dt.SetNumRows(rows)
rowloop:
	for ri := 0; ri < rows; ri++ {
		ci := 0
		rr := rec[ri+strow]
		if rr[0] == "_D:" { // emergent data row
			ci++
		}
		for j := 0; j < tc; j++ {
			tsr := dt.Cols[j]
			_, csz := tsr.RowCellSize()
			stoff := ri * csz
			for cc := 0; cc < csz; cc++ {
				str := rr[ci]
				if str == "" {
					tsr.SetNull1D(stoff+cc, true) // empty = missing
				} else {
					tsr.SetString1D(stoff+cc, str)
				}
				ci++
				if ci >= len(rr) {
					continue rowloop
				}
			}
		}
	}
	return nil
}

// SchemaFromHeaders attempts to configure a Table Schema based on the headers
// for non-Emergent headers, data is examined to
func SchemaFromHeaders(hdrs []string, rec [][]string) (Schema, error) {
	if hdrs[0] == "_H:" {
		return SchemaFromEmerHeaders(hdrs)
	}
	return SchemaFromPlainHeaders(hdrs, rec)
}

// SchemaFromEmerHeaders attempts to configure a Table Schema based on emergent DataTable headers
func SchemaFromEmerHeaders(hdrs []string) (Schema, error) {
	nc := len(hdrs) - 1
	sc := Schema{}
	for ci := 0; ci < nc; ci++ {
		hd := hdrs[ci+1]
		if hd == "" {
			continue
		}
		var typ etensor.Type
		typ, hd = EmerColType(hd)
		dimst := strings.Index(hd, "]<")
		if dimst > 0 {
			dims := hd[dimst+2 : len(hd)-1]
			lbst := strings.Index(hd, "[")
			hd = hd[:lbst]
			csh := ShapeFromString(dims)
			// new tensor starting
			sc = append(sc, Column{Name: hd, Type: etensor.Type(typ), CellShape: csh})
			continue
		}
		dimst = strings.Index(hd, "[")
		if dimst > 0 {
			continue
		}
		sc = append(sc, Column{Name: hd, Type: etensor.Type(typ), CellShape: nil})
	}
	return sc, nil
}

var EmerHdrCharToType = map[byte]etensor.Type{
	'$': etensor.STRING,
	'%': etensor.FLOAT32,
	'#': etensor.FLOAT64,
	'|': etensor.INT64,
	'@': etensor.UINT8,
	'&': etensor.STRING,
	'^': etensor.BOOl,
}

var EmerHdrTypeToChar map[etensor.Type]byte

func init() {
	EmerHdrTypeToChar = make(map[etensor.Type]byte)
	for k, v := range EmerHdrCharToType {
		if k != '&' {
			EmerHdrTypeToChar[v] = k
		}
	}
	EmerHdrTypeToChar[etensor.INT8] = '@'
	EmerHdrTypeToChar[etensor.INT16] = '|'
	EmerHdrTypeToChar[etensor.UINT16] = '|'
	EmerHdrTypeToChar[etensor.INT32] = '|'
	EmerHdrTypeToChar[etensor.UINT32] = '|'
	EmerHdrTypeToChar[etensor.UINT64] = '|'
}

// EmerColType parses the column header for type information using the emergent naming convention
func EmerColType(nm string) (etensor.Type, string) {
	typ, ok := EmerHdrCharToType[nm[0]]
	if ok {
		nm = nm[1:]
	} else {
		typ = etensor.STRING // most general, default
	}
	return typ, nm
}

// ShapeFromString parses string representation of shape as N:d,d,..
func ShapeFromString(dims string) []int {
	clni := strings.Index(dims, ":")
	nd, _ := strconv.Atoi(dims[:clni])
	sh := make([]int, nd)
	ci := clni + 1
	for i := 0; i < nd; i++ {
		dstr := ""
		if i < nd-1 {
			nci := strings.Index(dims[ci:], ",")
			dstr = dims[ci : ci+nci]
			ci += nci + 1
		} else {
			dstr = dims[ci:]
		}
		d, _ := strconv.Atoi(dstr)
		sh[i] = d
	}
	return sh
}

// SchemaFromPlainHeaders configures a Table Schema based on plain headers.
// All columns are of type String and must be converted later to numerical types
// as appropriate.
func SchemaFromPlainHeaders(hdrs []string, rec [][]string) (Schema, error) {
	nc := len(hdrs)
	sc := Schema{}
	nr := len(rec)
	for ci := 0; ci < nc; ci++ {
		hd := hdrs[ci]
		if hd == "" {
			hd = fmt.Sprintf("col_%d", ci)
		}
		dt := etensor.STRING
		nmatch := 0
		for ri := 1; ri < nr; ri++ {
			rv := rec[ri][ci]
			if rv == "" {
				continue
			}
			cdt := InferDataType(rv)
			switch {
			case cdt == etensor.STRING: // definitive
				dt = cdt
				break
			case dt == cdt && (nmatch > 1 || ri == nr-1): // good enough
				break
			case dt == cdt: // gather more info
				nmatch++
			case dt == etensor.STRING: // always upgrade from string default
				nmatch = 0
				dt = cdt
			case dt == etensor.INT64 && cdt == etensor.FLOAT64: // upgrade
				nmatch = 0
				dt = cdt
			}
		}
		sc = append(sc, Column{Name: hd, Type: dt, CellShape: nil})
	}
	return sc, nil
}

// InferDataType returns the inferred data type for the given string
// only deals with float64, int, and string types
func InferDataType(str string) etensor.Type {
	if strings.Contains(str, ".") {
		_, err := strconv.ParseFloat(str, 64)
		if err == nil {
			return etensor.FLOAT64
		}
	}
	_, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		return etensor.INT64
	}
	// try float again just in case..
	_, err = strconv.ParseFloat(str, 64)
	if err == nil {
		return etensor.FLOAT64
	}
	return etensor.STRING
}

//////////////////////////////////////////////////////////////////////////
// WriteCSV

// WriteCSV writes a table to a comma-separated-values (CSV) file (where comma = any delimiter,
//  specified in the delim arg).
// If headers = true then generate C++ emergent-tyle column headers and add _H: to the header line
// and _D: to the data lines.  These headers have full configuration information for the tensor
// columns.  Otherwise, only the data is written.
func (dt *Table) WriteCSV(w io.Writer, delim rune, headers bool) error {
	ncol := 0
	var err error
	if headers {
		ncol, err = dt.WriteCSVHeaders(w, delim)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	cw := csv.NewWriter(w)
	if delim != 0 {
		cw.Comma = delim
	}
	for ri := 0; ri < dt.Rows; ri++ {
		err = dt.WriteCSVRowWriter(cw, ri, headers, ncol)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	cw.Flush()
	return nil
}

// WriteCSVHeaders writes headers to a comma-separated-values (CSV) file (where comma = any delimiter,
//  specified in the delim arg).  Returns number of columns in header
func (dt *Table) WriteCSVHeaders(w io.Writer, delim rune) (int, error) {
	cw := csv.NewWriter(w)
	if delim != 0 {
		cw.Comma = delim
	}
	hdrs := dt.EmerHeaders()
	nc := len(hdrs)
	err := cw.Write(hdrs)
	if err != nil {
		return nc, err
	}
	cw.Flush()
	return nc, nil
}

// WriteCSVRow writes given row to a comma-separated-values (CSV) file
// (where comma = any delimiter, specified in the delim arg)
func (dt *Table) WriteCSVRow(w io.Writer, row int, delim rune, headers bool) error {
	cw := csv.NewWriter(w)
	if delim != 0 {
		cw.Comma = delim
	}
	err := dt.WriteCSVRowWriter(cw, row, headers, 0)
	cw.Flush()
	return err
}

// WriteCSVRowWriter uses csv.Writer to write one row
func (dt *Table) WriteCSVRowWriter(cw *csv.Writer, row int, headers bool, ncol int) error {
	prec := -1
	if ps, ok := dt.MetaData["precision"]; ok {
		prec, _ = strconv.Atoi(ps)
	}
	var rec []string
	if ncol > 0 {
		rec = make([]string, 0, ncol)
	} else {
		rec = make([]string, 0)
	}
	rc := 0
	if headers {
		vl := "_D:"
		if len(rec) <= rc {
			rec = append(rec, vl)
		} else {
			rec[rc] = vl
		}
		rc++
	}
	for i := range dt.Cols {
		tsr := dt.Cols[i]
		nd := tsr.NumDims()
		if nd == 1 {
			vl := ""
			if prec <= 0 || tsr.DataType() == etensor.STRING {
				vl = tsr.StringVal1D(row)
			} else {
				vl = strconv.FormatFloat(tsr.FloatVal1D(row), 'g', prec, 64)
			}
			if len(rec) <= rc {
				rec = append(rec, vl)
			} else {
				rec[rc] = vl
			}
			rc++
		} else {
			csh := etensor.NewShape(tsr.Shapes()[1:], nil, nil) // cell shape
			tc := csh.Len()
			for ti := 0; ti < tc; ti++ {
				vl := ""
				if prec <= 0 || tsr.DataType() == etensor.STRING {
					vl = tsr.StringVal1D(row*tc + ti)
				} else {
					vl = strconv.FormatFloat(tsr.FloatVal1D(row*tc+ti), 'g', prec, 64)
				}
				if len(rec) <= rc {
					rec = append(rec, vl)
				} else {
					rec[rc] = vl
				}
				rc++
			}
		}
	}
	err := cw.Write(rec)
	return err
}

// EmerHeaders generates emergent DataTable header strings from the table.
// These have full information about type and tensor cell dimensionality.
// Also includes the _H: header marker typically output to indicate a header row as first element.
func (dt *Table) EmerHeaders() []string {
	hdrs := []string{"_H:"}
	for i := range dt.Cols {
		tsr := dt.Cols[i]
		nm := dt.ColNames[i]
		nm = string([]byte{EmerHdrTypeToChar[tsr.DataType()]}) + nm
		if tsr.NumDims() == 1 {
			hdrs = append(hdrs, nm)
		} else {
			csh := etensor.NewShape(tsr.Shapes()[1:], nil, nil) // cell shape
			tc := csh.Len()
			nd := csh.NumDims()
			fnm := nm + fmt.Sprintf("[%v:", nd)
			dn := fmt.Sprintf("<%v:", nd)
			ffnm := fnm
			for di := 0; di < nd; di++ {
				ffnm += "0"
				dn += fmt.Sprintf("%v", csh.Dim(di))
				if di < nd-1 {
					ffnm += ","
					dn += ","
				}
			}
			ffnm += "]" + dn + ">"
			hdrs = append(hdrs, ffnm)
			for ti := 1; ti < tc; ti++ {
				idx := csh.Index(ti)
				ffnm := fnm
				for di := 0; di < nd; di++ {
					ffnm += fmt.Sprintf("%v", idx[di])
					if di < nd-1 {
						ffnm += ","
					}
				}
				ffnm += "]"
				hdrs = append(hdrs, ffnm)
			}
		}
	}
	return hdrs
}
