#!/usr/bin/env python3
"""Build a native YAML Viewer production binary."""

from __future__ import annotations

import argparse
import os
import subprocess
import sys
from pathlib import Path
from typing import Mapping, Sequence


SCRIPT_DIR = Path(__file__).resolve().parent
PROJECT_ROOT = SCRIPT_DIR.parent


def run_command(
    command: Sequence[str],
    *,
    cwd: Path | None = None,
    env: Mapping[str, str] | None = None,
) -> subprocess.CompletedProcess[str]:
    """Run a command without invoking a shell."""

    return subprocess.run(
        list(command),
        check=True,
        cwd=cwd,
        env=env,
        text=True,
    )


def go_env(name: str, *, cwd: Path = PROJECT_ROOT) -> str:
    """Read one value from the active Go toolchain."""

    return subprocess.check_output(
        ["go", "env", name], cwd=cwd, text=True
    ).strip()


def production_ldflags(target_os: str) -> str:
    """Return production linker flags for one target operating system."""

    flags = ["-s", "-w"]
    if target_os == "windows":
        flags.insert(0, "-H=windowsgui")
    return " ".join(flags)


def production_build_command(
    *,
    target_os: str,
    output: Path | None = None,
    package: str = ".",
) -> list[str]:
    """Create the Go production build command for a target platform."""

    command = [
        "go",
        "build",
        "-trimpath",
        "-ldflags",
        production_ldflags(target_os),
    ]
    if output is not None:
        command.extend(["-o", str(output)])
    command.append(package)
    return command


def build_environment(
    *,
    target_os: str,
    target_arch: str,
    base: Mapping[str, str] | None = None,
) -> dict[str, str]:
    """Create an isolated Go/cgo environment for a target platform."""

    environment = dict(os.environ if base is None else base)
    environment.update(
        {
            "CGO_ENABLED": "1",
            "GOOS": target_os,
            "GOARCH": target_arch,
        }
    )
    return environment


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Build the YAML Viewer native production binary."
    )
    parser.add_argument(
        "--arch",
        default=os.environ.get("TARGET_ARCH"),
        help="Target architecture (default: the local Go architecture).",
    )
    return parser.parse_args()


def build(args: argparse.Namespace) -> None:
    target_os = go_env("GOOS")
    target_arch = args.arch or go_env("GOARCH")
    environment = build_environment(
        target_os=target_os,
        target_arch=target_arch,
    )
    print(f"Building YAML Viewer for {target_os}/{target_arch}...")
    run_command(
        production_build_command(target_os=target_os),
        cwd=PROJECT_ROOT,
        env=environment,
    )


def main() -> int:
    try:
        build(parse_args())
    except (OSError, RuntimeError, subprocess.CalledProcessError) as error:
        print(f"Build failed: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
