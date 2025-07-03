# Homebrew Formula for jot
# This file is a template for creating a Homebrew tap
# To use this:
# 1. Create a repository: johncoder/homebrew-tap
# 2. Copy this to Formula/jot.rb in that repository
# 3. Update the sha256 checksums for each release

class Jot < Formula
  desc "Git-inspired CLI tool for capturing, refiling, and maintaining notes"
  homepage "https://github.com/johncoder/jot"
  version "0.9.0"  # Update this for each release
  
  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/johncoder/jot/releases/download/v#{version}/jot_v#{version}_darwin_arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_ARM64"  # Update this
    else
      url "https://github.com/johncoder/jot/releases/download/v#{version}/jot_v#{version}_darwin_amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_AMD64"  # Update this
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/johncoder/jot/releases/download/v#{version}/jot_v#{version}_linux_arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_LINUX_ARM64"  # Update this
    else
      url "https://github.com/johncoder/jot/releases/download/v#{version}/jot_v#{version}_linux_amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_LINUX_AMD64"  # Update this
    end
  end

  def install
    bin.install "jot_*" => "jot"
  end

  test do
    system "#{bin}/jot", "--version"
    system "#{bin}/jot", "--help"
  end
end
