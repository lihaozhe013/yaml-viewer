#!/usr/bin/env python3
"""Build a Windows installer for YAML Viewer with Inno Setup 6."""

from __future__ import annotations

import argparse
import ctypes
import os
import re
import shutil
import subprocess
import sys
import tempfile
import textwrap
from pathlib import Path
from typing import Sequence


APP_NAME = "YAML Viewer"
APP_EXECUTABLE = "yamlviewer.exe"
APP_IDENTIFIER = "com.yamlviewer.app"
APP_PUBLISHER = "YAML Viewer"
ICON_PNG_NAME = "Icon.png"
ICON_ICO_NAME = "Icon.ico"
ICON_SIZES = (256, 128, 96, 64, 48, 32, 16)
SCRIPT_DIR = Path(__file__).resolve().parent
PROJECT_ROOT = SCRIPT_DIR.parent


def run_command(
    command: Sequence[str],
    *,
    cwd: Path | None = None,
    env: dict[str, str] | None = None,
) -> subprocess.CompletedProcess[str]:
    """Run a command without invoking a shell."""

    return subprocess.run(
        list(command),
        check=True,
        cwd=cwd,
        env=env,
        text=True,
    )


def require_command(command: str) -> None:
    if shutil.which(command) is None:
        raise RuntimeError(f"Required command not found: {command}")


def metadata_value(name: str, fallback: str) -> str:
    metadata_path = PROJECT_ROOT / "FyneApp.toml"
    if not metadata_path.is_file():
        return fallback

    content = metadata_path.read_text(encoding="utf-8")
    match = re.search(
        rf'(?m)^{re.escape(name)}\s*=\s*(?:"([^"]+)"|([^\s#]+))\s*$',
        content,
    )
    if not match:
        return fallback
    return match.group(1) or match.group(2)


def short_windows_path(path: Path) -> str:
    """Return an 8.3 path when Windows toolchains contain spaces."""

    path_string = str(path)
    if sys.platform != "win32" or " " not in path_string:
        return path_string

    kernel32 = ctypes.WinDLL("kernel32", use_last_error=True)
    length = kernel32.GetShortPathNameW(path_string, None, 0)
    if not length:
        return path_string

    buffer = ctypes.create_unicode_buffer(length + 1)
    result = kernel32.GetShortPathNameW(path_string, buffer, length + 1)
    return buffer.value if result else path_string


def configure_cgo_environment(env: dict[str, str]) -> None:
    """Keep MinGW's GCC and search prefix free of spaces on Windows."""

    compiler = env.get("CC", "gcc")
    compiler_path = shutil.which(compiler)
    if not compiler_path or Path(compiler_path).stem.lower() != "gcc":
        return

    gcc_path = Path(compiler_path)
    short_gcc_path = short_windows_path(gcc_path)
    gcc_prefix = gcc_path.parent.parent / "lib" / "gcc"
    if short_gcc_path == str(gcc_path) or not gcc_prefix.is_dir():
        return

    env["CC"] = short_gcc_path
    short_prefix = short_windows_path(gcc_prefix)
    env["GCC_EXEC_PREFIX"] = short_prefix.rstrip("\\/") + "\\"
    print(f"[windows_packaging] Using short GCC path: {short_gcc_path}")


def write_icon(png_path: Path, ico_path: Path) -> None:
    """Generate a multi-resolution Icon.ico from the repository's PNG icon."""

    try:
        from PIL import Image
    except ImportError as error:
        raise RuntimeError(
            "Pillow is required to convert Icon.png to Icon.ico. "
            "Install it with `uv pip install Pillow` or use "
            "`uv run --with Pillow scripts/build_windows.py`."
        ) from error

    with Image.open(png_path) as source:
        source.load()
        sizes = [(size, size) for size in ICON_SIZES]
        source.save(ico_path, format="ICO", sizes=sizes)
    print(f"[windows_packaging] Generated icon: {ico_path}")


def find_rsrc() -> Path:
    go_bin = subprocess.check_output(
        ["go", "env", "GOPATH"], cwd=PROJECT_ROOT, text=True
    ).strip()
    candidates = [Path(go_bin) / "bin" / "rsrc.exe", Path(go_bin) / "bin" / "rsrc"]
    for candidate in candidates:
        if candidate.is_file():
            return candidate

    on_path = shutil.which("rsrc")
    if on_path:
        return Path(on_path)

    raise RuntimeError(
        "rsrc tool not found. Install it with "
        "`go install github.com/akavel/rsrc@latest`."
    )


def embed_icon(
    *,
    rsrc: Path,
    icon: Path,
    arch: str,
    output: Path,
) -> None:
    run_command([str(rsrc), "-arch", arch, "-ico", str(icon), "-o", str(output)])
    print(f"[windows_packaging] Generated resource object: {output}")


def find_iscc() -> Path:
    for command in ("ISCC.exe", "ISCC"):
        executable = shutil.which(command)
        if executable:
            return Path(executable)

    candidates: list[Path] = []
    for variable in ("ProgramFiles(x86)", "ProgramFiles"):
        root = os.environ.get(variable)
        if root:
            candidates.append(Path(root) / "Inno Setup 6" / "ISCC.exe")

    local_app_data = os.environ.get("LocalAppData")
    if local_app_data:
        candidates.append(
            Path(local_app_data) / "Programs" / "Inno Setup 6" / "ISCC.exe"
        )

    for candidate in candidates:
        if candidate.is_file():
            return candidate

    raise RuntimeError(
        "Inno Setup 6 compiler not found. Install it from "
        "https://jrsoftware.org/isdl.php or add ISCC.exe to PATH."
    )


def parse_args() -> argparse.Namespace:
    default_output = os.environ.get(
        "DIST_DIR", str(PROJECT_ROOT / "dist" / "windows")
    )

    parser = argparse.ArgumentParser(
        description="Build the YAML Viewer Windows executable and installer."
    )
    parser.add_argument(
        "--output-dir",
        default=default_output,
        help="Output directory (default: dist/windows).",
    )
    parser.add_argument(
        "--arch",
        default=os.environ.get("TARGET_ARCH"),
        help="Target architecture (default: the local Go architecture).",
    )
    parser.add_argument(
        "--version",
        default=os.environ.get("APP_VERSION"),
        help="Application version (default: FyneApp.toml).",
    )
    parser.add_argument(
        "--build",
        default=os.environ.get("APP_BUILD"),
        help="Application build number (default: FyneApp.toml).",
    )
    return parser.parse_args()


def windows_file_version(version: str, build_number: str) -> str:
    parts = version.split(".")
    if not 1 <= len(parts) <= 3 or any(not part.isdigit() for part in parts):
        raise RuntimeError(
            "Application version must contain one to three numeric components."
        )
    if not build_number.isdigit():
        raise RuntimeError("Application build number must be numeric.")

    components = [int(part) for part in parts]
    components.extend([0] * (3 - len(components)))
    components.append(int(build_number))
    if any(component > 65535 for component in components):
        raise RuntimeError("Version and build components must not exceed 65535.")
    return ".".join(str(component) for component in components)


def inno_path(path: Path) -> str:
    return str(path.resolve()).replace("\\", "/").replace('"', '""')


def write_installer_script(
    path: Path,
    *,
    executable: Path,
    output_dir: Path,
    version: str,
    file_version: str,
    icon_path: Path,
) -> None:
    script = rf"""
        [Setup]
        AppId={APP_IDENTIFIER}
        AppName={APP_NAME}
        AppVersion={version}
        AppPublisher={APP_PUBLISHER}
        DefaultDirName={{autopf}}\{APP_NAME}
        DefaultGroupName={APP_NAME}
        DisableProgramGroupPage=yes
        PrivilegesRequired=lowest
        ArchitecturesAllowed=x64compatible
        ArchitecturesInstallIn64BitMode=x64compatible
        OutputDir={inno_path(output_dir)}
        OutputBaseFilename={APP_NAME}-{version}-Setup
        SetupIconFile={inno_path(icon_path)}
        SetupLogging=yes
        Compression=lzma2
        SolidCompression=yes
        WizardStyle=modern
        MinVersion=10.0
        UninstallDisplayIcon={{app}}\{APP_EXECUTABLE}
        VersionInfoVersion={file_version}
        VersionInfoCompany={APP_PUBLISHER}
        VersionInfoDescription={APP_NAME} installer
        VersionInfoProductName={APP_NAME}
        VersionInfoProductVersion={file_version}

        [Languages]
        Name: "english"; MessagesFile: "compiler:Default.isl"

        [Tasks]
        Name: "desktopicon"; Description: "Create a desktop shortcut"; \
            GroupDescription: "Additional shortcuts:"

        [Files]
        Source: "{inno_path(executable)}"; DestDir: "{{app}}"; Flags: ignoreversion
        Source: "{inno_path(icon_path)}"; DestDir: "{{app}}"; Flags: ignoreversion

        [Icons]
        Name: "{{group}}\{APP_NAME}"; Filename: "{{app}}\{APP_EXECUTABLE}"; \
            IconFilename: "{{app}}\{ICON_ICO_NAME}"
        Name: "{{group}}\{{cm:UninstallProgram,{APP_NAME}}}"; \
            Filename: "{{uninstallexe}}"
        Name: "{{autodesktop}}\{APP_NAME}"; \
            Filename: "{{app}}\{APP_EXECUTABLE}"; \
            IconFilename: "{{app}}\{ICON_ICO_NAME}"; \
            Tasks: desktopicon

        [Run]
        Filename: "{{app}}\{APP_EXECUTABLE}"; \
            Description: "{{cm:LaunchProgram,{APP_NAME}}}"; \
            Flags: nowait postinstall skipifsilent
        """
    path.write_text(textwrap.dedent(script).lstrip(), encoding="utf-8")


def build(args: argparse.Namespace) -> Path:
    if sys.platform != "win32":
        raise RuntimeError("Windows installer packaging must run on Windows.")

    require_command("go")
    iscc = find_iscc()
    rsrc = find_rsrc()

    goos = subprocess.check_output(
        ["go", "env", "GOOS"], cwd=PROJECT_ROOT, text=True
    ).strip()
    if goos != "windows":
        raise RuntimeError("The Go toolchain must target windows.")

    arch = args.arch
    if not arch:
        arch = subprocess.check_output(
            ["go", "env", "GOARCH"], cwd=PROJECT_ROOT, text=True
        ).strip()

    version = args.version or metadata_value("Version", "0.0.1")
    build_number = args.build or metadata_value("Build", "1")
    file_version = windows_file_version(version, build_number)
    output_dir = Path(args.output_dir).resolve()
    output_dir.mkdir(parents=True, exist_ok=True)

    print(
        f"[windows_packaging] Building {APP_NAME} {version} ({build_number}) "
        f"for windows/{arch}..."
    )

    with tempfile.TemporaryDirectory(prefix="yaml-viewer-windows-") as temp_dir:
        staging_dir = Path(temp_dir)
        executable = staging_dir / APP_EXECUTABLE
        installer_script = staging_dir / "installer.iss"
        icon_path = staging_dir / ICON_ICO_NAME
        syso_path = staging_dir / f"rsrc_windows_{arch}.syso"

        write_icon(PROJECT_ROOT / ICON_PNG_NAME, icon_path)
        embed_icon(rsrc=rsrc, icon=icon_path, arch=arch, output=syso_path)

        build_env = os.environ.copy()
        build_env.update({"CGO_ENABLED": "1", "GOARCH": arch})
        configure_cgo_environment(build_env)
        run_command(
            [
                "go",
                "build",
                "-tags",
                "release",
                "-trimpath",
                "-ldflags",
                "-H=windowsgui -s -w",
                "-o",
                str(executable),
                ".",
            ],
            cwd=PROJECT_ROOT,
            env=build_env,
        )

        write_installer_script(
            installer_script,
            executable=executable,
            output_dir=output_dir,
            version=version,
            file_version=file_version,
            icon_path=icon_path,
        )
        run_command([str(iscc), str(installer_script)], cwd=PROJECT_ROOT)

    installer = output_dir / f"{APP_NAME}-{version}-Setup.exe"
    if not installer.is_file():
        raise RuntimeError(f"Inno Setup did not create the expected file: {installer}")

    print("\n[windows_packaging] Build completed successfully.")
    print(f"[windows_packaging] Installer: {installer}")
    return installer


def main() -> int:
    try:
        build(parse_args())
    except (OSError, RuntimeError, subprocess.CalledProcessError) as error:
        print(f"[windows_packaging] Error: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
