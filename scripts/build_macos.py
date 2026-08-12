#!/usr/bin/env python3
"""Build a macOS application bundle and disk image for YAML Viewer."""

from __future__ import annotations

import argparse
import os
import plistlib
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path
from typing import Sequence

from build import build_environment, production_build_command


APP_NAME = "YAML Viewer"
APP_EXECUTABLE = "yaml-viewer"
APP_IDENTIFIER = "com.yamlviewer.app"
SCRIPT_DIR = Path(__file__).resolve().parent
PROJECT_ROOT = SCRIPT_DIR.parent
ICON_SOURCE = PROJECT_ROOT / "Icon.png"


def run_command(
    command: Sequence[str],
    *,
    env: dict[str, str] | None = None,
    quiet: bool = False,
) -> subprocess.CompletedProcess[str]:
    """Run a command without invoking a shell."""

    return subprocess.run(
        list(command),
        check=True,
        env=env,
        stdout=subprocess.DEVNULL if quiet else None,
        text=True,
    )


def require_command(command: str) -> None:
    if shutil.which(command) is None:
        raise RuntimeError(f"Required command not found: {command}")


def remove_generated_path(path: Path) -> None:
    """Remove one known generated artifact without shell expansion."""

    if path.is_symlink() or path.is_file():
        path.unlink()
    elif path.is_dir():
        shutil.rmtree(path)


def metadata_value(name: str, fallback: str) -> str:
    metadata_path = PROJECT_ROOT / "FyneApp.toml"
    if not metadata_path.is_file():
        return fallback

    content = metadata_path.read_text(encoding="utf-8")
    match = re.search(rf"(?m)^{name} = [\"]([^\"]+)[\"]$", content)
    return match.group(1) if match else fallback


def parse_args() -> argparse.Namespace:
    default_output = os.environ.get(
        "DIST_DIR", str(PROJECT_ROOT / "dist" / "macos")
    )
    default_arch = os.environ.get("TARGET_ARCH")
    default_version = os.environ.get("APP_VERSION")
    default_build = os.environ.get("APP_BUILD")

    parser = argparse.ArgumentParser(
        description="Build the YAML Viewer macOS app and DMG."
    )
    parser.add_argument(
        "--output-dir",
        default=default_output,
        help="Output directory (default: dist/macos).",
    )
    parser.add_argument(
        "--arch",
        default=default_arch,
        help="Target architecture (default: the local Go architecture).",
    )
    parser.add_argument(
        "--version",
        default=default_version,
        help="Application version (default: FyneApp.toml).",
    )
    parser.add_argument(
        "--build",
        default=default_build,
        help="Application build number (default: FyneApp.toml).",
    )
    return parser.parse_args()


def create_icon(icon_source: Path, iconset_dir: Path, output: Path) -> None:
    iconset_dir.mkdir(parents=True, exist_ok=True)
    icon_sizes = ((16, 32), (32, 64), (128, 256), (256, 512), (512, 1024))
    for icon_size, retina_size in icon_sizes:
        regular_path = iconset_dir / f"icon_{icon_size}x{icon_size}.png"
        retina_path = iconset_dir / f"icon_{icon_size}x{icon_size}@2x.png"
        run_command(
            [
                "sips",
                "-z",
                str(icon_size),
                str(icon_size),
                str(icon_source),
                "--out",
                str(regular_path),
            ],
            quiet=True,
        )
        run_command(
            [
                "sips",
                "-z",
                str(retina_size),
                str(retina_size),
                str(icon_source),
                "--out",
                str(retina_path),
            ],
            quiet=True,
        )

    run_command(
        ["iconutil", "-c", "icns", str(iconset_dir), "-o", str(output)],
        quiet=True,
    )


def write_info_plist(path: Path, version: str, build: str) -> None:
    info = {
        "CFBundleDisplayName": APP_NAME,
        "CFBundleExecutable": APP_EXECUTABLE,
        "CFBundleIconFile": "icon.icns",
        "CFBundleIdentifier": APP_IDENTIFIER,
        "CFBundleInfoDictionaryVersion": "6.0",
        "CFBundleName": APP_NAME,
        "CFBundlePackageType": "APPL",
        "CFBundleShortVersionString": version,
        "CFBundleSupportedPlatforms": ["MacOSX"],
        "CFBundleVersion": build,
        "LSMinimumSystemVersion": "10.11",
        "NSHighResolutionCapable": True,
        "NSSupportsAutomaticGraphicsSwitching": True,
    }
    with path.open("wb") as plist_file:
        plistlib.dump(info, plist_file, fmt=plistlib.FMT_XML, sort_keys=False)


def quarantine_script() -> str:
    return '''#!/usr/bin/env python3
"""Remove the macOS quarantine attribute from YAML Viewer."""

from __future__ import annotations

import subprocess
import sys
from pathlib import Path


APP_NAME = "YAML Viewer.app"


def main() -> int:
    script_dir = Path(__file__).resolve().parent
    installed_app = Path("/Applications") / APP_NAME
    bundled_app = script_dir / APP_NAME

    if installed_app.is_dir():
        app_path = installed_app
    elif bundled_app.is_dir():
        app_path = bundled_app
    else:
        print(
            f"Error: {APP_NAME} was not found in /Applications or next to "
            "this script.",
            file=sys.stderr,
        )
        print(
            "Copy the application to /Applications or keep it beside this "
            "script, then try again.",
            file=sys.stderr,
        )
        return 1

    print("This will remove the macOS quarantine attribute from:")
    print(app_path)
    print("Administrator permission is required.")
    print("\\nRequesting administrator permission...", flush=True)
    try:
        subprocess.run(
            [
                "sudo",
                "-k",
                "-p",
                "Enter your macOS administrator password: ",
                "xattr",
                "-rd",
                "com.apple.quarantine",
                str(app_path),
            ],
            check=True,
        )
    except (FileNotFoundError, subprocess.CalledProcessError):
        print("\\nFailed to remove the quarantine attribute.", file=sys.stderr)
        return 1

    print("\\nQuarantine attribute removed successfully.")
    print(f"You can now open {APP_NAME}.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
'''


def build(args: argparse.Namespace) -> tuple[Path, Path]:
    if sys.platform != "darwin":
        raise RuntimeError("This script must be run on macOS.")

    for command in ("go", "sips", "iconutil", "hdiutil"):
        require_command(command)
    if not ICON_SOURCE.is_file():
        raise RuntimeError(f"Application icon not found: {ICON_SOURCE}")

    arch = args.arch
    if not arch:
        arch = subprocess.check_output(
            ["go", "env", "GOARCH"], text=True
        ).strip()
    version = args.version or metadata_value("Version", "0.0.1")
    build_number = args.build or metadata_value("Build", "1")
    output_dir = Path(args.output_dir).expanduser()
    if not output_dir.is_absolute():
        output_dir = (Path.cwd() / output_dir).resolve()

    app_bundle = output_dir / f"{APP_NAME}.app"
    dmg_path = output_dir / f"{APP_NAME}.dmg"

    print(
        f"Building {APP_NAME} {version} ({build_number}) "
        f"for darwin/{arch}..."
    )
    output_dir.mkdir(parents=True, exist_ok=True)

    with tempfile.TemporaryDirectory(prefix="yaml-viewer-macos-") as temp:
        temp_dir = Path(temp)
        binary_path = temp_dir / APP_EXECUTABLE
        bundle_source = temp_dir / f"{APP_NAME}.app"
        staging_dir = temp_dir / APP_NAME
        iconset_dir = temp_dir / f"{APP_NAME}.iconset"

        build_env = build_environment(
            target_os="darwin",
            target_arch=arch,
        )
        run_command(
            production_build_command(
                target_os="darwin",
                output=binary_path,
                package=str(PROJECT_ROOT),
            ),
            env=build_env,
        )

        contents_dir = bundle_source / "Contents"
        macos_dir = contents_dir / "MacOS"
        resources_dir = contents_dir / "Resources"
        macos_dir.mkdir(parents=True, exist_ok=True)
        resources_dir.mkdir(parents=True, exist_ok=True)
        shutil.copy2(binary_path, macos_dir / APP_EXECUTABLE)
        os.chmod(macos_dir / APP_EXECUTABLE, 0o755)
        write_info_plist(contents_dir / "Info.plist", version, build_number)
        create_icon(
            ICON_SOURCE,
            iconset_dir,
            resources_dir / "icon.icns",
        )

        staging_dir.mkdir(parents=True, exist_ok=True)
        shutil.copytree(bundle_source, staging_dir / f"{APP_NAME}.app")
        command_path = staging_dir / "remove-quarantine.command"
        command_path.write_text(quarantine_script(), encoding="utf-8")
        os.chmod(command_path, 0o755)
        (staging_dir / "Applications").symlink_to("/Applications")

        remove_generated_path(app_bundle)
        remove_generated_path(dmg_path)
        shutil.copytree(bundle_source, app_bundle)

        print(f"Creating {dmg_path}...")
        run_command(
            [
                "hdiutil",
                "create",
                "-volname",
                APP_NAME,
                "-srcfolder",
                str(staging_dir),
                "-ov",
                "-format",
                "UDZO",
                str(dmg_path),
            ],
            quiet=True,
        )

    print("\nBuild completed successfully.")
    print(f"Application: {app_bundle}")
    print(f"Disk image:  {dmg_path}")
    return app_bundle, dmg_path


def main() -> int:
    try:
        build(parse_args())
    except (OSError, RuntimeError, subprocess.CalledProcessError) as error:
        print(f"Error: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
