<div align="center">

# mdx 📚

mdx is a simple CLI application for downloading manga from the [MangaDex website](https://mangadex.org/). The program uses [MangaDex API](https://api.mangadex.org/docs/) to fetch manga content.

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/arimatakao/mdx)
![GitHub Release](https://img.shields.io/github/v/release/arimatakao/mdx)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/arimatakao/mdx/total)
![GitHub Repo stars](https://img.shields.io/github/stars/arimatakao/mdx)

![demo v1.4.0](./.github/assets/demo.gif)

</div>

## Features 💫

- Works on ***Windows, MacOS, Linux***.
- Downloads ***multiple chapters***.
- Saving manga in ***CBZ, PDF, EPUB formats***.
- Saving multiple chapters in ***one file***.
- Automatically generates metadata for downloaded files, ***adapted for e-readers***.
- Searches manga.
- Displays information about manga.

## Installation ⚙️

1. Download `.tar.gz` archive from [releases page](https://github.com/arimatakao/mdx/releases).
2. Unarchive the `.tar.gz` file you downloaded.

Open unarchived folder and execute `mdx` file for use application.

Also, you can install the application with `go`:

```
go install github.com/arimatakao/mdx@latest
```

## Usage examples️ 🖥️

Download manga:

```shell
# get help
mdx download --help

# by default 1 chapter is downloading
mdx download -u https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk
# or
mdx dl -u https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk
# or
mdx dl https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84
# or
mdx dl mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84

# download pdf format instead cbz
mdx dl -e pdf mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84

# download specific chapter
mdx dl -c 123 mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk

# download range chapters
mdx dl -c 12-34 mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84

# download range chapters and merge in one file
mdx dl -m -c 12-34 mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84

# specify language, by default is english (for get available languages execute info subcommand)
mdx dl -l it mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk

# specify output directory
mdx dl -o your/dir mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk

# specify translation
mdx dl -t "Some Group" mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk

# download compressed version (lower image quality and file size)
mdx dl -j mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk
```

Check available updates:

```shell
mdx update
```

Get help about subcommands and flags:

```shell
mdx
mdx -h
# ping subcommand is example
mdx ping
mdx ping -h
```

Search manga:

```shell
mdx find -t "Manga Title"
mdx search -t "Manga Title"
mdx f -t "Manga Title"
```

Get detail information about manga:

```shell
mdx info -u https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk
# or
mdx info mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk
```

Check connection to MangaDex API:

```shell
mdx ping
```

## TODO 📌

### Functionality

- [X] Remove Doujinshi from list in `find` subcommand and add `doujinshi` flag for show Doujinshi in list.
- [X] Add metadata for cbz downloaded archive.
- [X] Add check update subcommand.
- [ ] Add self update mechanism.
- [ ] Add search filter for `find` subcommand.
- [ ] Add flag `random` in `info` subcommand for get information about random manga.
- [ ] Add flag to `download`:
    - [ ] `last` - download latest chapter.
    - [ ] `this` - download specific chapter using link from user. Make download chapter get chapter link instead manga link.
    - [ ] `volume` - download all chapter of specified volume.
    - [ ] `volume-range` - download all chapter of specified volume range.
    - [ ] `oneshot` - download all oneshots of manga (if available).
    - [ ] `all` - download all chapters.
    - [X] `merge` - download chapter in one file.
    - [ ] `volume-bundle` - download all chapters of volume into one file.
    - [X] `extension` - sets the extension of the outpud file. Add file support formats:
        - [X] pdf (include metadata).
        - [X] epub (include metadata).
- [ ] Add interactive mode for `find` subcommand.
- [ ] Add interactive mode for `download` subcommand.

### Code

- [ ] Use `pterm` output instead `fmt`
- [ ] Add tests for `mangadexapi` package.
- [ ] Refactor `mangadexapi` package.
- [ ] Refactor `cmd` package.
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
