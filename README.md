<div align="center">

# mdx 📚

mdx is a simple CLI application for downloading manga from the [MangaDex website](https://mangadex.org/). The program uses [MangaDex API](https://api.mangadex.org/docs/) to fetch manga content.

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/arimatakao/mdx)
![GitHub Release](https://img.shields.io/github/v/release/arimatakao/mdx)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/arimatakao/mdx/total)
![GitHub Repo stars](https://img.shields.io/github/stars/arimatakao/mdx)

![demo](./.github/assets/demo.gif)

</div>

## Features 💫

- Works on ***Windows, MacOS, Linux, Android***.
- Saves manga in ***CBZ, PDF, EPUB formats*** or as a simple ***folder of images***.
- No dependencies required: distributed as a ***single small executable file*** with easy installation scripts for ***Windows, macOS, and Linux***.
- Works without a ***MangaDex account***.
- ***Interactive downloading mode*** for convenient use.
- Downloads ***multiple chapters with a single command***.
- Saves multiple chapters in ***one file***.
- Supports downloading chapters by ***language and translation group***.
- Automatically generates metadata for downloaded files, ***adapted for e-readers***.
- Searches manga.
- Displays information about manga.

## Installation ⚙️

**Download the latest release [HERE](https://github.com/arimatakao/mdx/releases)**

### Linux

Quick install (adds `mdx` to `~/.local/bin`, no sudo; if needed, also appends it to your shell `PATH`):

```sh
curl -fsSL https://raw.githubusercontent.com/arimatakao/mdx/main/install.sh | bash
```

Install via Linux package manager (supported: `apt`, `dnf`, `yum`, `apk`, `pacman`):

```sh
curl -fsSL https://raw.githubusercontent.com/arimatakao/mdx/main/install.sh | sudo bash -s -- --pkg
```

Manual way: first download the package file from the [Releases page](https://github.com/arimatakao/mdx/releases), then run the install command for your distro.

Debian/Ubuntu (`mdx_*_linux_*.deb`):

```sh
sudo apt install ./mdx_*_linux_*.deb
```

RHEL/Fedora (`mdx_*_linux_*.rpm`):

```sh
sudo dnf install ./mdx_*_linux_*.rpm
```

Alpine (`mdx_*_linux_*.apk`):

```sh
sudo apk add --allow-untrusted ./mdx_*_linux_*.apk
```

Arch Linux (`mdx_*_linux_*.pkg.tar.zst`):

```sh
sudo pacman -U ./mdx_*_linux_*.pkg.tar.zst
```

### MacOS

Install with the script (adds `mdx` to `~/.local/bin`, no sudo; if needed, also appends it to your shell `PATH`):

```sh
curl -fsSL https://raw.githubusercontent.com/arimatakao/mdx/main/install.sh | bash
```

Or download the macOS archive (`mdx_*_darwin_*.tar.gz`) and run:

```sh
tar -xzf mdx_*_darwin_*.tar.gz
./mdx --help
```

### Windows

Install automatically with PowerShell script:

```powershell
powershell -ExecutionPolicy Bypass -Command "iwr -useb https://raw.githubusercontent.com/arimatakao/mdx/main/install.ps1 | iex"
```

Or download `mdx-*-windows-installer.msi` and run this command in `cmd`:

```bat
msiexec /i "C:\path\to\mdx-...-windows-installer.msi"
```

Note: **Administrator permission may be required.**

If installation or running `mdx.exe` fails, see [Why can't I install or run mdx on Windows?](#why-cant-i-install-or-run-mdx-on-windows).

### Portable binaries

Windows (`mdx_*_windows_*.zip`, contains `mdx.exe`):
You can just extract the archive with File Explorer/7-Zip, or use this PowerShell command:
```powersh
Expand-Archive .\mdx_*_windows_*.zip -DestinationPath .\mdx
.\mdx\mdx.exe --help
```

Linux/macOS (`mdx_*_linux_*.tar.gz`, contains `mdx`):
```sh
tar -xzf mdx_*_linux_*.tar.gz
./mdx --help
```

### Android (Termux)

1. Install `curl` package:

```sh
pkg install curl -y
```

2. Execute this command:

```sh
bash <(curl -s https://raw.githubusercontent.com/arimatakao/mdx/main/android_installation.sh)
```

You can also install `mdx` manually by running these commands in sequence:

```sh
pkg update && pkg upgrade -y
```

```sh
pkg install -y golang
```

```sh
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
```

```sh
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
```

```sh
source ~/.bashrc
```

```sh
go install github.com/arimatakao/mdx@latest
```

For use `mdx` just execute:

```sh
mdx
```

### Go install

```sh
go install github.com/arimatakao/mdx@latest
```

### Nix/NixOS

Using flakes to run `mdx` directly from the default branch:

```sh
nix run github:arimatakao/mdx -- download --help
```

Using flakes to create a temporary shell with `mdx` available on the `$PATH`:

```sh
nix shell github:arimatakao/mdx
```

Using a pinned tag for reproducible installs:

```sh
nix run 'git+https://github.com/arimatakao/mdx?ref=refs/tags/v1.15.1' -- download --help
```

### Docker

1. Clone the repository:

```sh
git clone https://github.com/arimatakao/mdx.git
```

2. Build docker image:

```sh
docker build -t mdx .
```

Usage examples:

```sh
# Ping
docker run --rm mdx dl ping
# Download
docker run --rm -v /your/download/dir:/download mdx dl -o /download <url>
# Interactive download
docker run --rm -it -v /your/download/dir:/download mdx dl -o /download <url>
```

Also add useful alias for your sh:

```
alias containermdx="docker run --rm -it -v /your/download/dir:/download mdx"
```

It allows you to run mdx anywhere in your sh using the command `containermdx`

## Usage examples️ 🖥️

Interactive downloading mode:

```sh
mdx dl -i
```

Demo of interactive mode:

<div align="center">

![demo](./.github/assets/interactive_mode_demo.gif)

*Note: Your manga title should be more than 5 characters when searching to avoid errors.*

</div>


Download manga:

```sh
# get help
mdx download --help

# by default 1 chapter is being downloaded
mdx download -u https://mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370/this-gorilla-will-die-in-1-day
# or
mdx dl -u https://mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370/this-gorilla-will-die-in-1-day
# or
mdx dl https://mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370
# or
mdx dl mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370

# download pdf format instead of cbz
mdx dl -e pdf mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370
# or epub format
mdx dl -e epub mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370

# download all chapters
# i don't recommend using this flag - https://github.com/arimatakao/mdx?tab=readme-ov-file#getting-error-while-getting-manga-chapters-request-is-failed-i-cant-download-anything-why
mdx dl -a mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370

# download a specific chapter
mdx dl -c 3 mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370
# or set direct link to the chapter
mdx dl --this mangadex.org/chapter/7c5d2aea-ea55-47d9-8c65-a33c9e92df70
# or
mdx dl https://mangadex.org/chapter/7c5d2aea-ea55-47d9-8c65-a33c9e92df70

# download a range of chapters
mdx dl -c 1-3 mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370

# download 1 volume of manga
mdx dl -v 1 mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370

# download a range of chapters and merge them in one file
mdx dl -m -c 1-3 mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370

# download 1 volume of manga and merge chapters in one file
mdx dl -m -v 1 mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370

# download last chapter
mdx dl --last mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370

# specify language, default is english (to get the available languages, execute the info subcommand)
mdx dl -l it mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370

# specify the output directory
mdx dl -o your/dir mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370

# specify output file name template
# %1 language, %2 translator, %3 manga title, %4 volume, %5 chapter/range, %6 chapter title
mdx dl --file-name "%3 ch.%5" mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370
# include custom static text
mdx dl --file-name "YourTextHere ch. %5" mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370
# use file name template with interactive mode
mdx dl -i --file-name "%3 ch.%5"
# include volume number
mdx dl --file-name "%3 vol.%4 ch.%5" mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370
# include language and translator
mdx dl --file-name "[%1 %2] %3 ch.%5" mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370
# include chapter title
mdx dl --file-name "%3 ch.%5 - %6" mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370
# merged chapters use %5 as chapter range
mdx dl -m -c 1-2 --file-name "%3 ch.%5" mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370
# merged volumes use %4 as volume and %5 as chapter range
mdx dl -m -v 1 --file-name "%3 vol.%4 ch.%5" mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370

# specify translation
mdx dl -t "Black Cat" mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370

# download compressed version (lower image quality and file size)
mdx dl -j mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370
```

Check available updates:

```sh
mdx update
```

Get help about subcommands and flags:

```sh
mdx
mdx -h
# ping subcommand is example
mdx ping
mdx ping -h
```

Search manga:

```sh
mdx find -t "Manga Title"
mdx search -t "Manga Title"
mdx f -t "Manga Title"
```

Get detailed information about the manga:

```sh
mdx info -u https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk
# or
mdx info mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk
```

Check connection to MangaDex API:

```sh
mdx ping
```

## FAQ 💬

#### Where can I get the manga link?

You can find the manga link at mangadex.org. Choose the manga you like and open its page. The link to the manga looks like this:

```
https://mangadex.org/title/abc-123-abc/some-title
```

You can use this link to download chapters of the manga.

#### Where can I get the chapter link?

Go to mangadex.org, choose the manga you like, and open the specific chapter you want. The link to the chapter looks like this:

```
https://mangadex.org/chapter/abc-123-abc
```

You can use this link to download the specific chapter.

#### I can't download chapter X of manga X. Why?

Make sure you have specified the correct language, translation group, and number of chapters. If you are unable to download a specific chapter, try using the direct link to the chapter:

```
mdx dl https://mangadex.org/chapter/abc-123-abc
```

**Remember:** mdx can only download chapters from MangaDex.

#### Why can't I install or run mdx on Windows?

This may happen because of an issue with the [signed `.exe` file](https://en.wikipedia.org/wiki/Code_signing). If the installer does not work, download the Windows `.zip` archive manually from the [Releases page](https://github.com/arimatakao/mdx/releases), extract it, and run `mdx.exe` from the extracted folder.

#### Why is downloading so slow?

I don't know. It's a problem on MangaDex's side or on your side.

#### I downloaded a chapter but the output filename doesn't have the volume number or chapter number. Why?

This problem stems from the uploader failing to specify the correct volume or chapter details.

#### Why do pages in the PDF have different sizes?

The size of each page in the PDF corresponds to the size of the image.

#### Getting error "Chapters x-y not found, try another range, language, translation group etc."

Maybe you didn't specify the translation group, chapter range, or language correctly. **Make sure that the chapter can be opened on MangaDex (not on external resource).**

Sometimes it doesn't download because of some problems on the MangaDex side. Just try again later.

## TODO 📌

### Functionality

- [X] Remove Doujinshi from list in `find` subcommand and add `doujinshi` flag for show Doujinshi in list.
- [X] Add metadata for cbz downloaded archive.
- [X] Add check update subcommand.
- [ ] Add flag to `download`:
    - [X] `merge` - download chapters in one file.
    - [X] `last` - download latest chapter.
    - [X] `this` - download a specific chapter using a link provided by the user.
    - [X] `extension` - sets the extension of the output file. Add file support formats:
        - [X] pdf (include metadata).
        - [X] epub (include metadata).
        - [X] directory (folder with images).
    - [X] `all` - download all chapters.
    - [X] `volume` - download all chapters of specified volume.
    - [X] `volume-range` - download all chapters of specified volume range.
    - [X] `volume-bundle` - download all chapters of volume into one file.
    - [ ] `oneshot` - download all oneshots of manga (if available).
- [X] Add interactive mode for `download` subcommand.
- [X] Add self update mechanism. (user should execute script for manual update)
- [ ] Add search filter for `find` subcommand.
- [X] Add flag `random` in `info` subcommand to get information about random manga.
- [ ] ~~Add interactive mode for `find` subcommand.~~ (already added into `download` subcommand)

### Code

- [X] Use `pterm` output instead `fmt`.
- [X] Refactor `cmd` package.
- [X] Add rate limiter for client api.
- [X] Create a Github action to automate the creation of `.deb` `.rpm` `.pkg.tar.zst` packages when a new release is created.
- [ ] Add tests for `mangadexapi` package.
- [ ] Refactor `internal/mdx` package.
- [ ] Refactor `mangadexapi` package.
- [ ] Refactor `filekit` package.

## License 📜

This project is licensed under the MIT - see the LICENSE file for details.

### Third-party Libraries

This project uses the following third-party libraries:

- Cobra (https://github.com/spf13/cobra) - Licensed under the Apache License 2.0
- Resty (https://github.com/go-resty/resty) - Licensed under the MIT
- PTerm (https://github.com/pterm/pterm) - Licensed under the MIT
- gopdf (https://github.com/signintech/gopdf) - Licensed under the MIT
- go-epub (https://github.com/go-shiori/go-epub) - Licensed under the MIT
- consolesize-go (https://github.com/nathan-fiscaletti/consolesize-go) - Licensed under the MIT

## Contributors 🧑‍💻

- [arimatakao](https://github.com/arimatakao)
- [wolandark](https://github.com/wolandark)
- [nikololiahim](https://github.com/nikololiahim)
- [BelardoA](https://github.com/BelardoA)
