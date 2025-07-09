"""cx_Freeze setup script for Grimoire"""

import sys
from cx_Freeze import setup, Executable

# Dependencies
build_exe_options = {
    "packages": ["os", "sys", "click", "PIL", "cv2", "numpy", "grimoire"],
    "includes": ["grimoire.cli", "grimoire.mock_compiler"],
    "include_files": [],
    "excludes": ["tkinter", "test", "unittest"],
}

# Base for Windows GUI apps
base = None
if sys.platform == "win32":
    base = None  # Use None for console app

executables = [
    Executable(
        "src/grimoire/cli.py",
        base=base,
        target_name="grimoire",
    )
]

setup(
    name="grimoire",
    version="0.1.0",
    description="Grimoire Visual Programming Language",
    options={"build_exe": build_exe_options},
    executables=executables,
)