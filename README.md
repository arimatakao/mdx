# mdx

mdx is a command-line interface program for downloading manga from the [MangaDex website](https://mangadex.org/). The program uses [MangaDex API](https://api.mangadex.org/docs/) to fetch manga content.

## Installation

1. Download `.tar.gz` archive from [releases page](https://github.com/arimatakao/mdx/releases).
2. Unarchive the `.tar.gz` file you downloaded.

Open unarhived folder and execute `mdx` file for use application.

## Usage examples

Download manga:

```
# get information about available flags
mdx download --help

# by default 1 chapter is downloading
mdx download -u https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk
# or
mdx dl -u https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk

# specify specific chapter
mdx dl -c 123 -u https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk

# specify language (for get available languages execute info subcommand)
mdx dl -l it -c 123 -u https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk

# specify output directory
mdx dl -o your/dir -l it -c 123 -u https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk

# specify translation
mdx dl -t Marcelo -o your/dir -l it -c 123 -u https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk

# download compressed version (lower image quality and cbz file size)
mdx dl -j -t Marcelo -o your/dir -l it -c 123 -u https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk
```

Get help about subcommands and flags:

```
mdx
mdx -h
# ping subcommand is example
mdx ping
mdx ping -h
```

Search manga:

```
mdx find -t "Manga Title"
mdx search -t "Manga Title"
mdx f -t "Manga Title"
```

Get detail information about manga:

```
mdx info -u "https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk"
```

Check connection to MangaDex API:

```
mdx ping
```

## License

This project is licensed under the MIT - see the LICENSE file for details.

### Third-party Libraries

This project uses the following third-party libraries:

- Cobra (https://github.com/spf13/cobra) - Licensed under the Apache License 2.0
- Resty (https://github.com/go-resty/resty) - Licensed under the MIT
- PTerm (https://github.com/pterm/pterm) - Licensed under the MIT
