class Tide < Formula
  desc "High-speed RSS reader for the terminal"
  homepage "https://github.com/0xfig-labs/tide"
  license "MIT"
  version "0.1.1"

  BASE = "https://github.com/0xfig-labs/tide/releases/download/v#{version}"

  on_macos do
    if Hardware::CPU.arm?
      url "#{BASE}/tide-darwin-arm64.tar.gz"
      sha256 "2a99de3409cbe0ee06e5a7e9366903750cdd03b1cf025b8c7ff067829b0483a9"
    else
      url "#{BASE}/tide-darwin-amd64.tar.gz"
      sha256 "0d39a26840e22f250a6f5130eaada3511ee032863817934f74bc636a02a5a23e"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "#{BASE}/tide-linux-arm64.tar.gz"
      sha256 "a823b6ecdd271d9fc2e9be3901fa04b94cc79a706525012bbac12e62ac4ab5b0"
    else
      url "#{BASE}/tide-linux-amd64.tar.gz"
      sha256 "10990d76aab780e96dde8e6299e9c6c2aa8d2035ddd841270f43e07f0625b121"
    end
  end

  def install
    bin.install "tide"
  end

  test do
    system "#{bin}/tide", "--help"
  end
end
