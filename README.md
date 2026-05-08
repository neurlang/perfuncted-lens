# perfuncted-lens
perfuncted-lens

## Get started
```
go install github.com/nskaggs/perfuncted/cmd/pf@latest
pf session start
go install neurlang/perfuncted-lens/pflens/cmd/pflens@latest
pflens
```
Then run some app:
```
XDG_RUNTIME_DIR=/tmp/perfuncted-xdg-2830287735 WAYLAND_DISPLAY=wayland-1 ./go-wayland-smoke
```
