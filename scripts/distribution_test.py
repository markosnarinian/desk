#!/usr/bin/env python3
"""Exercise the published formula and Go installation interfaces offline."""

import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
import unittest
import zipfile


ROOT = Path(__file__).resolve().parents[1]
MODULE = "github.com/markosnarinian/desk"
VERSION = "v0.1.0"


class DistributionTests(unittest.TestCase):
    def test_formula_uses_each_platforms_actual_archive(self):
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            expected = {}
            for target in ("darwin_arm64", "darwin_amd64", "linux_arm64", "linux_amd64"):
                name = f"desk_{VERSION}_{target}.tar.gz"
                payload = (target + "\n").encode()
                (directory / name).write_bytes(payload)
                expected[name] = hashlib.sha256(payload).hexdigest()
            output = subprocess.check_output(
                ["python3", ROOT / "scripts/homebrew.py", VERSION, directory], text=True
            )
            matches = re.findall(
                r'url "https://github.com/markosnarinian/desk/releases/download/v0\.1\.0/([^"/]+)"\s+sha256 "([0-9a-f]{64})"',
                output,
            )
            self.assertEqual(dict(matches), expected)
            self.assertEqual(len(matches), 4)
            for platform, os_name in (("macos", "darwin"), ("linux", "linux")):
                block = output.split(f"on_{platform} do", 1)[1].split("  end", 1)[0]
                self.assertLess(block.index(f"{os_name}_arm64"), block.index("else"))
                self.assertGreater(block.index(f"{os_name}_amd64"), block.index("else"))
            self.assertIn('license "AGPL-3.0-only"', output)

    def test_formula_rejects_invalid_versions_and_missing_archives(self):
        with tempfile.TemporaryDirectory() as temporary:
            for version in ("../v1.2.3", "v01.2.3", "v1.2", "v1.2.3-beta", VERSION):
                with self.subTest(version=version):
                    result = subprocess.run(
                        ["python3", ROOT / "scripts/homebrew.py", version, temporary],
                        capture_output=True,
                        text=True,
                    )
                    self.assertNotEqual(result.returncode, 0)
                    self.assertEqual(result.stdout, "")

    def test_versioned_go_install_produces_desk_and_sets_up_shell(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            proxy = root / "proxy"
            versions = proxy / MODULE / "@v"
            versions.mkdir(parents=True)
            (versions / "list").write_text(VERSION + "\n")
            (versions / f"{VERSION}.mod").write_bytes((ROOT / "go.mod").read_bytes())
            (versions / f"{VERSION}.info").write_text(
                json.dumps({"Version": VERSION, "Time": "2026-11-03T00:00:00Z"})
            )
            with zipfile.ZipFile(versions / f"{VERSION}.zip", "w") as archive:
                paths = [ROOT / "go.mod", ROOT / "LICENSE", *(ROOT / "cmd/desk").rglob("*")]
                for path in paths:
                    if path.is_file():
                        archive.write(path, f"{MODULE}@{VERSION}/{path.relative_to(ROOT)}")
            home = root / "home"
            home.mkdir()
            environment = dict(os.environ, GOPROXY=proxy.as_uri(), GOSUMDB="off",
                               GOMODCACHE=str(root / "modules"), GOBIN=str(root / "bin"),
                               GOTOOLCHAIN="local")
            result = subprocess.run(
                ["go", "install", f"{MODULE}/cmd/desk@{VERSION}"],
                cwd=root, env=environment, capture_output=True, text=True,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            binary = root / "bin/desk"
            environment.update(HOME=str(home), SHELL="/bin/bash", DESK_HOME=str(root / "data"))
            output = subprocess.check_output([binary, "version"], env=environment, text=True)
            self.assertEqual(output, VERSION + "\n")
            subprocess.run([binary, "init"], env=environment, check=True, capture_output=True)
            self.assertIn("desk shell integration", (home / ".bashrc").read_text())


if __name__ == "__main__":
    unittest.main()
