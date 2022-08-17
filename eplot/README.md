# eplot

Docs: [GoDoc](https://pkg.go.dev/github.com/emer/etable/eplot)

eplot provides an interactive, graphical plotting utility for etable data, as a GoGi Widget, with multiple Y axis values and options for XY vs. Bar plots.

To use, create a `Plot2D` widget in a GoGi scenegraph, e.g.:

```Go
		plt := gui.TabView.AddNewTab(eplot.KiT_Plot2D, mode+" "+time+" Plot").(*eplot.Plot2D)
		plt.SetTable(lt.Table)
		plt.Params.FmMetaMap(lt.Meta)
```


