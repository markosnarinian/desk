#!/usr/bin/env python3
"""Generate the tap formula from the exact release archives being published."""

import hashlib
from pathlib import Path
import re
import sys


def formula(version, directory):
    if not re.fullmatch(r"v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)", version):
        raise ValueError("expected a stable version such as v0.1.0")
    blocks = []
    for os in ("darwin", "linux"):
        platforms = []
        for arch in ("arm64", "amd64"):
            name = f"desk_{version}_{os}_{arch}.tar.gz"
            digest = hashlib.sha256((directory / name).read_bytes()).hexdigest()
            platforms.append(
                f'      url "https://github.com/markosnarinian/desk/releases/download/{version}/{name}"\n'
                f'      sha256 "{digest}"'
            )
        platform = "macos" if os == "darwin" else "linux"
        blocks.append(
            f"  on_{platform} do\n"
            f"    if Hardware::CPU.arm?\n{platforms[0]}\n"
            f"    else\n{platforms[1]}\n    end\n  end"
        )
    return f'''class Desk < Formula
  desc "Workspaces for temporary local work"
  homepage "https://github.com/markosnarinian/desk"
  version "{version[1:]}"
  license "AGPL-3.0-only"

{chr(10).join(blocks)}

  def install
    bin.install "desk"
    doc.install "LICENSE"
  end

  def caveats
    "Run desk init, then open a new terminal to enable workspace switching."
  end

  test do
    assert_equal "v#{{version}}", shell_output("#{{bin}}/desk version").strip
  end
end
'''


if __name__ == "__main__":
    try:
        if len(sys.argv) != 3:
            raise ValueError("usage: scripts/homebrew.py VERSION ARCHIVE_DIRECTORY")
        print(formula(sys.argv[1], Path(sys.argv[2])), end="")
    except (OSError, ValueError) as error:
        sys.exit(str(error))
