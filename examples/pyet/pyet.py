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
    """
    et = 0
    narf = np.reshape(nar, -1) # flat view
    if nar.dtype == np.uint8:
        et = etensor.NewUint8(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(narf)
    elif nar.dtype == np.int8:
        et = etensor.NewInt8(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(narf)
    elif nar.dtype == np.uint16:
        et = etensor.NewUint16(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(narf)
    elif nar.dtype == np.int16:
        et = etensor.NewInt16(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(narf)
    elif nar.dtype == np.uint32:
        et = etensor.NewUint32(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(narf)
    elif nar.dtype == np.int32:
        et = etensor.NewInt32(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(narf)
    elif nar.dtype == np.uint64:
        et = etensor.NewUint64(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(narf)
    elif nar.dtype == np.int64:
        et = etensor.NewInt64(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(narf)
    elif nar.dtype == np.float32:
        et = etensor.NewFloat32(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(narf)
    elif nar.dtype == np.float64:
        et = etensor.NewFloat64(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(narf)
    elif nar.dtype.type is np.string_ or nar.dtype.type is np.str_:
        et = etensor.NewString(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(narf)
    elif nar.dtype == np.int_ or nar.dtype == np.intc:
        et = etensor.NewInt(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        et.Values.copy(narf)
    elif nar.dtype == np.bool_:
        et = etensor.NewBits(go.Slice_int(list(nar.shape)), go.nil, go.nil)
        rnar = narf
        sz = len(rnar)
        for i in range(sz):
            et.Set1D(i, rnar[i])
    else:
        raise TypeError("numpy ndarray with type %s cannot be converted" % (nar.dtype))
        return 0
    return et

#########################
#  Copying
    
def copy_etensor_to_numpy(nar, et):
    """
    copies data from etensor.Tensor (et, source) to existing numpy ndarray (nar, dest).
    """
    narf = np.reshape(nar, -1)
    etv = et
    if et.DataType() == etensor.UINT8:
        etv = etensor.Uint8(et).Values
    elif et.DataType() == etensor.INT8:
        etv = etensor.Int8(et).Values
    elif et.DataType() == etensor.UINT16:
        etv = etensor.Uint16(et).Values
    elif et.DataType() == etensor.INT16:
        etv = etensor.Int16(et).Values
    elif et.DataType() == etensor.UINT32:
        etv = etensor.Uint32(et).Values
    elif et.DataType() == etensor.INT32:
        etv = etensor.Int32(et).Values
    elif et.DataType() == etensor.UINT64:
        etv = etensor.Uint64(et).Values
    elif et.DataType() == etensor.INT64:
        etv = etensor.Int64(et).Values
    elif et.DataType() == etensor.FLOAT32:
        etv = etensor.Float32(et).Values
    elif et.DataType() == etensor.FLOAT64:
        etv = etensor.Float64(et).Values
    elif et.DataType() == etensor.STRING:
        etv = etensor.String(et).Values
    elif et.DataType() == etensor.INT:
        etv = etensor.Int(et).Values
    elif et.DataType() == etensor.BOOL:
        etb = etensor.Bits(et)
        sz = min(etb.Len(), len(narf))
        for i in range(sz):
            narf[i] = etb.Value1D(i)
        return
    else:
        raise TypeError("tensor with type %s cannot be copied" % (et.DataType().String()))
        return 0
    np.copyto(narf, etv, casting='unsafe')

def copy_numpy_to_etensor(et, nar):
    """
    copies data from numpy ndarray (nar, source) to existing etensor.Tensor (et, dest) 
    """
    narf = np.reshape(nar, -1)
    etv = et
    if et.DataType() == etensor.UINT8:
        etv = etensor.Uint8(et).Values
    elif et.DataType() == etensor.INT8:
        etv = etensor.Int8(et).Values
    elif et.DataType() == etensor.UINT16:
        etv = etensor.Uint16(et).Values
    elif et.DataType() == etensor.INT16:
        etv = etensor.Int16(et).Values
    elif et.DataType() == etensor.UINT32:
        etv = etensor.Uint32(et).Values
    elif et.DataType() == etensor.INT32:
        etv = etensor.Int32(et).Values
    elif et.DataType() == etensor.UINT64:
        etv = etensor.Uint64(et).Values
    elif et.DataType() == etensor.INT64:
        etv = etensor.Int64(et).Values
    elif et.DataType() == etensor.FLOAT32:
        etv = etensor.Float32(et).Values
    elif et.DataType() == etensor.FLOAT64:
        etv = etensor.Float64(et).Values
    elif et.DataType() == etensor.STRING:
        etv = etensor.String(et).Values
    elif et.DataType() == etensor.INT:
        etv = etensor.Int(et).Values
    elif et.DataType() == etensor.BOOL:
        etb = etensor.Bits(et)
        sz = min(etb.Len(), len(narf))
        for i in range(sz):
            narf[i] = etb.Value1D(i)
        return
    else:
        raise TypeError("tensor with type %s cannot be copied" % (et.DataType().String()))
        return 0
    etv.copy(narf)  # go slice copy, not python copy = clone


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
        return "Columns: %s\nRows: %d Cols:\n%s\n" % (dt.ColNameMap, dt.Rows, dt.Cols)
        
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
        
    def MergeCols(dt, st_nm, n):
        """
        MergeCols merges n sequential columns into a multidimensional array, starting at given column name
        Resulting columns are all stored at st_nm
        """
        sti = dt.ColNameMap[st_nm]
        cls = dt.Cols[sti:sti+n]
        nc = np.column_stack(cls)
        dt.Cols[sti] = nc
        del dt.Cols[sti+1:sti+n]
        del dt.ColNames[sti+1:sti+n]
        dt.UpdateColNameMap()
        
    def ReshapeCol(dt, colnm, shp):
        """
        ReshapeCol reshapes column to given shape
        """
        ci = dt.ColNameMap[colnm]
        dc = dt.Cols[ci]
        dt.Cols[ci] = dc.reshape(shp)

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
    for md in et.MetaData:
        pt.MetaData[md[0]] = md[1]
    return pt
        
def py_to_etable(pt):
    """
    returns an etable.Table version of given PyEtable.
    """
    et = etable.Table()
    et.Rows = pt.Rows
    nc = len(pt.Cols)
    for ci in range(nc):
        pc = pt.Cols[ci]
        cn = pt.ColNames[ci]
        tsr = numpy_to_etensor(pc)
        et.AddCol(tsr, cn)
    for md in pt.MetaData:
        et.SetMetaData(md, pt.MetaData[md])
    return et

def copy_etable_to_py(pt, et):
    """
    copies values in columns of same name from etable to PyEtable
    """
    nc = len(pt.Cols)
    for ci in range(nc):
        pc = pt.Cols[ci]
        cn = pt.ColNames[ci]
        try:
            dc = et.ColByNameTry(cn)
            copy_etensor_to_numpy(pc, dc)
        except:
            pass

def copy_py_to_etable(et, pt):
    """
    copies values in columns of same name from PyEtable to etable
    """
    nc = len(et.Cols)
    for ci in range(nc):
        dc = et.Cols[ci]
        cn = et.ColNames[ci]
        try:
            pc = pt.ColByName(cn)
            copy_numpy_to_etensor(dc, pc)
        except:
            pass
    
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

def etable_to_pandas(et, skip_tensors=False):
    """
    returns a pandas DataFrame constructed from the columns
    of the given PyEtable, spreading tensor cells over sequential
    1d columns, if they aren't skipped over.
    """
    ed = {} 
    nc = len(et.Cols)
    for ci in range(nc):
        dc = et.Cols[ci]
        cn = et.ColNames[ci]
        if dc.ndim == 1:
            ed[cn] = dc
            continue
        if skip_tensors:
            continue
        csz = int(dc.size / et.Rows)  # cell size
        rs = dc.reshape([et.Rows, csz])
        for i in range(csz):
            cnn = "%s_%d" % (cn, i)
            ed[cnn] = rs[:,i]
    df = pd.DataFrame(data=ed)
    return df

def pandas_to_etable(df):
    """
    returns a PyEtable constructed from given pandas DataFrame
    """
    pt = PyEtable()
    pt.Rows = len(df.index)
    for cn in df.columns:
        dc = df.loc[:, cn].values
        pt.AddCol(dc, cn)
    return pt
    
