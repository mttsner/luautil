# LuaUtil

LuaUtil is a small library written in go that provides the toolset to work with lua code in a structured manner.

# For contributors

To update the yacc stuff you need goyacc.

```bash
go get modernc.org/goyacc
```

Command to generate the go file from yacc: 

```bash
goyacc -o parser.go parser.y
```

You can delete the y.output file afterwards.

# Sources

The parser and ast is forked from [gopher-lua](https://github.com/yuin/gopher-lua) and somewhat modified.

The ssa implementation is inspired by the [GoLang ssa](https://pkg.go.dev/golang.org/x/tools/go/ssa) package.