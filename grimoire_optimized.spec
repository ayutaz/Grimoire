# -*- mode: python ; coding: utf-8 -*-
import sys
import os
from PyInstaller.utils.hooks import collect_all, collect_submodules

# OpenCV最適化のための設定
block_cipher = None

# OpenCVの最適化
os.environ['OPENCV_VIDEOIO_PRIORITY_BACKEND'] = '0'
os.environ['OPENCV_OPENCL_RUNTIME'] = ''

# 分析
a = Analysis(
    ['src/grimoire/__main__.py'],
    pathex=[],
    binaries=[],
    datas=[
        ('src/grimoire', 'grimoire'),
    ],
    hiddenimports=[
        'cv2',
        'numpy',
        'PIL',
        'PIL.Image',
    ] + collect_submodules('cv2'),
    hookspath=['hooks'],
    hooksconfig={},
    runtime_hooks=['runtime_hook.py'],
    excludes=[
        'matplotlib',
        'scipy',
        'pandas',
        'tkinter',
        'PyQt5',
        'PyQt6',
        'PySide2',
        'PySide6',
        'wx',
        'gtk',
        'gi',
        'IPython',
        'jupyter',
        'notebook',
        'tornado',
        'jedi',
        'parso',
        'multiprocessing',
    ],
    noarchive=False,
)

# バイナリの最適化
pyz = PYZ(a.pure, a.zipped_data, cipher=block_cipher)

exe = EXE(
    pyz,
    a.scripts,
    a.binaries,
    a.datas,
    [],
    name='grimoire',
    debug=False,
    bootloader_ignore_signals=False,
    strip=True,  # シンボルを削除してサイズを削減
    upx=False,   # UPX圧縮を無効化（起動時間を短縮）
    upx_exclude=[],
    runtime_tmpdir=None,
    console=True,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
    icon=None,
)