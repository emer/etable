# Copyright (c) 2020, The Emergent Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# code for converting `etable` to / from a torch `TensorDataset`,
# which has the same structure as an `etable`, and is used in the
# `pytorch` neural network framework.

from leabra import go, etable, etensor

import numpy as np
import torch
import torch.utils.data as data_utils

def etable_to_torch(et):
    """
    returns a torch.utils.data.TensorDataset constructed from the given etable.Table
    """
    lbls = []
    tsrs = []
    nc = len(et.Cols)
    for ci in range(nc):
        dc = et.Cols[ci]
        cn = et.ColNames[ci]
        nar = 0
        if dc.DataType() == etensor.FLOAT64:
            nar = np.array(etensor.Float64(dc).Values)
        elif dc.DataType() == etensor.FLOAT32:
            nar = np.array(etensor.Float32(dc).Values)
        elif dc.DataType() == etensor.INT64:
            nar = np.array(etensor.Int64(dc).Values)
        elif dc.DataType() == etensor.INT32:
            nar = np.array(etensor.Int32(dc).Values)
        elif dc.DataType() == etensor.INT:
            nar = np.array(etensor.Int(dc).Values)
        elif dc.DataType() == etensor.STRING:
            nar = np.array(etensor.String(dc).Values)
            nar = nar.reshape(dc.Shapes())
            lbls.append(nar)
            continue
        else:
            print("column %s with type %d cannot be converted" % (cn, dc.DataType()))
            continue
        # there does not appear to be a way to set the shape at the same time as initializing 
        nar = nar.reshape(dc.Shapes())
        tsr = torch.from_numpy(nar)
        # tsr.names=dc.DimNames() # this doesn't work
        tsrs.append(tsr)
    ds = data_utils.TensorDataset(*tsrs)
    return ds
        
        
