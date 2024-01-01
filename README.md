<div align="center">
	<span style="font-weight: bold"> <a> English </a> </span>
</div>

# Gochronize - Go synchronize
[![GitHub release](https://img.shields.io/github/v/release/XayahSuSuSu/gochronize?color=orange)](https://github.com/XayahSuSuSu/gochronize/releases) [![License](https://img.shields.io/github/license/XayahSuSuSu/gochronize?color=ff69b4)](./LICENSE)

## Overview
A tool for synchronizing releases from GitHub with local.

## Usage
```
gochronize --config "example.yml" --history "history.yml"

Available arguments:
  -config string
        The configuration path of yaml file format.
  -dry-run
        Do a trial run with no downloads.
  -help
        Print the usage.
  -history string
        The history configuration path of yaml file format. (default "history.yml")
  -version
        Print the version.
```

## Config
Refer to [example.yml](./example.yml)

## LICENSE
[GNU General Public License v3.0](./LICENSE)
