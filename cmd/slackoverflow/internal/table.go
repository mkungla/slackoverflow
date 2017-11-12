// Copyright © 2016 -2017 A-Frame authors.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package internal

import (
	"fmt"
	"strings"

	"github.com/logrusorgru/aurora"
)

// Table struct when generating table view in console
type Table struct {
	colCount int
	rowCount int
	cols     []TableCol
	rows     [][]TableCol
}

// TableCol is struct for column in table
type TableCol struct {
	data      string
	colLenght int
}

// NewTable creates new table
func NewTable(h ...string) *Table {
	t := new(Table)
	t.colCount = len(h)
	for _, heading := range h {
		col := TableCol{heading, len(heading)}
		t.cols = append(t.cols, col)
	}
	return t
}

// AddRow to the table
func (t *Table) AddRow(a ...interface{}) {
	if t.colCount == len(a) {
		row := []TableCol{}
		for i, rowColRaw := range a {
			rowCol := fmt.Sprintf("%v", rowColRaw)
			colLen := len(rowCol)
			col := TableCol{rowCol, colLen}
			if colLen > t.cols[i].colLenght {
				t.cols[i].colLenght = colLen
			}
			row = append(row, col)
		}
		t.rows = append(t.rows, row)
	} else {
		fmt.Printf("This table accepts exactly %d arguments, but %d provided\n", t.colCount, len(a))
	}
}

// Print the table
func (t *Table) Print() {
	t.border()
	row := []string{""}
	for _, col := range t.cols {
		headingLen := len(col.data)
		raw := aurora.Bold(col.data).String()
		if col.colLenght > headingLen {
			padRight := col.colLenght - headingLen
			raw = strings.Join([]string{raw, strings.Repeat(" ", padRight)}, "")
		}
		row = append(row, raw)
	}
	fmt.Printf("%s\n", strings.Join(row, aurora.Magenta(" | ").String()))
	t.border()
	for _, rowData := range t.rows {
		row := []string{""}
		for i, col := range rowData {
			raw := col.data
			if t.cols[i].colLenght > col.colLenght {
				padRight := t.cols[i].colLenght - col.colLenght
				raw = strings.Join([]string{raw, strings.Repeat(" ", padRight)}, "")
			}
			row = append(row, raw)
		}

		fmt.Printf("%s\n",
			strings.Join(row, aurora.Magenta(" | ").String()),
		)
	}
	t.border()
}

// Border for table
func (t *Table) border() {
	fmt.Println(aurora.Magenta(strings.Repeat("-", 120)))
}
