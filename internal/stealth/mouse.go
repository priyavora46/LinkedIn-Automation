package stealth

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// Point represents a 2D coordinate
type Point struct {
	X, Y float64
}

// BezierCurve generates points along a Bezier curve for natural mouse movement
func BezierCurve(start, end Point, control1, control2 Point, steps int) []Point {
	points := make([]Point, steps)

	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)

		// Cubic Bezier formula
		x := math.Pow(1-t, 3)*start.X +
			3*math.Pow(1-t, 2)*t*control1.X +
			3*(1-t)*math.Pow(t, 2)*control2.X +
			math.Pow(t, 3)*end.X

		y := math.Pow(1-t, 3)*start.Y +
			3*math.Pow(1-t, 2)*t*control1.Y +
			3*(1-t)*math.Pow(t, 2)*control2.Y +
			math.Pow(t, 3)*end.Y

		points[i] = Point{X: x, Y: y}
	}

	return points
}

// HumanMouseMove moves mouse in a human-like pattern using Bezier curves
func HumanMouseMove(page *rod.Page, targetX, targetY float64) error {
	// Get current mouse position (start from a random nearby position)
	startX := rand.Float64() * 100
	startY := rand.Float64() * 100

	// Generate random control points for natural curve
	dx := targetX - startX
	dy := targetY - startY

	control1 := Point{
		X: startX + dx*0.25 + rand.Float64()*50 - 25,
		Y: startY + dy*0.25 + rand.Float64()*50 - 25,
	}

	control2 := Point{
		X: startX + dx*0.75 + rand.Float64()*50 - 25,
		Y: startY + dy*0.75 + rand.Float64()*50 - 25,
	}

	// Generate curve points
	steps := 15 + rand.Intn(10) // 15-25 steps
	points := BezierCurve(
		Point{X: startX, Y: startY},
		Point{X: targetX, Y: targetY},
		control1,
		control2,
		steps,
	)

	// Move along the curve with variable speed using low-level protocol
	for i, p := range points {
		// Variable speed - slower at start/end, faster in middle
		var delay time.Duration
		progress := float64(i) / float64(len(points))
		if progress < 0.2 || progress > 0.8 {
			delay = time.Duration(15+rand.Intn(10)) * time.Millisecond
		} else {
			delay = time.Duration(5+rand.Intn(5)) * time.Millisecond
		}

		// Use proto directly for mouse movement
		err := proto.InputDispatchMouseEvent{
			Type: proto.InputDispatchMouseEventTypeMouseMoved,
			X:    p.X,
			Y:    p.Y,
		}.Call(page)

		if err != nil {
			return err
		}
		time.Sleep(delay)

		// Add occasional micro-corrections
		if rand.Float64() < 0.1 {
			jitterX := p.X + rand.Float64()*4 - 2
			jitterY := p.Y + rand.Float64()*4 - 2

			err = proto.InputDispatchMouseEvent{
				Type: proto.InputDispatchMouseEventTypeMouseMoved,
				X:    jitterX,
				Y:    jitterY,
			}.Call(page)

			if err != nil {
				return err
			}
			time.Sleep(time.Duration(5+rand.Intn(5)) * time.Millisecond)
		}
	}

	// Final positioning with slight overshoot and correction
	overshootX := targetX + rand.Float64()*6 - 3
	overshootY := targetY + rand.Float64()*6 - 3

	err := proto.InputDispatchMouseEvent{
		Type: proto.InputDispatchMouseEventTypeMouseMoved,
		X:    overshootX,
		Y:    overshootY,
	}.Call(page)

	if err != nil {
		return err
	}
	time.Sleep(time.Duration(10+rand.Intn(10)) * time.Millisecond)

	err = proto.InputDispatchMouseEvent{
		Type: proto.InputDispatchMouseEventTypeMouseMoved,
		X:    targetX,
		Y:    targetY,
	}.Call(page)

	if err != nil {
		return err
	}
	time.Sleep(time.Duration(20+rand.Intn(20)) * time.Millisecond)

	return nil
}

// HumanClick performs a human-like click with natural timing
func HumanClick(page *rod.Page, el *rod.Element) error {
	// Scroll element into view first
	el.MustScrollIntoView()
	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

	// Brief pause before clicking
	time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

	// Click the element
	return el.Click(proto.InputMouseButtonLeft, 1)
}

// RandomMouseWander simulates idle mouse movement
func RandomMouseWander(page *rod.Page) {
	if rand.Float64() > 0.3 { // 30% chance to wander
		return
	}

	// Simple random movement using eval
	x := rand.Intn(500) + 100
	y := rand.Intn(500) + 100
	page.MustEval(fmt.Sprintf(`
		() => {
			const event = new MouseEvent('mousemove', {
				clientX: %d,
				clientY: %d,
				bubbles: true
			});
			document.dispatchEvent(event);
		}
	`, x, y))
	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
}

// HoverElement simulates hovering over an element
func HoverElement(page *rod.Page, el *rod.Element) error {
	// Scroll into view and hover
	el.MustScrollIntoView()
	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
	return el.Hover()
}
