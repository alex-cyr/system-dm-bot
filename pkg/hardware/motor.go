package hardware

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"math/rand"
	"os"
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

	// Observability: Save to disk so developers can conceptually "see" what the bot sees
	_ = os.WriteFile("debug_vision.jpg", buf.Bytes(), 0644)

	return buf.Bytes(), nil
}

// ScrollDown scrolls the mouse wheel down to see more DMs.
func ScrollDown() {
	fmt.Println("Hardware: Scrolling down")
	// robotgo.Scroll(x, y) where x is horizontal and y is vertical.
	// Positive y is scroll down in some OSes, but usually negative y is down. 
	// In RobotGo on Windows, y > 0 is scroll up, y < 0 is scroll down. 
	robotgo.Scroll(0, -100)
	time.Sleep(1 * time.Second)
}

// RefreshPage presses the F5 key to reload the browser.
func RefreshPage() {
	fmt.Println("Hardware: Pressing F5 to refresh page")
	robotgo.KeyTap("f5")
	// Give the browser 5 seconds to fully reload the page
	time.Sleep(5 * time.Second)
}

// PressEnter presses the physical Enter key.
func PressEnter() {
	fmt.Println("Hardware: Pressing Enter key")
	robotgo.KeyTap("enter")
	time.Sleep(500 * time.Millisecond)
}
