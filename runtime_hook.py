"""Runtime hook to optimize OpenCV initialization"""
import os
import sys

# OpenCVの不要な機能を無効化
os.environ['OPENCV_VIDEOIO_PRIORITY_BACKEND'] = '0'
os.environ['OPENCV_OPENCL_RUNTIME'] = ''
os.environ['OPENCV_OPENCL_DEVICE'] = 'disabled'
os.environ['OPENCV_DNN_OPENCL_ALLOW_ALL_DEVICES'] = '0'

# NumPyのマルチスレッドを制限（起動時間短縮のため）
os.environ['OMP_NUM_THREADS'] = '1'
os.environ['OPENBLAS_NUM_THREADS'] = '1'
os.environ['MKL_NUM_THREADS'] = '1'
os.environ['VECLIB_MAXIMUM_THREADS'] = '1'
os.environ['NUMEXPR_NUM_THREADS'] = '1'

# cv2を事前にインポートして初期化を早める
try:
    import cv2
    # 基本的な設定のみ
    cv2.setNumThreads(1)
    cv2.setUseOptimized(True)
except:
    pass