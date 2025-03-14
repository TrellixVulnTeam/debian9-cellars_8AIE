class Python < Formula
  desc "Interpreted, interactive, object-oriented programming language"
  homepage "https://www.python.org/"
  url "https://www.python.org/ftp/python/2.7.13/Python-2.7.13.tar.xz"
  sha256 "35d543986882f78261f97787fd3e06274bfa6df29fac9b4a94f73930ff98f731"
  revision 1
  head "https://github.com/python/cpython.git", :branch => "2.7"

  # Homebrew doesn't accept a wide/ucs4 option because narrow build is the de facto standard
  # on Windows and OSX, but wide seems to be the default for linux
  # More details in: https://github.com/Homebrew/homebrew/pull/32368
  option "with-unicode-ucs4", "Build unicode support with UCS4"
  option "with-quicktest", "Run `make quicktest` after the build (for devs; may fail)"
  option "with-tcl-tk", "Use Homebrew's Tk instead of macOS Tk (has optional Cocoa and threads support)"
  if OS.mac?
    option "with-poll", "Enable select.poll, which is not fully implemented on macOS (https://bugs.python.org/issue5154)"
  else
    option "without-poll", "Disable select.poll"
  end

  # sphinx-doc depends on python, but on 10.6 or earlier python is fulfilled by
  # brew, which would lead to circular dependency.
  if MacOS.version > :snow_leopard
    option "with-sphinx-doc", "Build HTML documentation"
    depends_on "sphinx-doc" => [:build, :optional]
  end

  deprecated_option "quicktest" => "with-quicktest"
  deprecated_option "with-brewed-tk" => "with-tcl-tk"

  depends_on "pkg-config" => :build
  depends_on "readline" => :recommended
  depends_on "sqlite" => :recommended
  depends_on "gdbm" => :recommended
  depends_on "openssl"
  depends_on "tcl-tk" => :optional
  if OS.mac?
    depends_on "berkeley-db@4" => :optional
  else
    depends_on :x11 if build.with? "tcl-tk"
    depends_on "berkeley-db@4" => :recommended
    depends_on "bzip2"
    depends_on "ncurses"
    depends_on "zlib"
  end

  resource "setuptools" do
    url "https://files.pythonhosted.org/packages/26/d1/dc7fe14ce4a3ff3faebf1ac11350de4104ea2d2a80c98393b55c84362b0c/setuptools-32.1.0.tar.gz"
    sha256 "86d57bf86edc0ecfd2dc0907ed3710bc4501fb13a06c0fcaf7632305b00ce832"
  end

  resource "pip" do
    url "https://files.pythonhosted.org/packages/11/b6/abcb525026a4be042b486df43905d6893fb04f05aac21c32c638e939e447/pip-9.0.1.tar.gz"
    sha256 "09f243e1a7b461f654c26a725fa373211bb7ff17a9300058b205c61658ca940d"
  end

  resource "wheel" do
    url "https://files.pythonhosted.org/packages/c9/1d/bd19e691fd4cfe908c76c429fe6e4436c9e83583c4414b54f6c85471954a/wheel-0.29.0.tar.gz"
    sha256 "1ebb8ad7e26b448e9caa4773d2357849bf80ff9e313964bcaf79cbf0201a1648"
  end

  # Patch to disable the search for Tk.framework, since Homebrew's Tk is
  # a plain unix build. Remove `-lX11`, too because our Tk is "AquaTk".
  if build.with? "tcl-tk"
    patch do
      url "https://raw.githubusercontent.com/Homebrew/patches/42fcf22/python/brewed-tk-patch.diff"
      sha256 "15c153bdfe51a98efe48f8e8379f5d9b5c6c4015e53d3f9364d23c8689857f09"
    end
  end

  def lib_cellar
    prefix / (OS.mac? ? "Frameworks/Python.framework/Versions/2.7" : "") /
      "lib/python2.7"
  end

  def site_packages_cellar
    lib_cellar/"site-packages"
  end

  # The HOMEBREW_PREFIX location of site-packages.
  def site_packages
    HOMEBREW_PREFIX/"lib/python2.7/site-packages"
  end

  # setuptools remembers the build flags python is built with and uses them to
  # build packages later. Xcode-only systems need different flags.
  pour_bottle? do
    reason <<-EOS.undent
    The bottle needs the Apple Command Line Tools to be installed.
      You can install them, if desired, with:
        xcode-select --install
    EOS
    satisfy { MacOS::CLT.installed? }
  end

  def install
    ENV.permit_weak_imports

    if OS.mac? && build.with?("poll")
      opoo "The given option --with-poll enables a somewhat broken poll() on macOS (https://bugs.python.org/issue5154)."
    end

    # Unset these so that installing pip and setuptools puts them where we want
    # and not into some other Python the user has installed.
    ENV["PYTHONHOME"] = nil
    ENV["PYTHONPATH"] = nil

    args = %W[
      --prefix=#{prefix}
      --enable-ipv6
      --datarootdir=#{share}
      --datadir=#{share}
      #{OS.mac? ? "--enable-framework=#{frameworks}" : "--enable-shared"}
      --without-ensurepip
    ]

    args << "--without-gcc" if ENV.compiler == :clang
    args << "--enable-unicode=ucs4" if build.with? "unicode-ucs4"

    cflags   = []
    ldflags  = []
    cppflags = []

    unless MacOS::CLT.installed?
      # Help Python's build system (setuptools/pip) to build things on Xcode-only systems
      # The setup.py looks at "-isysroot" to get the sysroot (and not at --sysroot)
      cflags   << "-isysroot #{MacOS.sdk_path}"
      ldflags  << "-isysroot #{MacOS.sdk_path}"
      cppflags << "-I#{MacOS.sdk_path}/usr/include" # find zlib
      # For the Xlib.h, Python needs this header dir with the system Tk
      if build.without? "tcl-tk"
        cflags << "-I#{MacOS.sdk_path}/System/Library/Frameworks/Tk.framework/Versions/8.5/Headers"
      end
    end

    # Python's setup.py parses CPPFLAGS and LDFLAGS to learn search
    # paths for the dependencies of the compiled extension modules.
    # See Homebrew/linuxbrew#420, Homebrew/linuxbrew#460, and Homebrew/linuxbrew#875
    unless OS.mac?
      if build.bottle?
        # Configure Python to use cc and c++ to build extension modules.
        ENV["CC"] = "cc"
        ENV["CXX"] = "c++"
      end
      cppflags << ENV.cppflags << " -I#{HOMEBREW_PREFIX}/include"
      ldflags << ENV.ldflags << " -L#{HOMEBREW_PREFIX}/lib"
    end

    # Avoid linking to libgcc https://code.activestate.com/lists/python-dev/112195/
    args << "MACOSX_DEPLOYMENT_TARGET=#{MacOS.version}"

    # We want our readline and openssl! This is just to outsmart the detection code,
    # superenv handles that cc finds includes/libs!
    inreplace "setup.py" do |s|
      s.gsub! "do_readline = self.compiler.find_library_file(lib_dirs, 'readline')",
              "do_readline = '#{Formula["readline"].opt_lib}/libhistory.dylib'"
      s.gsub! "/usr/local/ssl", Formula["openssl"].opt_prefix
      s.gsub! "/usr/include/db4", Formula["berkeley-db@4"].opt_include
    end

    if build.with? "sqlite"
      inreplace "setup.py" do |s|
        s.gsub! "sqlite_setup_debug = False", "sqlite_setup_debug = True"
        s.gsub! "for d_ in inc_dirs + sqlite_inc_paths:",
                "for d_ in ['#{Formula["sqlite"].opt_include}']:"

        # Allow sqlite3 module to load extensions:
        # https://docs.python.org/library/sqlite3.html#f1
        s.gsub! 'sqlite_defines.append(("SQLITE_OMIT_LOAD_EXTENSION", "1"))', ""
      end
    end

    # Allow python modules to use ctypes.find_library to find homebrew's stuff
    # even if homebrew is not a /usr/local/lib. Try this with:
    # `brew install enchant && pip2 install pyenchant`
    inreplace "./Lib/ctypes/macholib/dyld.py" do |f|
      f.gsub! "DEFAULT_LIBRARY_FALLBACK = [", "DEFAULT_LIBRARY_FALLBACK = [ '#{HOMEBREW_PREFIX}/lib',"
      f.gsub! "DEFAULT_FRAMEWORK_FALLBACK = [", "DEFAULT_FRAMEWORK_FALLBACK = [ '#{HOMEBREW_PREFIX}/Frameworks',"
    end

    if build.with? "tcl-tk"
      tcl_tk = Formula["tcl-tk"].opt_prefix
      cppflags << "-I#{tcl_tk}/include"
      ldflags  << "-L#{tcl_tk}/lib"
    end

    args << "CFLAGS=#{cflags.join(" ")}" unless cflags.empty?
    args << "LDFLAGS=#{ldflags.join(" ")}" unless ldflags.empty?
    args << "CPPFLAGS=#{cppflags.join(" ")}" unless cppflags.empty?

    system "./configure", *args

    # HAVE_POLL is "broken" on macOS. See:
    # https://trac.macports.org/ticket/18376
    # https://bugs.python.org/issue5154
    if build.without? "poll"
      inreplace "pyconfig.h", /.*?(HAVE_POLL[_A-Z]*).*/, '#undef \1'
    end

    system "make"
    if build.with?("quicktest") || build.bottle?
      system "make", "quicktest", "TESTPYTHONOPTS=-s", "TESTOPTS=-j#{ENV.make_jobs} -w"
    end

    ENV.deparallelize do
      # Tell Python not to install into /Applications
      system "make", "install", "PYTHONAPPSDIR=#{prefix}"
      system "make", "frameworkinstallextras", "PYTHONAPPSDIR=#{pkgshare}" if OS.mac?
    end

    # Fixes setting Python build flags for certain software
    # See: https://github.com/Homebrew/homebrew/pull/20182
    # https://bugs.python.org/issue3588
    inreplace lib_cellar/"config/Makefile" do |s|
      s.change_make_var! "LINKFORSHARED",
        "-u _PyMac_Error $(PYTHONFRAMEWORKINSTALLDIR)/Versions/$(VERSION)/$(PYTHONFRAMEWORK)"
    end if OS.mac?

    if OS.mac?
      # Prevent third-party packages from building against fragile Cellar paths
      inreplace [lib_cellar/"_sysconfigdata.py",
                 lib_cellar/"config/Makefile",
                 frameworks/"Python.framework/Versions/Current/lib/pkgconfig/python-2.7.pc"],
                 prefix, opt_prefix

      # Symlink the pkgconfig files into HOMEBREW_PREFIX so they're accessible.
      (lib/"pkgconfig").install_symlink Dir[frameworks/"Python.framework/Versions/Current/lib/pkgconfig/*"]
    end
    # Remove the site-packages that Python created in its Cellar.
    site_packages_cellar.rmtree

    (libexec/"setuptools").install resource("setuptools")
    (libexec/"pip").install resource("pip")
    (libexec/"wheel").install resource("wheel")

    if MacOS.version > :snow_leopard && build.with?("sphinx-doc")
      cd "Doc" do
        system "make", "html"
        doc.install Dir["build/html/*"]
      end
    end

    # Remove commands shadowing system python.
    {
      "2to3" => "2to3-2",
      "easy_install" => "easy_install-2.7",
      "idle" => "idle2",
      "pip" => "pip2",
      "pydoc" => "pydoc2",
      "python" => "python2",
      "python-config" => "python2-config",
      "pythonw" => "pythonw2",
      "smtpd.py" => "smtpd2.py",
      "wheel" => nil,
    }.each do |unversioned_name, versioned_name|
      rm_f bin/unversioned_name
      next unless versioned_name
      (libexec/"bin").install_symlink bin/versioned_name => unversioned_name
    end
  end

  def post_install
    # Fix up the site-packages so that user-installed Python software survives
    # minor updates, such as going from 2.7.0 to 2.7.1:

    # Create a site-packages in HOMEBREW_PREFIX/lib/python2.7/site-packages
    site_packages.mkpath

    # Symlink the prefix site-packages into the cellar.
    site_packages_cellar.unlink if site_packages_cellar.exist?
    site_packages_cellar.parent.install_symlink site_packages

    # Write our sitecustomize.py
    rm_rf Dir["#{site_packages}/sitecustomize.py[co]"]
    (site_packages/"sitecustomize.py").atomic_write(sitecustomize) if OS.mac?

    # Remove old setuptools installations that may still fly around and be
    # listed in the easy_install.pth. This can break setuptools build with
    # zipimport.ZipImportError: bad local file header
    # setuptools-0.9.5-py3.3.egg
    rm_rf Dir["#{site_packages}/setuptools*"]
    rm_rf Dir["#{site_packages}/distribute*"]
    rm_rf Dir["#{site_packages}/pip[-_.][0-9]*", "#{site_packages}/pip"]

    setup_args = ["-s", "setup.py", "--no-user-cfg", "install", "--force",
                  "--verbose",
                  "--single-version-externally-managed",
                  "--record=installed.txt",
                  "--install-scripts=#{bin}",
                  "--install-lib=#{site_packages}"]

    (libexec/"setuptools").cd { system "#{bin}/python2", *setup_args }
    (libexec/"pip").cd { system "#{bin}/python2", *setup_args }
    (libexec/"wheel").cd { system "#{bin}/python2", *setup_args }

    # When building from source, these symlinks will not exist, since
    # post_install happens after linking.
    %w[pip2 pip2.7 easy_install-2.7].each do |e|
      (HOMEBREW_PREFIX/"bin").install_symlink bin/e
    end

    # Help distutils find brewed stuff when building extensions
    include_dirs = [HOMEBREW_PREFIX/"include", Formula["openssl"].opt_include]
    library_dirs = [HOMEBREW_PREFIX/"lib", Formula["openssl"].opt_lib]

    if build.with? "sqlite"
      include_dirs << Formula["sqlite"].opt_include
      library_dirs << Formula["sqlite"].opt_lib
    end

    if build.with? "tcl-tk"
      include_dirs << Formula["tcl-tk"].opt_include
      library_dirs << Formula["tcl-tk"].opt_lib
    end

    cfg = lib_cellar/"distutils/distutils.cfg"
    cfg.atomic_write <<-EOF.undent
      [install]
      prefix=#{HOMEBREW_PREFIX}

      [build_ext]
      include_dirs=#{include_dirs.join ":"}
      library_dirs=#{library_dirs.join ":"}
    EOF
  end

  def sitecustomize
    <<-EOF.undent
      # This file is created by Homebrew and is executed on each python startup.
      # Don't print from here, or else python command line scripts may fail!
      # <https://docs.brew.sh/Homebrew-and-Python.html>
      import re
      import os
      import sys

      if sys.version_info[0] != 2:
          # This can only happen if the user has set the PYTHONPATH for 3.x and run Python 2.x or vice versa.
          # Every Python looks at the PYTHONPATH variable and we can't fix it here in sitecustomize.py,
          # because the PYTHONPATH is evaluated after the sitecustomize.py. Many modules (e.g. PyQt4) are
          # built only for a specific version of Python and will fail with cryptic error messages.
          # In the end this means: Don't set the PYTHONPATH permanently if you use different Python versions.
          exit('Your PYTHONPATH points to a site-packages dir for Python 2.x but you are running Python ' +
               str(sys.version_info[0]) + '.x!\\n     PYTHONPATH is currently: "' + str(os.environ['PYTHONPATH']) + '"\\n' +
               '     You should `unset PYTHONPATH` to fix this.')

      # Only do this for a brewed python:
      if os.path.realpath(sys.executable).startswith('#{rack}'):
          # Shuffle /Library site-packages to the end of sys.path and reject
          # paths in /System pre-emptively (#14712)
          library_site = '/Library/Python/2.7/site-packages'
          library_packages = [p for p in sys.path if p.startswith(library_site)]
          sys.path = [p for p in sys.path if not p.startswith(library_site) and
                                             not p.startswith('/System')]
          # .pth files have already been processed so don't use addsitedir
          sys.path.extend(library_packages)

          # the Cellar site-packages is a symlink to the HOMEBREW_PREFIX
          # site_packages; prefer the shorter paths
          long_prefix = re.compile(r'#{rack}/[0-9\._abrc]+/Frameworks/Python\.framework/Versions/2\.7/lib/python2\.7/site-packages')
          sys.path = [long_prefix.sub('#{site_packages}', p) for p in sys.path]

          # LINKFORSHARED (and python-config --ldflags) return the
          # full path to the lib (yes, "Python" is actually the lib, not a
          # dir) so that third-party software does not need to add the
          # -F/#{HOMEBREW_PREFIX}/Frameworks switch.
          try:
              from _sysconfigdata import build_time_vars
              build_time_vars['LINKFORSHARED'] = '-u _PyMac_Error #{opt_prefix}/Frameworks/Python.framework/Versions/2.7/Python'
          except:
              pass  # remember: don't print here. Better to fail silently.

          # Set the sys.executable to use the opt_prefix
          sys.executable = '#{opt_bin}/python2.7'
    EOF
  end

  def caveats; <<-EOS.undent
    This formula installs a python2 executable to #{HOMEBREW_PREFIX}/bin.
    If you wish to have this formula's python executable in your PATH then add
    the following to #{shell_profile}:
      export PATH="#{opt_libexec}/bin:$PATH"

    Pip and setuptools have been installed. To update them
      pip2 install --upgrade pip setuptools

    You can install Python packages with
      pip2 install <package>

    They will install into the site-package directory
      #{site_packages}

    See: https://docs.brew.sh/Homebrew-and-Python.html
    EOS
  end

  test do
    # Check if sqlite is ok, because we build with --enable-loadable-sqlite-extensions
    # and it can occur that building sqlite silently fails if OSX's sqlite is used.
    system "#{bin}/python2", "-c", "import sqlite3"
    # Check if some other modules import. Then the linked libs are working.
    if OS.mac? || build.with?("tcl-tk") && Tab.for_name("tcl-tk").with?("x11")
      system "#{bin}/python2", "-c", "import Tkinter; root = Tkinter.Tk()"
    end
    system bin/"pip2", "list"
  end
end
