class PkgConfig < Formula
  desc "Manage compile and link flags for libraries"
  homepage "https://freedesktop.org/wiki/Software/pkg-config/"
  url "https://pkgconfig.freedesktop.org/releases/pkg-config-0.29.2.tar.gz"
  mirror "https://dl.bintray.com/homebrew/mirror/pkg-config-0.29.2.tar.gz"
  sha256 "6fc69c01688c9458a57eb9a1664c9aba372ccda420a02bf4429fe610e7e7d591"

  def install
    pc_path = %W[
      #{HOMEBREW_PREFIX}/lib/pkgconfig
      #{HOMEBREW_PREFIX}/share/pkgconfig
    ].join(File::PATH_SEPARATOR)
    if OS.mac?
      pc_path += %W[/usr/local/lib/pkgconfig /usr/lib/pkgconfig #{HOMEBREW_LIBRARY}/Homebrew/os/mac/pkgconfig/#{MacOS.version}"]
    else
      pc_path << "#{HOMEBREW_LIBRARY}/Homebrew/os/linux/pkgconfig"
    end

    system "./configure", "--disable-debug",
                          "--prefix=#{prefix}",
                          "--disable-host-tool",
                          "--with-internal-glib",
                          "--with-pc-path=#{pc_path}"
    system "make"
    system "make", "check"
    system "make", "install"
  end

  test do
    system "#{bin}/pkg-config", "--libs", "libpcre"
  end
end
