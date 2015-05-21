# andy
Andy helps Android projects

## install
```
go get github.com/mcginty/andy
```

## use
`andy dpi <asset>` will take an asset you dumped into your res folder appropriately and resize for any lower densities.
```
andy dpi icon.png
```

`andy convert <Xdp>` quickly converts a density independent value to corresponding pixel values. Comes in handy when doing asset designs.

```
andy convert 3.2dp
```
