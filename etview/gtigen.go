// Code generated by "core generate -add-types"; DO NOT EDIT.

package etview

import (
	"sync"

	"cogentcore.org/core/colors/colormap"
	"cogentcore.org/core/giv"
	"cogentcore.org/core/gti"
	"cogentcore.org/core/ki"
	"cogentcore.org/core/mat32"
	"github.com/emer/etable/v2/etensor"
)

// SimMatGridType is the [gti.Type] for [SimMatGrid]
var SimMatGridType = gti.AddType(&gti.Type{Name: "github.com/emer/etable/v2/etview.SimMatGrid", IDName: "sim-mat-grid", Doc: "SimMatGrid is a widget that displays a similarity / distance matrix\nwith tensor values as a grid of colored squares, and labels for rows, cols", Directives: []gti.Directive{{Tool: "gti", Directive: "add"}}, Embeds: []gti.Field{{Name: "TensorGrid"}}, Fields: []gti.Field{{Name: "SimMat", Doc: "the similarity / distance matrix"}, {Name: "rowMaxSz"}, {Name: "rowMinBlank"}, {Name: "rowNGps"}, {Name: "colMaxSz"}, {Name: "colMinBlank"}, {Name: "colNGps"}}, Instance: &SimMatGrid{}})

// NewSimMatGrid adds a new [SimMatGrid] with the given name to the given parent:
// SimMatGrid is a widget that displays a similarity / distance matrix
// with tensor values as a grid of colored squares, and labels for rows, cols
func NewSimMatGrid(par ki.Ki, name ...string) *SimMatGrid {
	return par.NewChild(SimMatGridType, name...).(*SimMatGrid)
}

// KiType returns the [*gti.Type] of [SimMatGrid]
func (t *SimMatGrid) KiType() *gti.Type { return SimMatGridType }

// New returns a new [*SimMatGrid] value
func (t *SimMatGrid) New() ki.Ki { return &SimMatGrid{} }

// SetRowMaxSz sets the [SimMatGrid.rowMaxSz]
func (t *SimMatGrid) SetRowMaxSz(v mat32.Vec2) *SimMatGrid { t.rowMaxSz = v; return t }

// SetRowMinBlank sets the [SimMatGrid.rowMinBlank]
func (t *SimMatGrid) SetRowMinBlank(v int) *SimMatGrid { t.rowMinBlank = v; return t }

// SetRowNgps sets the [SimMatGrid.rowNGps]
func (t *SimMatGrid) SetRowNgps(v int) *SimMatGrid { t.rowNGps = v; return t }

// SetColMaxSz sets the [SimMatGrid.colMaxSz]
func (t *SimMatGrid) SetColMaxSz(v mat32.Vec2) *SimMatGrid { t.colMaxSz = v; return t }

// SetColMinBlank sets the [SimMatGrid.colMinBlank]
func (t *SimMatGrid) SetColMinBlank(v int) *SimMatGrid { t.colMinBlank = v; return t }

// SetColNgps sets the [SimMatGrid.colNGps]
func (t *SimMatGrid) SetColNgps(v int) *SimMatGrid { t.colNGps = v; return t }

// SetTooltip sets the [SimMatGrid.Tooltip]
func (t *SimMatGrid) SetTooltip(v string) *SimMatGrid { t.Tooltip = v; return t }

// SetDisp sets the [SimMatGrid.Disp]
func (t *SimMatGrid) SetDisp(v TensorDisp) *SimMatGrid { t.Disp = v; return t }

// SetColorMap sets the [SimMatGrid.ColorMap]
func (t *SimMatGrid) SetColorMap(v *colormap.Map) *SimMatGrid { t.ColorMap = v; return t }

// TableViewType is the [gti.Type] for [TableView]
var TableViewType = gti.AddType(&gti.Type{Name: "github.com/emer/etable/v2/etview.TableView", IDName: "table-view", Doc: "etview.TableView provides a GUI interface for etable.Table's", Embeds: []gti.Field{{Name: "SliceViewBase"}}, Fields: []gti.Field{{Name: "Table", Doc: "the idx view of the table that we're a view of"}, {Name: "TsrDisp", Doc: "overall display options for tensor display"}, {Name: "ColTsrDisp", Doc: "per column tensor display params"}, {Name: "ColTsrBlank", Doc: "per column blank tensor values"}, {Name: "NCols", Doc: "number of columns in table (as of last update)"}, {Name: "SortIdx", Doc: "current sort index"}, {Name: "SortDesc", Doc: "whether current sort order is descending"}, {Name: "HeaderWidths", Doc: "HeaderWidths has number of characters in each header, per visfields"}, {Name: "ColMaxWidths", Doc: "ColMaxWidths records maximum width in chars of string type fields"}, {Name: "BlankString", Doc: "\tblank values for out-of-range rows"}, {Name: "BlankFloat"}}, Instance: &TableView{}})

// NewTableView adds a new [TableView] with the given name to the given parent:
// etview.TableView provides a GUI interface for etable.Table's
func NewTableView(par ki.Ki, name ...string) *TableView {
	return par.NewChild(TableViewType, name...).(*TableView)
}

// KiType returns the [*gti.Type] of [TableView]
func (t *TableView) KiType() *gti.Type { return TableViewType }

// New returns a new [*TableView] value
func (t *TableView) New() ki.Ki { return &TableView{} }

// SetTsrDisp sets the [TableView.TsrDisp]:
// overall display options for tensor display
func (t *TableView) SetTsrDisp(v TensorDisp) *TableView { t.TsrDisp = v; return t }

// SetColTsrDisp sets the [TableView.ColTsrDisp]:
// per column tensor display params
func (t *TableView) SetColTsrDisp(v map[int]*TensorDisp) *TableView { t.ColTsrDisp = v; return t }

// SetColTsrBlank sets the [TableView.ColTsrBlank]:
// per column blank tensor values
func (t *TableView) SetColTsrBlank(v map[int]*etensor.Float64) *TableView {
	t.ColTsrBlank = v
	return t
}

// SetNcols sets the [TableView.NCols]:
// number of columns in table (as of last update)
func (t *TableView) SetNcols(v int) *TableView { t.NCols = v; return t }

// SetSortIdx sets the [TableView.SortIdx]:
// current sort index
func (t *TableView) SetSortIdx(v int) *TableView { t.SortIdx = v; return t }

// SetSortDesc sets the [TableView.SortDesc]:
// whether current sort order is descending
func (t *TableView) SetSortDesc(v bool) *TableView { t.SortDesc = v; return t }

// SetHeaderWidths sets the [TableView.HeaderWidths]:
// HeaderWidths has number of characters in each header, per visfields
func (t *TableView) SetHeaderWidths(v ...int) *TableView { t.HeaderWidths = v; return t }

// SetBlankString sets the [TableView.BlankString]:
//
//	blank values for out-of-range rows
func (t *TableView) SetBlankString(v string) *TableView { t.BlankString = v; return t }

// SetBlankFloat sets the [TableView.BlankFloat]
func (t *TableView) SetBlankFloat(v float64) *TableView { t.BlankFloat = v; return t }

// SetTooltip sets the [TableView.Tooltip]
func (t *TableView) SetTooltip(v string) *TableView { t.Tooltip = v; return t }

// SetStackTop sets the [TableView.StackTop]
func (t *TableView) SetStackTop(v int) *TableView { t.StackTop = v; return t }

// SetMinRows sets the [TableView.MinRows]
func (t *TableView) SetMinRows(v int) *TableView { t.MinRows = v; return t }

// SetViewPath sets the [TableView.ViewPath]
func (t *TableView) SetViewPath(v string) *TableView { t.ViewPath = v; return t }

// SetViewMu sets the [TableView.ViewMu]
func (t *TableView) SetViewMu(v *sync.Mutex) *TableView { t.ViewMu = v; return t }

// SetSelVal sets the [TableView.SelVal]
func (t *TableView) SetSelVal(v any) *TableView { t.SelVal = v; return t }

// SetSelIdx sets the [TableView.SelIdx]
func (t *TableView) SetSelIdx(v int) *TableView { t.SelIdx = v; return t }

// SetInitSelIdx sets the [TableView.InitSelIdx]
func (t *TableView) SetInitSelIdx(v int) *TableView { t.InitSelIdx = v; return t }

// SetTmpSave sets the [TableView.TmpSave]
func (t *TableView) SetTmpSave(v giv.Value) *TableView { t.TmpSave = v; return t }

var _ = gti.AddType(&gti.Type{Name: "github.com/emer/etable/v2/etview.TensorLayout", IDName: "tensor-layout", Doc: "TensorLayout are layout options for displaying tensors", Directives: []gti.Directive{{Tool: "gti", Directive: "add"}}, Fields: []gti.Field{{Name: "OddRow", Doc: "even-numbered dimensions are displayed as Y*X rectangles -- this determines along which dimension to display any remaining odd dimension: OddRow = true = organize vertically along row dimension, false = organize horizontally across column dimension"}, {Name: "TopZero", Doc: "if true, then the Y=0 coordinate is displayed from the top-down; otherwise the Y=0 coordinate is displayed from the bottom up, which is typical for emergent network patterns."}, {Name: "Image", Doc: "display the data as a bitmap image.  if a 2D tensor, then it will be a greyscale image.  if a 3D tensor with size of either the first or last dim = either 3 or 4, then it is a RGB(A) color image"}}})

var _ = gti.AddType(&gti.Type{Name: "github.com/emer/etable/v2/etview.TensorDisp", IDName: "tensor-disp", Doc: "TensorDisp are options for displaying tensors", Directives: []gti.Directive{{Tool: "gti", Directive: "add"}}, Embeds: []gti.Field{{Name: "TensorLayout"}}, Fields: []gti.Field{{Name: "Range", Doc: "range to plot"}, {Name: "MinMax", Doc: "if not using fixed range, this is the actual range of data"}, {Name: "ColorMap", Doc: "the name of the color map to use in translating values to colors"}, {Name: "GridFill", Doc: "what proportion of grid square should be filled by color block -- 1 = all, .5 = half, etc"}, {Name: "DimExtra", Doc: "amount of extra space to add at dimension boundaries, as a proportion of total grid size"}, {Name: "GridMinSize", Doc: "minimum size for grid squares -- they will never be smaller than this"}, {Name: "GridMaxSize", Doc: "maximum size for grid squares -- they will never be larger than this"}, {Name: "TotPrefSize", Doc: "total preferred display size along largest dimension.\ngrid squares will be sized to fit within this size,\nsubject to harder GridMin / Max size constraints"}, {Name: "FontSize", Doc: "font size in standard point units for labels (e.g., SimMat)"}, {Name: "GridView", Doc: "our gridview, for update method"}}})

// TensorGridType is the [gti.Type] for [TensorGrid]
var TensorGridType = gti.AddType(&gti.Type{Name: "github.com/emer/etable/v2/etview.TensorGrid", IDName: "tensor-grid", Doc: "TensorGrid is a widget that displays tensor values as a grid of colored squares.", Methods: []gti.Method{{Name: "EditSettings", Directives: []gti.Directive{{Tool: "gti", Directive: "add"}}}}, Embeds: []gti.Field{{Name: "WidgetBase"}}, Fields: []gti.Field{{Name: "Tensor", Doc: "the tensor that we view"}, {Name: "Disp", Doc: "display options"}, {Name: "ColorMap", Doc: "the actual colormap"}}, Instance: &TensorGrid{}})

// NewTensorGrid adds a new [TensorGrid] with the given name to the given parent:
// TensorGrid is a widget that displays tensor values as a grid of colored squares.
func NewTensorGrid(par ki.Ki, name ...string) *TensorGrid {
	return par.NewChild(TensorGridType, name...).(*TensorGrid)
}

// KiType returns the [*gti.Type] of [TensorGrid]
func (t *TensorGrid) KiType() *gti.Type { return TensorGridType }

// New returns a new [*TensorGrid] value
func (t *TensorGrid) New() ki.Ki { return &TensorGrid{} }

// SetDisp sets the [TensorGrid.Disp]:
// display options
func (t *TensorGrid) SetDisp(v TensorDisp) *TensorGrid { t.Disp = v; return t }

// SetColorMap sets the [TensorGrid.ColorMap]:
// the actual colormap
func (t *TensorGrid) SetColorMap(v *colormap.Map) *TensorGrid { t.ColorMap = v; return t }

// SetTooltip sets the [TensorGrid.Tooltip]
func (t *TensorGrid) SetTooltip(v string) *TensorGrid { t.Tooltip = v; return t }

// TensorViewType is the [gti.Type] for [TensorView]
var TensorViewType = gti.AddType(&gti.Type{Name: "github.com/emer/etable/v2/etview.TensorView", IDName: "tensor-view", Doc: "etview.TensorView provides a GUI interface for etable.Tensor's\nusing a tabular rows-and-columns interface using textfields for editing.\nThis provides an editable complement to the TensorGrid graphical display.", Embeds: []gti.Field{{Name: "WidgetBase"}}, Instance: &TensorView{}})

// NewTensorView adds a new [TensorView] with the given name to the given parent:
// etview.TensorView provides a GUI interface for etable.Tensor's
// using a tabular rows-and-columns interface using textfields for editing.
// This provides an editable complement to the TensorGrid graphical display.
func NewTensorView(par ki.Ki, name ...string) *TensorView {
	return par.NewChild(TensorViewType, name...).(*TensorView)
}

// KiType returns the [*gti.Type] of [TensorView]
func (t *TensorView) KiType() *gti.Type { return TensorViewType }

// New returns a new [*TensorView] value
func (t *TensorView) New() ki.Ki { return &TensorView{} }

// SetTooltip sets the [TensorView.Tooltip]
func (t *TensorView) SetTooltip(v string) *TensorView { t.Tooltip = v; return t }

var _ = gti.AddType(&gti.Type{Name: "github.com/emer/etable/v2/etview.TensorGridValue", IDName: "tensor-grid-value", Doc: "TensorGridValue manages a TensorGrid view of an etensor.Tensor", Embeds: []gti.Field{{Name: "ValueBase"}}})

var _ = gti.AddType(&gti.Type{Name: "github.com/emer/etable/v2/etview.TensorValue", IDName: "tensor-value", Doc: "TensorValue presents a button that pulls up the TensorView viewer for an etensor.Tensor", Embeds: []gti.Field{{Name: "ValueBase"}}})

var _ = gti.AddType(&gti.Type{Name: "github.com/emer/etable/v2/etview.TableValue", IDName: "table-value", Doc: "TableValue presents a button that pulls up the TableView viewer for an etable.Table", Embeds: []gti.Field{{Name: "ValueBase"}}})

var _ = gti.AddType(&gti.Type{Name: "github.com/emer/etable/v2/etview.SimMatValue", IDName: "sim-mat-value", Doc: "SimMatValue presents a button that pulls up the SimMatGridView viewer for an etable.Table", Embeds: []gti.Field{{Name: "ValueBase"}}})
