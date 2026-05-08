package main

import "github.com/nskaggs/perfuncted"
import "image"
import "os"
import "image/png"
import (
	"math/rand"
	"time"

	cairo "github.com/neurlang/wayland/cairoshim"
	"github.com/neurlang/wayland/window"
	"github.com/neurlang/wayland/wl"

	"fmt"
)

type lens struct {
	display     *window.Display
	window      *window.Window
	widget      *window.Widget
	width       int32
	height      int32

	pf          *perfuncted.Perfuncted
}

func (lens *lens) Resize(_ *window.Widget, _ int32, _ int32, width int32, height int32) {

	size := int(width) * int(height)

	println("new size", size)

	lens.width = width
	lens.height = height

	//lens.widget.ScheduleResize(lens.width, lens.height)
}

func renderFrame(surface cairo.Surface, path string) {
    file, err := os.Open(path)
    if err != nil {
        return
    }
    defer file.Close()

    img, err := png.Decode(file)
    if err != nil {
        return
    }

    bounds := img.Bounds()

    dst := surface.ImageSurfaceGetData()
    stride := surface.ImageSurfaceGetStride()

    if dst == nil {
        return
    }

    for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
        for x := bounds.Min.X; x < bounds.Max.X; x++ {
            r, g, b, a := img.At(x, y).RGBA()

            off := y*stride + x*4
	if off >= len(dst) {continue;}

            dst[off+0] = byte(b >> 8)
            dst[off+1] = byte(g >> 8)
            dst[off+2] = byte(r >> 8)
            dst[off+3] = byte(a >> 8)
        }
    }
}

func (lens *lens) Redraw(widget *window.Widget) {
    surface := lens.window.WindowGetSurface()
    if surface != nil {
        renderFrame(surface, "/tmp/frame.png")
        surface.Destroy()
    }

    lens.widget.ScheduleRedraw()
}
func lensMotionHandler(lens *lens, x float32, y float32) {
	lens.pf.Input.MouseMove(int(x), int(y))
}
func (lens *lens) Key(
	_ *window.Window,
	input *window.Input,
	time uint32,
	key uint32,
	notUnicode uint32,
	state wl.KeyboardKeyState,
	_ window.WidgetHandler,
) {
	var entered = string(input.GetRune(&notUnicode, key))
	if state == wl.KeyboardKeyStatePressed {
		lens.pf.Input.KeyDown(entered)
		lens.pf.Input.Type(entered)
	} else {
		lens.pf.Input.KeyUp(entered)
	}
}
func (*lens) Focus(_ *window.Window, _ *window.Input) {

}
func (*lens) Enter(_ *window.Widget, _ *window.Input, x float32, y float32) {
}
func (*lens) Leave(_ *window.Widget, _ *window.Input) {
}

func (lens *lens) Motion(
	_ *window.Widget,
	_ *window.Input,
	time uint32,
	x float32,
	y float32,
) int {
	lensMotionHandler(lens, x, y)

	return window.CursorHand1
}

func (lens *lens) Button(
	_ *window.Widget,
	_ *window.Input,
	time uint32,
	button uint32,
	state wl.PointerButtonState,
	_ window.WidgetHandler,
) {
	var btn int
	switch button {
	case 0:
		btn = 1
	case 1:
		btn = 3
	case 2:
		btn = 2
	default:
		return
	}

	if state == wl.PointerButtonStatePressed {
		lens.pf.Input.MouseDown(btn)
	} else {
		lens.pf.Input.MouseUp(btn)
	}
}

func (*lens) TouchUp(
	_ *window.Widget,
	_ *window.Input,
	serial uint32,
	time uint32,
	id int32,
) {
}

func (*lens) TouchDown(
	_ *window.Widget,
	_ *window.Input,
	serial uint32,
	time uint32,
	id int32,
	x float32,
	y float32,
) {
}

func (lens *lens) TouchMotion(
	_ *window.Widget,
	_ *window.Input,
	time uint32,
	id int32,
	x float32,
	y float32,
) {

	lensMotionHandler(lens, x, y)

}
func (*lens) TouchFrame(_ *window.Widget, _ *window.Input) {
}
func (*lens) TouchCancel(_ *window.Widget, width int32, height int32) {
}

func (lens *lens) Axis(
	_ *window.Widget,
	_ *window.Input,
	time uint32,
	axis uint32,
	value float32,
) {
	if value == 0 {
		return
	}
	clicks := int(value)
	if axis == 1 {
		if clicks > 0 {
			lens.pf.Input.ScrollUp(clicks)
		} else {
			lens.pf.Input.ScrollDown(-clicks)
		}
	} else if axis == 0 {
		if clicks > 0 {
			lens.pf.Input.ScrollRight(clicks)
		} else {
			lens.pf.Input.ScrollLeft(-clicks)
		}
	}
}
func (*lens) AxisSource(_ *window.Widget, _ *window.Input, source uint32) {
}
func (*lens) AxisStop(_ *window.Widget, _ *window.Input, time uint32, axis uint32) {
}

func (lens *lens) AxisDiscrete(
	_ *window.Widget,
	_ *window.Input,
	axis uint32,
	discrete int32,
) {
	if discrete == 0 {
		return
	}
	if axis == 1 {
		if discrete > 0 {
			lens.pf.Input.ScrollUp(int(discrete))
		} else {
			lens.pf.Input.ScrollDown(-int(discrete))
		}
	} else if axis == 0 {
		if discrete > 0 {
			lens.pf.Input.ScrollRight(int(discrete))
		} else {
			lens.pf.Input.ScrollLeft(-int(discrete))
		}
	}
}
func (*lens) PointerFrame(_ *window.Widget, _ *window.Input) {
}

func (lens *lens) free() {
}

func main() {

	var lens lens

	d, err := window.DisplayCreate([]string{})
	if err != nil {
		fmt.Println(err)
		return
	}

	go loop(&lens)

	lens.width = 200
	lens.height = 200
	lens.display = d
	lens.window = window.Create(d)

	lens.widget = lens.window.AddWidget(&lens)

	lens.window.SetTitle("lens")
	lens.window.SetBufferType(window.BufferTypeShm)
	lens.window.SetKeyboardHandler(&lens)
	rand.Seed(int64(time.Now().Nanosecond()))

	lens.widget.SetUserDataWidgetHandler(&lens)

	lens.widget.ScheduleResize(lens.width, lens.height)

	window.DisplayRun(d)

	lens.widget.Destroy()
	lens.window.Destroy()
	d.Destroy()

}

func loop(lens *lens) {
	pf, err := perfuncted.New(perfuncted.Options{
		Nested: true,


	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer pf.Close()

	lens.pf = pf

	for {
		pf.Screen.CaptureRegion(
			image.Rect(0, 0, 800, 600),
			"/tmp/frame.png",
		)

		// reload image in viewer

		time.Sleep(time.Millisecond * 33)
	}
}
