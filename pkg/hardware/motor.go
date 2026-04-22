package hardware

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"math/rand"
	"time"

	"github.com/go-vgo/robotgo"
)

// InitMotor initializes any required hardware configurations.
func InitMotor() {
	robotgo.MouseSleep = 100
}

// MoveSmooth moves the mouse cursor to (x,y) along a human-like Bezier curve.
func MoveSmooth(x, y int) {
	fmt.Printf("Hardware: Moving mouse to (%d, %d)\n", x, y)
	
	// Adding slight randomization to end coordinates
	posX := x + rand.Intn(5) - 2
	posY := y + rand.Intn(5) - 2

	// Low/High dictate the speed and curve variance
	robotgo.MoveSmooth(posX, posY, 1.0, 2.0)
	
	// Human hesitation before click
	delay := time.Duration(rand.Intn(300)+200) * time.Millisecond
	time.Sleep(delay)
}

// Click initiates a standard left mouse click.
func Click() {
	fmt.Println("Hardware: Left Click")
	robotgo.Click("left")
}

// TypeStrDelay slowly types out text imitating human speed rhythms.
func TypeStrDelay(text string) {
	fmt.Printf("Hardware: Typing '%s'\n", text)
	for _, char := range text {
		robotgo.TypeStr(string(char))
		// Random delay between 80ms to 200ms
		delay := time.Duration(rand.Intn(120)+80) * time.Millisecond
		time.Sleep(delay)
	}
}

// CaptureScreen returns the current screen buffer as a JPEG byte slice.
func CaptureScreen() ([]byte, error) {
	fmt.Println("Hardware: Capturing screen buffer")
	
	img, err := robotgo.CaptureImg()
	if err != nil {
		return nil, fmt.Errorf("failed to capture screen: %w", err)
	}

	buf := new(bytes.Buffer)
	// Compress to 85% JPEG to save Vertex AI bandwidth
	err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return nil, fmt.Errorf("failed to encode screen to JPEG: %w", err)
	}

	return buf.Bytes(), nil
}
