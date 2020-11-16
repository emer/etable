# pyet: Python etable functions

The `pyet` library code here, which is installed by default in emergent python packages such as pyleabra, provides functions to convert between `etable` data structures (including `etensor`) and standard Python data structures such as `numpy` `ndarray`, `pandas`, and `pytorch` `TensorDataset`.

As of the current implementation, the core conversion involves copying data between etensor and numpy ndarray's, which are then used in various aggregations.  The copy keeps everything simple in terms of each side owning its own data structures, at the cost of doing an element-wise for loop copy that can be slow in Python.  In the future apache arrow could be used for direct memory sharing, but this introduces additional extra complexity in terms of explicit reference counting logic, and the potential for C-like memory management bugs that users of Python and Go are less likely to be familiar with. Typically, Python users are used to absorbing significant performance costs for simplicity benefits.

# etensor <-> numpy.ndarray

There are just two methods, one for each direction:

* `etensor_to_numpy(etensor.Tensor) numpy.ndarray` -- takes a tensor and returns the equivalently shaped ndarray with the tensor data.  Loses any dimensions names as numpy apparently does not support those.   Also, assumes default row-major layout which is common between both frameworks.

* `numpy_to_etensor(numpy.ndarray) etensor.Tensor` -- takes an ndarray and returns equivalently shaped Tensor with the ndarray data.

# etable -> Python

Because the various Python DataFrame style 

# etable -> torch TensorDataset

This example provides code for converting `etable` to / from a torch `TensorDataset`, which has the same structure as an `etable`, and is used in the `pytorch` neural network framework.

Except that the `TensorDataset` does NOT support string columns as labels.  That is unfortunate.


# Arrow

Here's some notes for potential future arrow implementations:

* In Go, Tensor returns convenient `array.Data` struct that manages a slice of `memory.Buffer` elements.  But in Python, it seems like you have to deal directly with the Buffers.  This would entail various wrappers for converting the lists, etc.

* The arrow.Tensor returned by etensor.Tensor would be a new temp object -- refcounting logic etc not directly in original etensor, so it is unclear exactly how this would work..
