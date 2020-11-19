# pyet: Python etable functions

The `pyet` library code here, which is installed by default in emergent python packages such as `pyleabra` and `etorch`, provides functions to convert between `etable` data structures (including `etensor`) and standard Python data structures such as `numpy` `ndarray`, `pandas`, and `pytorch` `TensorDataset`.

As of the current implementation, the core conversion involves copying data between etensor and numpy ndarray's, which are then used in various aggregations.  The copy keeps everything simple in terms of each side owning its own data structures, at the cost of doing an element-wise for loop copy that can be slow in Python.  In the future apache arrow could be used for direct memory sharing, but this introduces additional extra complexity in terms of explicit reference counting logic, and the potential for C-like memory management bugs that users of Python and Go are less likely to be familiar with. Typically, Python users are used to absorbing significant performance costs for simplicity benefits.

# etensor <-> numpy.ndarray

There are two methods for creating a new object in either direction:

* `pyet.etensor_to_numpy(etensor.Tensor) -> numpy.ndarray` -- takes a tensor and returns the equivalently shaped ndarray with the tensor data.  Loses any dimensions names as numpy apparently does not support those.   Also, assumes default row-major layout which is common between both frameworks.

* `pyet.numpy_to_etensor(numpy.ndarray) -> etensor.Tensor` -- takes an ndarray and returns equivalently shaped Tensor with the ndarray data.

And two similar methods for copying between existing objects (dest first arg, src next):

* `pyet.copy_etensor_to_numpy(numpy.ndarray, etensor.Tensor)` -- copy from etensor to numpy

* `pyet.copy_numpy_to_etensor(etensor.Tensor, numpy.ndarray)` -- copy from numpy to etensor

# etable -> Python

Because the various Python DataFrames don't quite capture the same columns-of-tensors structure of the etable.Table, we have a `pyet.PyEtable` class that just holds a converted Table as a list of numpy.ndarray columns, with a dictionary for accessing by name.  

Thus, the procedure is to first convert an `etable.Table` to `pyet.PyEtable` using `pyet.etable_to_py`, and then from there you can do further conversions.

* `pyet.etable_to_py(etable.Table) -> pyet.PyEtable` returns converted `pyet.PyEtable`, which can be used directly by accessing the `numpy.ndarray` columns of data, or converted further.

* `pyet.py_to_etable(PyEtable) -> etable.Table` returns PyEtable converted to an etable.Table.

* `pyet.copy_etable_to_py(PyEtable, etable.Table)` copies etable.Table values to same-named PyEtable columns.

* `pyet.copy_py_to_etable(etable.Table, PyEtable)` copies PyEtable values to same-named etable.Table columns.

* `pyet.etable_to_torch(PyEtable)` returns a pytorch `TensorDataset`, which has the same structure as an `etable`, and is used in the `pytorch` neural network framework, except that the `TensorDataset` does NOT support string columns as labels, so those are skipped.

* `pyet.etable_to_pandas(PyEtable) -> pandas.DataFrame` returns a `pandas.DataFrame` with data from the table -- if there are tensor (multidimensional) columns, they are splayed out across sequential 1D columns, numbered with _idx subscripts.  Optional skip_tensors arg instead just skips over tensors.

* `pyet.pandas_to_etable(pandas.DataFrame) -> pyet.PyEtable` returns a PyEtable from pandas dataframe.  By definition, all columns will be 1D.  See the Pandas test case in `etest.py` for use of `MergeCols` and `ReshapeCol` to turn 1D cols back into multidimensional tensor columns.

# Arrow

Here's some notes for potential future arrow implementations:

* In Go, Tensor returns convenient `array.Data` struct that manages a slice of `memory.Buffer` elements.  But in Python, it seems like you have to deal directly with the Buffers.  This would entail various wrappers for converting the lists, etc.

* The arrow.Tensor returned by etensor.Tensor would be a new temp object -- refcounting logic etc not directly in original etensor, so it is unclear exactly how this would work..


