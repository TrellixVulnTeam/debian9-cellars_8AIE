# gawk: Build a bottle for Linuxbrew
class Gawk < Formula
  desc "GNU awk utility"
  homepage "https://www.gnu.org/software/gawk/"
  url "https://ftpmirror.gnu.org/gawk/gawk-4.1.4.tar.xz"
  mirror "https://ftp.gnu.org/gnu/gawk/gawk-4.1.4.tar.xz"
  sha256 "53e184e2d0f90def9207860531802456322be091c7b48f23fdc79cda65adc266"
  revision 1

  fails_with :llvm do
    build 2326
    cause "Undefined symbols when linking"
  end

  depends_on "mpfr"
  depends_on "readline"

  def install
    system "./configure", "--disable-debug",
                          "--disable-dependency-tracking",
                          "--prefix=#{prefix}",
                          "--without-libsigsegv-prefix"
    system "make"
    if which "cmp"
      system "make", "check"
    else
      opoo "Skipping `make check` due to unavailable `cmp`"
    end
    system "make", "install"
  end

  test do
    output = pipe_output("#{bin}/gawk '{ gsub(/Macro/, \"Home\"); print }' -", "Macrobrew")
    assert_equal "Homebrew", output.strip
  end
end
