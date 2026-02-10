"""Console entry point for KothaSet.

This module locates the bundled Go binary and executes it,
passing through all command-line arguments and exit code.
"""

from __future__ import annotations

import os
import subprocess
import sys


def main() -> None:
    """Run the KothaSet Go binary with the given CLI arguments."""
    from kothaset import find_binary

    binary = find_binary()

    # Ensure the binary is executable (no-op on Windows)
    if sys.platform != "win32":
        current = os.stat(binary).st_mode
        if not current & 0o111:
            os.chmod(binary, current | 0o555)

    try:
        result = subprocess.run(
            [binary, *sys.argv[1:]],
            stdin=sys.stdin,
            stdout=sys.stdout,
            stderr=sys.stderr,
        )
        raise SystemExit(result.returncode)
    except KeyboardInterrupt:
        raise SystemExit(130)
    except FileNotFoundError:
        print(
            "Error: KothaSet binary not found.\n"
            "Try reinstalling: pip install --force-reinstall kothaset",
            file=sys.stderr,
        )
        raise SystemExit(1)


if __name__ == "__main__":
    main()
