package optics

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"cloud.google.com/go/vertexai/genai"
)

// VisionClient wraps the Vertex AI client
type VisionClient struct {
	projectID string
	location  string
	client    *genai.Client
}

// NewVisionClient initializes the connection to Vertex AI
func NewVisionClient(projectID, location string) (*VisionClient, error) {
	client, err := genai.NewClient(context.Background(), projectID, location)
	if err != nil {
		return nil, fmt.Errorf("failed to create vertex ai client: %w", err)
	}

	return &VisionClient{
		projectID: projectID,
		location:  location,
		client:    client,
	}, nil
}

// LocateElement takes image bytes and a prompt, returning normalized coordinates [ymin, xmin, ymax, xmax].
func (vc *VisionClient) LocateElement(ctx context.Context, imageBytes []byte, prompt string) ([]float64, error) {
	fmt.Printf("Optics: Sending image (%d bytes) to Vertex AI for spatial analysis...\n", len(imageBytes))
	fmt.Printf("Optics Prompt: %s\n", prompt)

	model := vc.client.GenerativeModel("gemini-2.5-flash")
	model.SetTemperature(0.0) // We want deterministic bounding boxes

	imgData := genai.ImageData("jpeg", imageBytes)
	
	// Enforce strict bounding box instruction just to be safe
	strictPrompt := prompt + " Return only the bounding box in the format [ymin, xmin, ymax, xmax]."
	
	resp, err := model.GenerateContent(ctx, imgData, genai.Text(strictPrompt))
	if err != nil {
		return nil, fmt.Errorf("vertex ai inference failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("vertex ai returned empty response")
	}

	var responseText string
	switch part := resp.Candidates[0].Content.Parts[0].(type) {
	case genai.Text:
		responseText = string(part)
	default:
		return nil, fmt.Errorf("unexpected response type from vertex ai")
	}
	
	fmt.Printf("Raw Gemini Response: %s\n", responseText)

	// Extract the [ymin, xmin, ymax, xmax] coordinates
	re := regexp.MustCompile(`\[(\d+),\s*(\d+),\s*(\d+),\s*(\d+)\]`)
	matches := re.FindStringSubmatch(responseText)

	if len(matches) < 5 {
		return nil, fmt.Errorf("could not find bounding box in response: %s", responseText)
	}

	ymin, _ := strconv.ParseFloat(matches[1], 64)
	xmin, _ := strconv.ParseFloat(matches[2], 64)
	ymax, _ := strconv.ParseFloat(matches[3], 64)
	xmax, _ := strconv.ParseFloat(matches[4], 64)

	return []float64{ymin, xmin, ymax, xmax}, nil
}

// AnalyzeImage takes image bytes and a prompt, returning the raw text string from Vertex.
// Used for cognitive decision making (e.g., YES/NO filters).
func (vc *VisionClient) AnalyzeImage(ctx context.Context, imageBytes []byte, prompt string) (string, error) {
	fmt.Printf("Optics: Sending image (%d bytes) to Vertex AI for cognitive analysis...\n", len(imageBytes))
	fmt.Printf("Optics Prompt: %s\n", prompt)

	model := vc.client.GenerativeModel("gemini-2.5-flash")
	model.SetTemperature(0.1) // Slight variance for text generation, but mostly deterministic

	imgData := genai.ImageData("jpeg", imageBytes)
	
	resp, err := model.GenerateContent(ctx, imgData, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("vertex ai inference failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("vertex ai returned empty response")
	}

	var responseText string
	switch part := resp.Candidates[0].Content.Parts[0].(type) {
	case genai.Text:
		responseText = string(part)
	default:
		return "", fmt.Errorf("unexpected response type from vertex ai")
	}
	
	fmt.Printf("Raw Cognitive Response: %s\n", responseText)
	return responseText, nil
}
