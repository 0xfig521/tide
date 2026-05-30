class Tide < Formula
  desc "High-speed RSS reader for the terminal"
  homepage "https://github.com/0xfig521/tide"
  license "MIT"
  version "0.1.1"

  BASE = "https://github.com/0xfig521/tide/releases/download/v#{version}"

  on_macos do
    if Hardware::CPU.arm?
      url "#{BASE}/tide-darwin-arm64.tar.gz"
      sha256 ""
    else
      url "#{BASE}/tide-darwin-amd64.tar.gz"
      sha256 ""
    end
  end

  on_linux do
    url "#{BASE}/tide-linux-amd64.tar.gz"
    sha256 ""
  end

  def install
    bin.install "tide"
  end

  test do
    system "#{bin}/tide", "--help"
  end
end
