# Copyright (c) 2020, The Emergent Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# code for converting etensor and etable to / from various python data formats
# including numpy, pandas, and pytorch `TensorDataset`,
# which has the same structure as an `etable`, and is used in the
# `pytorch` neural network framework.

from leabra import go, etable, etensor

import numpy as np
import pandas as pd
import torch
import torch.utils.data as data_utils

def etensor_to_numpy(et):
    """
    returns a numpy ndarray constructed from the given etensor.Tensor.
    data is copied into the numpy ndarray -- it is not a view.
    unfortunately it is not especially fast, using element-wise copy.
    """
    nar = 0
    if et.DataType() == etensor.UINT8:
        nar = np.array(etensor.Uint8(et).Values, dtype=np.uint8)
    elif et.DataType() == etensor.INT8:
        nar = np.array(etensor.Int8(et).Values, dtype=np.int8)
    elif et.DataType() == etensor.UINT16:
        nar = np.array(etensor.Uint16(et).Values, dtype=np.uint16)
    elif et.DataType() == etensor.INT16:
        nar = np.array(etensor.Int16(et).Values, dtype=np.int16)
    elif et.DataType() == etensor.UINT32:
        nar = np.array(etensor.Uint32(et).Values, dtype=np.uint32)
    elif et.DataType() == etensor.INT32:
        nar = np.array(etensor.Int32(et).Values, dtype=np.int32)
    elif et.DataType() == etensor.UINT64:
        nar = np.array(etensor.Uint64(et).Values, dtype=np.uint64)
    elif et.DataType() == etensor.INT64:
        nar = np.array(etensor.Int64(et).Values, dtype=np.int64)
    elif et.DataType() == etensor.FLOAT32:
        nar = np.array(etensor.Float32(et).Values, dtype=np.float32)
    elif et.DataType() == etensor.FLOAT64:
        nar = np.array(etensor.Float64(et).Values, dtype=np.float64)
    elif et.DataType() == etensor.STRING:
        nar = np.array(etensor.String(et).Values)
    elif et.DataType() == etensor.INT:
        nar = np.array(etensor.Int(et).Values, dtype=np.intc)
    elif et.DataType() == etensor.BOOL:
        etb = etensor.Bits(et)
        sz = etb.Len()
        nar = np.zeros(sz, dtype=np.bool_)
        for i in range(sz):
            nar[i] = etb.Value1D(i)
    else:
        raise TypeError("tensor with type %s cannot be converted" % (et.DataType().String()))
        return 0
    # there does not appear to be a way to set the shape at the same time as initializing 
    return nar.reshape(et.Shapes())


def numpy_to_etensor(nar):
    """
    returns an etensor.Tensor constructed from the given etensor.Tensor
    data is copied into the Tensor -- it is not a view.
    unfortunately it is not especially fast, using element-wise copy.
    """
    et = 0
    if nar.dtype == np.uint8:
        et = etensor.NewUint8(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(np.ravel(nar))
    elif nar.dtype == np.int8:
        et = etensor.NewInt8(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(np.ravel(nar))
    elif nar.dtype == np.uint16:
        et = etensor.NewUint16(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(np.ravel(nar))
    elif nar.dtype == np.int16:
        et = etensor.NewInt16(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(np.ravel(nar))
    elif nar.dtype == np.uint32:
        et = etensor.NewUint32(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(np.ravel(nar))
    elif nar.dtype == np.int32:
        et = etensor.NewInt32(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(np.ravel(nar))
    elif nar.dtype == np.uint64:
        et = etensor.NewUint64(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(np.ravel(nar))
    elif nar.dtype == np.int64:
        et = etensor.NewInt64(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(np.ravel(nar))
    elif nar.dtype == np.float32:
        et = etensor.NewFloat32(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(np.ravel(nar))
    elif nar.dtype == np.float64:
        et = etensor.NewFloat64(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(np.ravel(nar))
    elif nar.dtype.type is np.string_ or nar.dtype.type is np.str_:
        et = etensor.NewString(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(np.ravel(nar))
    elif nar.dtype == np.int_ or nar.dtype == np.intc:
        et = etensor.NewInt(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(np.ravel(nar))
    elif nar.dtype == np.bool_:
        et = etensor.NewBits(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        rnar = np.ravel(nar)
        sz = len(rnar)
        for i in range(sz):
            et.Set1D(i, rnar[i])
    else:
        raise TypeError("numpy ndarray with type %s cannot be converted" % (nar.dtype))
        return 0
    return et

##########################################
# Tables
    
class PyEtable(object):
    """
    PyEtable is a Python version of the Go etable.Table, with slices of columns 
    as numpy ndarrays, and corresponding column names, along with a coordinated
    dictionary of names to col indexes.  This is returned by basic
    etable_to_py() function to convert all data from an etable,
    and can then be used to convert into other python datatable / frame 
    structures.
    """
    def __init__(self):
        self.Cols = []
        self.ColNames = []
        self.Rows = 0
        self.ColNameMap = {}
        self.MetaData = {}

    def __str__(dt):
        return "Columns: %s\nRows: %d Cols:\n%s\n" % (dt.ColNames, dt.Rows, dt.Cols)
        
    def UpdateColNameMap(dt):
        """
        UpdateColNameMap updates the column name map
        """
        dt.ColNameMap = {}
        for i, nm in enumerate(dt.ColNames):
            dt.ColNameMap[nm] = i

    def AddCol(dt, nar, name):
        """
        AddCol adds a numpy ndarray as a new column, with given name
        """
        dt.Cols.append(nar)
        dt.ColNames.append(name)
        dt.UpdateColNameMap()
        
    def ColByName(dt, name):
        """
        ColByName returns column of given name, or raises a LookupError if not found
        """
        if name in dt.ColNameMap:
            return dt.Cols[dt.ColNameMap[name]]
        raise LookupError("column named: %s not found" % (name))

def etable_to_py(et):
    """
    returns a PyEtable python version of given etable.Table.
    The PyEtable can then be converted into other standard Python formats,
    but most of them don't quite capture exactly the same information, so
    the PyEtable can be handy to keep around.
    """
    pt = PyEtable()
    pt.Rows = et.Rows
    nc = len(et.Cols)
    for ci in range(nc):
        dc = et.Cols[ci]
        cn = et.ColNames[ci]
        nar = etensor_to_numpy(dc)
        pt.AddCol(nar, cn)
        for k in et.MetaData:
            pt.MetaData[k] = et.MetaData[k]
    return pt
        
def etable_to_torch(et):
    """
    returns a torch.utils.data.TensorDataset constructed from the numeric columns
    of the given PyEtable (string columns are not allowed in TensorDataset)
    """
    tsrs = []
    nc = len(et.Cols)
    for ci in range(nc):
        dc = et.Cols[ci]
        cn = et.ColNames[ci]

        if dc.dtype.type is np.string_ or dc.dtype.type is np.str_:
            continue
        
        tsr = torch.from_numpy(dc)
        tsrs.append(tsr)
    ds = data_utils.TensorDataset(*tsrs)
    return ds

def etable_to_pandas(et):
    """
    returns a pandas DataFrame constructed from the columns
    of the given PyEtable, spreading tensor cells over sequential
    1d columns.
    """
    pass
        
