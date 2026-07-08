# PR / setup: Homebrew tap `kabzhanov/homebrew-tap`

## Goal
Let macOS users install AISentinel with:
```bash
brew install kabzhanov/tap/aisentinel
```

## Files to create in NEW repo `kabzhanov/homebrew-tap`

### `Formula/aisentinel.rb`
```ruby
class Aisentinel < Formula
  desc "Security, control, and observability layer for AI agents (MCP server + sidecar)"
  homepage "https://github.com/Kabzhanov/AISentinel"
  url "https://github.com/Kabzhanov/AISentinel/archive/refs/tags/v1.0.3.tar.gz"
  sha256 "COMPUTE_SHA256_OF_v1.0.3_TARBALL"   # see "computing the SHA" below
  license "Apache-2.0"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "-o", bin/"aisentinel", "./cmd/aisentinel"
    system "go", "build", *std_go_args(ldflags: "-s -w"), "-o", bin/"aisentinel-sidecar", "./cmd/aisentinel-sidecar"
  end

  test do
    assert_match "aisentinel", shell_output("#{bin}/aisentinel version")
  end
end
```

## Steps
1. Create the new repo on GitHub: `kabzhanov/homebrew-tap` (Public).
2. Add an empty initial commit on `main`.
3. In the new repo, add `Formula/aisentinel.rb` with the content above.
4. Update `sha256` for each release:
   ```bash
   curl -sL https://github.com/Kabzhanov/AISentinel/archive/refs/tags/v1.0.3.tar.gz | shasum -a 256
   ```
5. Bump `version` and `url` when cutting a new release.
6. (Optional) Tag the tap repo for each formula version (`v1.0.3`).

## Compute SHA256 for new releases
```bash
RELEASE=v1.0.4
curl -sL https://github.com/Kabzhanov/AISentinel/archive/refs/tags/${RELEASE}.tar.gz | shasum -a 256
# paste result into the Formula's sha256 line
```

## Verify after install
```bash
brew install kabzhanov/tap/aisentinel
aisentinel --version
aisentinel-sidecar --help
```
