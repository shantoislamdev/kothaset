#!/usr/bin/env python3
"""Build platform-specific wheel packages for KothaSet.

This script downloads GoReleaser-built binaries from GitHub Releases,
and creates a platform-specific .whl file for each OS/architecture target.

Usage:
    python build_wheels.py --version 1.0.0
    python build_wheels.py --version 1.0.0 --dry-run
    python build_wheels.py --version 1.0.0 --binaries-dir ./dist
"""

from __future__ import annotations

import argparse
import base64
import hashlib
import os
import re
import shutil
import stat
import sys
import tarfile
import tempfile
import urllib.request
import zipfile
from pathlib import Path

# --------------------------------------------------------------------------- #
#  Constants
# --------------------------------------------------------------------------- #

PACKAGE_NAME = "kothaset"
REPO = "shantoislamdev/kothaset"
SRC_DIR = Path(__file__).resolve().parent.parent / "src" / "kothaset"

# Mapping from GoReleaser targets to Python wheel platform tags.
PLATFORM_TARGETS = [
    {
        "goos": "linux",
        "goarch": "amd64",
        "archive_ext": "tar.gz",
        "binary_name": "kothaset",
        "wheel_platform": "manylinux_2_17_x86_64.manylinux2014_x86_64",
    },
    {
        "goos": "linux",
        "goarch": "arm64",
        "archive_ext": "tar.gz",
        "binary_name": "kothaset",
        "wheel_platform": "manylinux_2_17_aarch64.manylinux2014_aarch64",
    },
    {
        "goos": "darwin",
        "goarch": "amd64",
        "archive_ext": "tar.gz",
        "binary_name": "kothaset",
        "wheel_platform": "macosx_10_12_x86_64",
    },
    {
        "goos": "darwin",
        "goarch": "arm64",
        "archive_ext": "tar.gz",
        "binary_name": "kothaset",
        "wheel_platform": "macosx_11_0_arm64",
    },
    {
        "goos": "windows",
        "goarch": "amd64",
        "archive_ext": "zip",
        "binary_name": "kothaset.exe",
        "wheel_platform": "win_amd64",
    },
]


# --------------------------------------------------------------------------- #
#  Helpers
# --------------------------------------------------------------------------- #


def sha256_digest(filepath: Path) -> str:
    h = hashlib.sha256()
    with open(filepath, "rb") as f:
        for chunk in iter(lambda: f.read(8192), b""):
            h.update(chunk)
    return h.hexdigest()


def urlsafe_b64_nopad(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).rstrip(b"=").decode("ascii")


def record_hash(filepath: Path) -> str:
    """Return sha256=<urlsafe-b64> hash for a RECORD entry."""
    h = hashlib.sha256()
    with open(filepath, "rb") as f:
        for chunk in iter(lambda: f.read(8192), b""):
            h.update(chunk)
    return f"sha256={urlsafe_b64_nopad(h.digest())}"


def download_archive(version: str, goos: str, goarch: str, ext: str, dest: Path) -> Path:
    """Download a GoReleaser archive from GitHub Releases."""
    filename = f"kothaset_{version}_{goos}_{goarch}.{ext}"
    url = f"https://github.com/{REPO}/releases/download/v{version}/{filename}"
    dest_path = dest / filename

    print(f"  Downloading: {url}")
    urllib.request.urlretrieve(url, dest_path)
    return dest_path


def extract_binary(archive_path: Path, binary_name: str, dest_dir: Path) -> Path:
    """Extract the Go binary from an archive."""
    dest_binary = dest_dir / binary_name

    if archive_path.suffix == ".zip" or str(archive_path).endswith(".zip"):
        with zipfile.ZipFile(archive_path) as zf:
            for name in zf.namelist():
                if os.path.basename(name) == binary_name:
                    with zf.open(name) as src, open(dest_binary, "wb") as dst:
                        dst.write(src.read())
                    break
            else:
                raise FileNotFoundError(f"{binary_name} not found in {archive_path}")
    else:
        with tarfile.open(archive_path) as tf:
            for member in tf.getmembers():
                if os.path.basename(member.name) == binary_name:
                    f = tf.extractfile(member)
                    if f is None:
                        raise FileNotFoundError(f"Cannot extract {binary_name}")
                    with open(dest_binary, "wb") as dst:
                        dst.write(f.read())
                    break
            else:
                raise FileNotFoundError(f"{binary_name} not found in {archive_path}")

    # Make executable
    dest_binary.chmod(dest_binary.stat().st_mode | stat.S_IEXEC | stat.S_IXGRP | stat.S_IXOTH)
    return dest_binary


def copy_binary_from_local(binaries_dir: Path, goos: str, goarch: str, binary_name: str, dest_dir: Path) -> Path:
    """Copy a pre-downloaded binary from a local directory."""
    # Try common naming patterns from GoReleaser dist/
    patterns = [
        binaries_dir / f"kothaset_{goos}_{goarch}" / binary_name,
        binaries_dir / f"kothaset_{goos}_{goarch}_v1" / binary_name,
        binaries_dir / binary_name,
    ]
    for src in patterns:
        if src.exists():
            dest = dest_dir / binary_name
            shutil.copy2(src, dest)
            dest.chmod(dest.stat().st_mode | stat.S_IEXEC)
            return dest

    raise FileNotFoundError(
        f"Binary not found for {goos}/{goarch}. Searched:\n"
        + "\n".join(f"  {p}" for p in patterns)
    )


# --------------------------------------------------------------------------- #
#  Wheel builder
# --------------------------------------------------------------------------- #


def build_wheel(
    version: str,
    platform_tag: str,
    binary_path: Path,
    binary_name: str,
    output_dir: Path,
    dry_run: bool = False,
) -> Path:
    """Build a platform-specific wheel containing the Go binary."""
    # Normalize version for wheel filename (PEP 440)
    wheel_version = version

    # Wheel filename: {name}-{ver}-{python}-{abi}-{platform}.whl
    wheel_name = f"{PACKAGE_NAME}-{wheel_version}-py3-none-{platform_tag}.whl"
    wheel_path = output_dir / wheel_name

    if dry_run:
        print(f"  [DRY-RUN] Would create: {wheel_name}")
        print(f"             Binary: {binary_path} ({binary_path.stat().st_size:,} bytes)")
        return wheel_path

    dist_info = f"{PACKAGE_NAME}-{wheel_version}.dist-info"

    with tempfile.TemporaryDirectory() as tmpdir:
        tmp = Path(tmpdir)

        # 1. Copy Python source files
        pkg_dir = tmp / PACKAGE_NAME
        pkg_dir.mkdir()
        for py_file in SRC_DIR.glob("*.py"):
            shutil.copy2(py_file, pkg_dir / py_file.name)

        # Update version in __init__.py (replace placeholder with actual version)
        init_file = pkg_dir / "__init__.py"
        content = init_file.read_text()
        # Replace any version placeholder (handles both "0.0.0" and any prior version)
        content = re.sub(
            r'__version__\s*=\s*"[^"]*".*',
            f'__version__ = "{version}"',
            content,
        )
        init_file.write_text(content)

        # 2. Copy the Go binary into the package
        dest_binary = pkg_dir / binary_name
        shutil.copy2(binary_path, dest_binary)
        dest_binary.chmod(dest_binary.stat().st_mode | stat.S_IEXEC | stat.S_IXGRP | stat.S_IXOTH)

        # 3. Create dist-info
        info_dir = tmp / dist_info
        info_dir.mkdir()

        # METADATA
        (info_dir / "METADATA").write_text(
            f"Metadata-Version: 2.1\n"
            f"Name: {PACKAGE_NAME}\n"
            f"Version: {wheel_version}\n"
            f"Summary: High-quality dataset generation CLI for LLM training\n"
            f"Home-page: https://github.com/{REPO}\n"
            f"Author: Shanto Islam\n"
            f"Author-email: shantoislamdev@gmail.com\n"
            f"License: Apache-2.0\n"
            f"Requires-Python: >=3.8\n"
            f"Classifier: Development Status :: 5 - Production/Stable\n"
            f"Classifier: Environment :: Console\n"
            f"Classifier: Intended Audience :: Developers\n"
            f"Classifier: Intended Audience :: Science/Research\n"
            f"Classifier: License :: OSI Approved :: Apache Software License\n"
            f"Classifier: Topic :: Scientific/Engineering :: Artificial Intelligence\n"
        )

        # WHEEL
        (info_dir / "WHEEL").write_text(
            f"Wheel-Version: 1.0\n"
            f"Generator: kothaset-build-wheels\n"
            f"Root-Is-Purelib: false\n"
            f"Tag: py3-none-{platform_tag}\n"
        )

        # entry_points.txt
        (info_dir / "entry_points.txt").write_text(
            "[console_scripts]\n"
            "kothaset = kothaset._main:main\n"
        )

        # top_level.txt
        (info_dir / "top_level.txt").write_text(f"{PACKAGE_NAME}\n")

        # RECORD (must be last — lists all files with hashes)
        record_lines = []
        all_files = list(pkg_dir.rglob("*")) + list(info_dir.rglob("*"))
        for filepath in all_files:
            if filepath.is_file() and filepath.name != "RECORD":
                rel = filepath.relative_to(tmp)
                size = filepath.stat().st_size
                h = record_hash(filepath)
                record_lines.append(f"{rel},{h},{size}")

        # RECORD itself has no hash
        record_lines.append(f"{dist_info}/RECORD,,")
        (info_dir / "RECORD").write_text("\n".join(record_lines) + "\n")

        # 4. Build the .whl (it's a zip file)
        print(f"  Building: {wheel_name}")
        with zipfile.ZipFile(wheel_path, "w", zipfile.ZIP_DEFLATED) as whl:
            for filepath in sorted(tmp.rglob("*")):
                if filepath.is_file():
                    arcname = str(filepath.relative_to(tmp))
                    # Use forward slashes in zip
                    arcname = arcname.replace("\\", "/")
                    whl.write(filepath, arcname)

    size_mb = wheel_path.stat().st_size / (1024 * 1024)
    print(f"  Created:  {wheel_name} ({size_mb:.1f} MB)")
    return wheel_path


# --------------------------------------------------------------------------- #
#  Main
# --------------------------------------------------------------------------- #


def main():
    parser = argparse.ArgumentParser(
        description="Build platform-specific Python wheels for KothaSet"
    )
    parser.add_argument(
        "--version", required=True, help="Release version (e.g. 1.0.0)"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show what would be built without actually building",
    )
    parser.add_argument(
        "--binaries-dir",
        type=Path,
        default=None,
        help="Local directory containing pre-built binaries (GoReleaser dist/). "
        "If not set, binaries are downloaded from GitHub Releases.",
    )
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=Path(__file__).resolve().parent.parent / "dist",
        help="Output directory for wheel files (default: pip/dist/)",
    )
    parser.add_argument(
        "--platforms",
        nargs="*",
        default=None,
        help="Specific platforms to build (e.g. linux-amd64 darwin-arm64). "
        "Default: all platforms.",
    )
    args = parser.parse_args()

    # Filter platforms if specified
    targets = PLATFORM_TARGETS
    if args.platforms:
        requested = {p.replace("-", "/") for p in args.platforms}
        targets = [t for t in targets if f"{t['goos']}/{t['goarch']}" in requested]
        if not targets:
            print(f"Error: No matching platforms found for: {args.platforms}")
            sys.exit(1)

    args.output_dir.mkdir(parents=True, exist_ok=True)

    print(f"\n📦 Building KothaSet v{args.version} wheels\n")
    print(f"   Source:  {SRC_DIR}")
    print(f"   Output:  {args.output_dir}")
    print(f"   Targets: {len(targets)} platforms\n")

    built_wheels = []

    with tempfile.TemporaryDirectory() as tmpdir:
        tmp = Path(tmpdir)

        for target in targets:
            goos = target["goos"]
            goarch = target["goarch"]
            label = f"{goos}/{goarch}"

            print(f"▸ {label} → {target['wheel_platform']}")

            if args.dry_run:
                # Create a dummy binary for dry-run size estimation
                dummy = tmp / target["binary_name"]
                dummy.write_bytes(b"\x00" * 100)
                wheel = build_wheel(
                    version=args.version,
                    platform_tag=target["wheel_platform"],
                    binary_path=dummy,
                    binary_name=target["binary_name"],
                    output_dir=args.output_dir,
                    dry_run=True,
                )
            else:
                # Get the binary
                binary_dir = tmp / f"{goos}_{goarch}"
                binary_dir.mkdir(exist_ok=True)

                if args.binaries_dir:
                    binary_path = copy_binary_from_local(
                        args.binaries_dir, goos, goarch,
                        target["binary_name"], binary_dir,
                    )
                else:
                    archive_path = download_archive(
                        args.version, goos, goarch,
                        target["archive_ext"], tmp,
                    )
                    binary_path = extract_binary(
                        archive_path, target["binary_name"], binary_dir,
                    )

                wheel = build_wheel(
                    version=args.version,
                    platform_tag=target["wheel_platform"],
                    binary_path=binary_path,
                    binary_name=target["binary_name"],
                    output_dir=args.output_dir,
                )
                built_wheels.append(wheel)

            print()

    if not args.dry_run and built_wheels:
        print("✅ All wheels built successfully!\n")
        print("Upload to PyPI with:")
        print(f"  twine upload {args.output_dir}/*.whl\n")
    elif args.dry_run:
        print("✅ Dry run complete. No files were created.\n")


if __name__ == "__main__":
    main()
