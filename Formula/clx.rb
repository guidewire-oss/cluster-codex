class ClusterCodex < Formula
  desc "Generate Kubernetes Bill of Materials for a Kubernetes cluster"
  homepage "https://github.com/guidewire-oss/cluster-codex"
  license "Apache 2"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/guidewire-oss/cluster-codex/releases/download/v0.0.1/clx_0.0.1_darwin_arm64.tar.gz"
      sha256 "5a1f9f94ccc08c6e1efb9d616eeb69022f1316029b9c6c774254010f0bd2f3b6"
    elsif Hardware::CPU.intel?
      url "https://github.com/guidewire-oss/cluster-codex/releases/download/v0.0.1/clx_0.0.1_darwin_amd64.tar.gz .tar.gz"
      sha256 "5674ebc97c8113a72bc36a8d5fc849791cfe4f6df7d0a7d409c998fa4065bd52"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/guidewire-oss/cluster-codex/releases/download/v0.0.1/clx_0.0.1_linux_arm64.tar.gz"
      sha256 "5674ebc97c8113a72bc36a8d5fc849791cfe4f6df7d0a7d409c998fa4065bd52"
    elsif Hardware::CPU.intel?
      url "https://github.com/guidewire-oss/cluster-codex/releases/download/v0.0.1/clx_0.0.1_linux_amd64.tar.gz"
      sha256 "5674ebc97c8113a72bc36a8d5fc849791cfe4f6df7d0a7d409c998fa4065bd52"
    end
  end

  def install
    bin.install "clx"
  end

  test do
    system "#{bin}/clx", "--version"
  end
end
