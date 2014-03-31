# seekaur

seekaur is a simple command-line tool for talking to the [AUR RPC interface](https://aur.archlinux.org/rpc.php). I generally prefer to install AUR packages with `makepkg`, so I didn't need a real [AUR helper](https://wiki.archlinux.org/index.php/AUR_Helpers). I just needed a way to query the AUR and this was the result. seekaur is little more than HTTP GETs, JSON decoding and pretty-printing, so there are a lot of ways to make an equivalent tool.

seekaur does not implement a pacman-style interface like some AUR helpers do, because its feature set is much smaller. It doesn't install packages or even download packages (I prefer to curl the tarball and untar it myself, personally). All it does is retrieve info about packages from the AUR.

## Searching

Search for packages with the `search` command.

```
$ seekaur search chruby
aur/devel/chruby-git v0.3.6.26.ge6e4073-1
    Changes the current Ruby
aur/system/chruby 0.3.8-1
    changes the current ruby. Supports both zsh and bash.
```

Search output is colored and formatted the same way as `pacman -Ss`. It is sorted by categories, then names.

Packages that are out of date will have their versions printed in red instead of green. pacman's output does not have this feature, but I personally find it useful, since it can be a warning sign that a package is no longer being maintained. Plus it's useful information in general.

## Detailed info

To get more information on packages, use the `info` command.

```
$ seekaur info jq yaourt
Category        : editors
Name            : jq
Version         : 1.3-2
Description     : Command-line JSON processor
URL             : http://stedolan.github.io/jq/
Licenses        : custom
Maintainer      : jahiy
First Submitted : Mon 22 Oct 2012 05:02:48 PM EDT
Last Modified   : Wed 23 Oct 2013 04:38:47 AM EDT
Votes           : 31

Category        : system
Name            : yaourt
Version         : 1.3-1
Description     : A pacman wrapper with extended features and AUR support
URL             : http://www.archlinux.fr/yaourt-en/
Licenses        : GPL
Maintainer      : tuxce
First Submitted : Tue 04 Jul 2006 04:37:58 PM EDT
Last Modified   : Fri 05 Apr 2013 05:32:10 PM EDT
Votes           : 2840

```

Info output is formatted the same way as `pacman -Si`. As with `pacman -Si`, the info is listed in the same order as the arguments were given, and nonexistent packages will be printed as an error. Packages that are out of date will have an `[out of date]` indicator appended to their version.

## tarball links and PKGBUILDs

You can retrieve the link for an AUR package's tarball with the `tarball` command, or display its PKGBUILD on stdout with the `pkgbuild` command. Both commands have the same syntax and behavior as the `info` command, so you can invoke them on multiple packages at once (although in practice, you probably only want to see one PKGBUILD at a time).

As I mentioned earlier, I generally like to install packages with `makepkg`, so seekaur lets me do something like `curl -L $(seekaur tarball libgit2-git) | tar -zx` to fetch the tarball quickly from the command line.

AUR tarball/PKGBUILD links follow a fairly predictable structure, so these commands do not perform an HTTP request to query the info of the packages first. Therefore, if you attempt to invoke these commands on a package that does not exist, no error will be raised...

```
$ seekaur tarball foobarbazqux    # this package does not exist
https://aur.archlinux.org/packages/fo/foobarbazqux/foobarbazqux.tar.gz
$ seekaur pkgbuild foobarbazqux
<html>
<head><title>404 Not Found</title></head>
<body bgcolor="white">
<center><h1>404 Not Found</h1></center>
<hr><center>nginx/1.4.5</center>
</body>
</html>

```

I'll probably improve the 404 behavior in the future for the pkgbuild command, with a nice error message.

## License

MIT/expat, see [LICENSE.md](LICENSE.md).
