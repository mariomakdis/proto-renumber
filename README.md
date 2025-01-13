# proto-renumber

A CLI tool to automatically renumber field tags in your Protocol Buffer files.

## Screenshots

Example Proto schema:
![img](/images/before.png)

Running `proto-renumber` with `--replace`:
![img](/images/command.png)

Results
![img](/images/after.png)

## Installation
```
go install github.com/mariomakdis/proto-renumber@v1.1.0
```

## Usage

```bash
proto-renumber [--ignore] MessageA,NestedMessage [--replace] path/to/file.proto
```
