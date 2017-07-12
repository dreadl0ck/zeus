class Zeus < Formula
  desc "An Electrifying Build System"
  homepage "https://github.com/dreadl0ck/zeus"
  url "https://github.com/dreadl0ck/zeus/releases/download/v0.8/zeus_0.8_darwin_amd64.tar.gz"
  version "0.8"
  sha256 "3728eecd858d69cbfba26ea3cdb0143ad6f1d58f63a7c8065e8af486b06cfc68"

  def install
    bin.install "zeus"
  end
end
