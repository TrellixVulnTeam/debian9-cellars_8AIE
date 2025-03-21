class Fontconfig < Formula
  desc "XML-based font configuration API for X Windows"
  homepage "https://wiki.freedesktop.org/www/Software/fontconfig/"
  url "https://www.freedesktop.org/software/fontconfig/release/fontconfig-2.12.4.tar.bz2"
  sha256 "668293fcc4b3c59765cdee5cee05941091c0879edcc24dfec5455ef83912e45c"
  revision 1 unless OS.mac?

  # The bottle tooling is too lenient and thinks fontconfig
  # is relocatable, but it has hardcoded paths in the executables.
  pour_bottle? do
    default_prefix = BottleSpecification::DEFAULT_PREFIX
    reason "The bottle needs to be installed into #{default_prefix}."
    # c.f. the identical hack in lua
    # https://github.com/Homebrew/homebrew/issues/47173
    satisfy { HOMEBREW_PREFIX.to_s == default_prefix }
  end

  head do
    url "https://anongit.freedesktop.org/git/fontconfig", :using => :git

    depends_on "autoconf" => :build
    depends_on "automake" => :build
    depends_on "libtool" => :build
  end

  keg_only :provided_pre_mountain_lion

  option "without-docs", "Skip building the fontconfig docs"

  depends_on "pkg-config" => :build
  depends_on "freetype"
  unless OS.mac?
    depends_on "bzip2"
    depends_on "expat"
    depends_on "autoconf" => :build
    depends_on "automake" => :build
    depends_on "gperf" => :build
    depends_on "libtool" => :build
  end

  def install
    font_dirs = %w[
      /System/Library/Fonts
      /Library/Fonts
      ~/Library/Fonts
    ]

    if MacOS.version == :sierra
      font_dirs << "/System/Library/Assets/com_apple_MobileAsset_Font3"
    elsif MacOS.version == :high_sierra
      font_dirs << "/System/Library/Assets/com_apple_MobileAsset_Font4"
    end

    system "autoreconf", "-iv" if build.head? || !OS.mac?
    system "./configure", "--disable-dependency-tracking",
                          "--disable-silent-rules",
                          "--enable-static",
                          "--with-add-fonts=#{font_dirs.join(",")}",
                          "--prefix=#{prefix}",
                          "--localstatedir=#{var}",
                          "--sysconfdir=#{etc}",
                          ("--disable-docs" if build.without? "docs")
    system "make", "install", "RUN_FC_CACHE_TEST=false"
  end

  def post_install
    ohai "Regenerating font cache, this may take a while"
    system "#{bin}/fc-cache", "-frv"
  end

  test do
    system "#{bin}/fc-list"
  end
end
