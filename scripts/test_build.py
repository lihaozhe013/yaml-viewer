"""Tests for shared production build command generation."""

from __future__ import annotations

import unittest
from pathlib import Path

import build


class ProductionBuildCommandTests(unittest.TestCase):
    def test_windows_build_uses_gui_subsystem(self) -> None:
        command = build.production_build_command(target_os="windows")

        self.assertIn("-H=windowsgui -s -w", command)
        self.assertNotIn("-o", command)

    def test_macos_build_uses_portable_linker_flags(self) -> None:
        command = build.production_build_command(target_os="darwin")

        self.assertIn("-s -w", command)
        self.assertNotIn("-H=windowsgui", " ".join(command))

    def test_packaging_build_accepts_an_explicit_output(self) -> None:
        output = Path("dist") / "yaml-viewer"

        command = build.production_build_command(
            target_os="darwin",
            output=output,
            package="/project",
        )

        self.assertEqual(command[-3:], ["-o", str(output), "/project"])


class BuildEnvironmentTests(unittest.TestCase):
    def test_target_values_override_the_base_environment(self) -> None:
        environment = build.build_environment(
            target_os="windows",
            target_arch="amd64",
            base={"GOOS": "darwin", "CUSTOM": "kept"},
        )

        self.assertEqual(environment["CGO_ENABLED"], "1")
        self.assertEqual(environment["GOOS"], "windows")
        self.assertEqual(environment["GOARCH"], "amd64")
        self.assertEqual(environment["CUSTOM"], "kept")


if __name__ == "__main__":
    unittest.main()
