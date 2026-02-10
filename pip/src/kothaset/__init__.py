"""KothaSet — High-quality dataset generation CLI for LLM training."""

from __future__ import annotations

import os
import sys

__version__ = "0.0.0"  # Auto-set from git tag during CI build


def find_binary() -> str:
    """Locate the kothaset binary bundled with this package.

    Returns:
        Absolute path to the kothaset binary.

    Raises:
        FileNotFoundError: If the binary is not found in the package.
    """
    binary_name = "kothaset.exe" if sys.platform == "win32" else "kothaset"
    package_dir = os.path.dirname(os.path.abspath(__file__))
    binary_path = os.path.join(package_dir, binary_name)

    if os.path.isfile(binary_path):
        return binary_path

    raise FileNotFoundError(
        f"KothaSet binary not found at {binary_path}.\n"
        "This likely means the package was not installed correctly.\n"
        "Try reinstalling: pip install --force-reinstall kothaset"
    )
