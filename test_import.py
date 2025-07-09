#!/usr/bin/env python3
"""インポートテストスクリプト"""

import sys
print(f"Python version: {sys.version}")
print(f"Python path: {sys.path}")

try:
    import grimoire
    print(f"[OK] grimoire imported successfully: {grimoire.__version__}")
except ImportError as e:
    print(f"[FAIL] Failed to import grimoire: {e}")
    sys.exit(1)

try:
    from grimoire import cli
    print("[OK] grimoire.cli imported successfully")
except ImportError as e:
    print(f"[FAIL] Failed to import grimoire.cli: {e}")
    sys.exit(1)

try:
    from grimoire import mock_compiler
    print("[OK] grimoire.mock_compiler imported successfully")
except ImportError as e:
    print(f"[FAIL] Failed to import grimoire.mock_compiler: {e}")
    sys.exit(1)

print("\nAll imports successful!")