"""Lazy import system to speed up startup time"""

import sys
from importlib import import_module


class LazyImport:
    """Lazy import wrapper that delays actual import until first use"""
    
    def __init__(self, module_name):
        self.module_name = module_name
        self._module = None
    
    def __getattr__(self, name):
        if self._module is None:
            self._module = import_module(self.module_name)
        return getattr(self._module, name)
    
    def __dir__(self):
        if self._module is None:
            self._module = import_module(self.module_name)
        return dir(self._module)


# Create lazy imports
cv2 = None
np = None

def init_cv2():
    """Initialize cv2 when needed"""
    global cv2
    if cv2 is None:
        cv2 = import_module('cv2')
    return cv2

def init_numpy():
    """Initialize numpy when needed"""
    global np
    if np is None:
        np = import_module('numpy')
    return np