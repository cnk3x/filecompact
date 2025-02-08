# filecompact

filecompact is a tool to find duplicate files in a directory.

Usage:

```
filecompact -s source -e exclude -d destination
```

Options:

```
  -d, --debug             debug mode
      --delete            delete files
  -e, --exclude strings   exclude files
      --load string       load collect database file
      --save string       save collect database file
  -s, --source strings    source directories
  -S, --strict            strict mode
```

Examples:

```
filecompact -s . -e .git -d compacted
``` 


