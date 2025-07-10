#!/usr/bin/env python3
"""
Fix for OpenCV import delay in PyInstaller on macOS.
This script modifies the PyInstaller spec file to handle OpenCV properly.
"""

spec_content = '''# -*- mode: python ; coding: utf-8 -*-
# PyInstaller spec file for Grimoire

import sys
import os
from PyInstaller.utils.hooks import collect_submodules, collect_data_files

block_cipher = None

# Collect all OpenCV submodules
cv2_hidden_imports = collect_submodules('cv2')

a = Analysis(
    ['src/grimoire/__main__.py'],
    pathex=['src'],
    binaries=[],
    datas=[],
    hiddenimports=[
        'grimoire',
        'grimoire.cli',
        'grimoire.mock_compiler',
        'PIL',
        'PIL.Image',
        'PIL.ImageDraw',
        'PIL.ImageFont',
        'numpy',
        'click',
        'pkg_resources.py2_warn',
    ] + cv2_hidden_imports,
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[
        'matplotlib',
        'PyQt5',
        'PyQt4',
        'PySide2',
        'tkinter',
        'wx',
    ],
    win_no_prefer_redirects=False,
    win_private_assemblies=False,
    cipher=block_cipher,
    noarchive=False,
)

# Remove duplicate binaries
seen = set()
new_binaries = []
for b in a.binaries:
    if b[0] not in seen:
        seen.add(b[0])
        new_binaries.append(b)
a.binaries = new_binaries

pyz = PYZ(a.pure, a.zipped_data, cipher=block_cipher)

exe = EXE(
    pyz,
    a.scripts,
    a.binaries,
    a.zipfiles,
    a.datas,
    [],
    name='grimoire',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=False,  # Disable UPX compression for faster startup
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
'''

with open('grimoire_optimized.spec', 'w') as f:
    f.write(spec_content)

print("Created optimized spec file: grimoire_optimized.spec")
print("\nKey optimizations:")
print("1. Disabled UPX compression (upx=False)")
print("2. Added explicit cv2 submodule collection")
print("3. Excluded unnecessary GUI frameworks")
print("4. Removed duplicate binaries")
print("\nNow build with: pyinstaller grimoire_optimized.spec")