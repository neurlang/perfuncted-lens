package main

import "github.com/nskaggs/perfuncted"
import "image"
import "context"
import (
	"math/rand"
	"time"

	cairo "github.com/neurlang/wayland/cairoshim"
	"github.com/neurlang/wayland/window"
	"github.com/neurlang/wayland/wl"

	"fmt"
	"sync"
)

type lens struct {
	display *window.Display
	window  *window.Window
	widget  *window.Widget
	width   int32
	height  int32

	img      *image.Image
	imgMutex sync.Mutex

	pf *perfuncted.Perfuncted
}

func (lens *lens) Resize(_ *window.Widget, _ int32, _ int32, width int32, height int32) {

	size := int(width) * int(height)

	println("new size", size)
	//lens.widget.ScheduleResize(lens.width, lens.height)
}

func (lens *lens) renderFrame(surface cairo.Surface) {
	lens.imgMutex.Lock()
	img := *lens.img
	lens.imgMutex.Unlock()

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
			if off >= len(dst) {
				continue
			}

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
		lens.renderFrame(surface)
		surface.Destroy()
	}

	lens.widget.ScheduleRedraw()
}
func lensMotionHandler(lens *lens, x float32, y float32) {
	lens.pf.Input.MouseMove(context.Background(), int(x), int(y))
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
		lens.pf.Input.KeyDown(context.Background(), entered)
		lens.pf.Input.Type(context.Background(), entered)
	} else {
		lens.pf.Input.KeyUp(context.Background(), entered)
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
	var btn = int(button)

	if state == wl.PointerButtonStatePressed {
		lens.pf.Input.MouseDown(context.Background(), btn)
	} else {
		lens.pf.Input.MouseUp(context.Background(), btn)
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
			lens.pf.Input.ScrollUp(context.Background(), clicks)
		} else {
			lens.pf.Input.ScrollDown(context.Background(), -clicks)
		}
	} else if axis == 0 {
		if clicks > 0 {
			lens.pf.Input.ScrollRight(context.Background(), clicks)
		} else {
			lens.pf.Input.ScrollLeft(context.Background(), -clicks)
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
			lens.pf.Input.ScrollUp(context.Background(), int(discrete))
		} else {
			lens.pf.Input.ScrollDown(context.Background(), -int(discrete))
		}
	} else if axis == 0 {
		if discrete > 0 {
			lens.pf.Input.ScrollRight(context.Background(), int(discrete))
		} else {
			lens.pf.Input.ScrollLeft(context.Background(), -int(discrete))
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
		pixels, err := pf.Screen.GetAllPixels(
			context.Background(),
		)
		if err != nil {
			fmt.Println(err)
			return
		}
		lens.imgMutex.Lock()
		lens.img = &pixels
		lens.imgMutex.Unlock()
		// reload image in viewer

		time.Sleep(time.Millisecond * 1)
	}
}
