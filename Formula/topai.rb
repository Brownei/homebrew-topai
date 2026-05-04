# Formula/topai.rb
class Topai < Formula
  desc "AI-powered process monitor - identifies stuck vs busy processes"
  homepage "https://github.com/Brownei/topia"
  url "https://github.com/Brownei/topai/releases/download/v1.0.0/topai-macos-arm64"
  sha256 "YOUR_SHA256_HERE"
  version "1.0.0"

  depends_on "go" => :build

  def install
    bin.install "topai-macos-arm64" => "topai"
  end

  def post_install
    puts "To use topai, you need an Anthropic API key:"
    puts "Get one at: https://console.anthropic.com/"
    puts "topai will prompt you on first run!"
  end

  test do
    system "#{bin}/topai", "--version"
  end
end
