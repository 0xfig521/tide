class Tide < Formula
  desc "High-speed RSS reader for the terminal"
  homepage "https://github.com/0xfig-labs/tide"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/0xfig-labs/tide/releases/latest/download/tide-darwin-arm64.tar.gz"
      sha256 ""  # populated by `brew fetch` or release pipeline
    else
      url "https://github.com/0xfig-labs/tide/releases/latest/download/tide-darwin-amd64.tar.gz"
      sha256 ""
    end
  end

  on_linux do
    url "https://github.com/0xfig-labs/tide/releases/latest/download/tide-linux-amd64.tar.gz"
    sha256 ""
  end

  def install
    bin.install "tide"
  end

  test do
    system "#{bin}/tide", "--help"
  end
end
