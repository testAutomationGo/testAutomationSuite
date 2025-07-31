package videoTypeCreation

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type VideoGenerator struct {
	Duration  int // seconds
	Width     int
	Height    int
	Framerate int
	uniqueID  string
}

func NewVideoGenerator() *VideoGenerator {
	return &VideoGenerator{
		Duration:  10,
		Width:     640,
		Height:    480,
		Framerate: 30,
		uniqueID:  generateUniqueID(),
	}
}

func generateUniqueID() string {
	timestamp := time.Now().Unix()
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("%d_%d", timestamp, randomNum)
}

func getRandomElement(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(slice))))
	return slice[idx.Int64()]
}

func getRandomInt(min, max int) int {
	if min >= max {
		return min
	}
	diff := max - min + 1
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(diff)))
	return min + int(n.Int64())
}

func (vg *VideoGenerator) CreateM4V(outputFolder string) (string, error) {

	filename := fmt.Sprintf("video_%s_%s.m4v", vg.uniqueID, generateUniqueID())
	outputPath := filepath.Join(outputFolder, filename)

	colors := []string{"red", "blue", "green", "yellow", "purple", "orange", "pink", "cyan"}
	texts := []string{"Test Video", "Sample Content", "Generated Video", "M4V Test"}

	color := getRandomElement(colors)
	text := getRandomElement(texts)
	fontSize := getRandomInt(20, 48)
	textColor := getRandomElement([]string{"white", "black", "yellow"})

	videoFilter := fmt.Sprintf("color=%s:duration=%d:size=%dx%d:rate=%d,drawtext=text='%s %s':fontcolor=%s:fontsize=%d:x=w-mod(t*50\\,w+tw):y=(h-text_h)/2",
		color, vg.Duration, vg.Width, vg.Height, vg.Framerate, text, vg.uniqueID[:8], textColor, fontSize)

	if getRandomInt(1, 3) == 1 {
		videoFilter += fmt.Sprintf(",noise=alls=%d", getRandomInt(5, 15))
	}

	args := []string{
		"-f", "lavfi",
		"-i", videoFilter,
		"-f", "lavfi",
		"-i", fmt.Sprintf("sine=frequency=%d:duration=%d", getRandomInt(400, 800), vg.Duration),
		"-c:v", "libx264",
		"-c:a", "aac",
		"-pix_fmt", "yuv420p",
		"-shortest",
		"-y", outputPath,
	}

	return vg.runFFmpeg(args, outputPath, "M4V")
}

func (vg *VideoGenerator) CreateMP4(outputFolder string) (string, error) {

	filename := fmt.Sprintf("video_%s_%s.mp4", vg.uniqueID, generateUniqueID())
	outputPath := filepath.Join(outputFolder, filename)

	colors := []string{"red", "blue", "green", "yellow", "purple", "orange", "pink", "cyan", "magenta", "lime"}
	texts := []string{"Test Video", "Sample Content", "Generated Video", "MP4 Test", "Automation Video"}
	movements := []string{"scroll", "bounce", "fade"}

	color := getRandomElement(colors)
	text := getRandomElement(texts)
	movement := getRandomElement(movements)
	fontSize := getRandomInt(24, 52)
	textColor := getRandomElement([]string{"white", "black", "yellow", "red"})

	var videoFilter string

	switch movement {
	case "scroll":
		videoFilter = fmt.Sprintf("color=%s:duration=%d:size=%dx%d:rate=%d,drawtext=text='%s %s':fontcolor=%s:fontsize=%d:x=w-mod(t*60\\,w+tw):y=(h-text_h)/2",
			color, vg.Duration, vg.Width, vg.Height, vg.Framerate, text, vg.uniqueID[:8], textColor, fontSize)
	case "bounce":
		videoFilter = fmt.Sprintf("color=%s:duration=%d:size=%dx%d:rate=%d,drawtext=text='%s %s':fontcolor=%s:fontsize=%d:x=(w-text_w)/2:y=h/2+sin(t*2)*h/4",
			color, vg.Duration, vg.Width, vg.Height, vg.Framerate, text, vg.uniqueID[:8], textColor, fontSize)
	case "fade":
		videoFilter = fmt.Sprintf("color=%s:duration=%d:size=%dx%d:rate=%d,drawtext=text='%s %s':fontcolor=%s:fontsize=%d:x=(w-text_w)/2:y=(h-text_h)/2:alpha='sin(t)'",
			color, vg.Duration, vg.Width, vg.Height, vg.Framerate, text, vg.uniqueID[:8], textColor, fontSize)
	}

	effectChance := getRandomInt(1, 4)
	switch effectChance {
	case 1:
		videoFilter += fmt.Sprintf(",noise=alls=%d", getRandomInt(5, 20))
	case 2:
		videoFilter += fmt.Sprintf(",colorbalance=rm=%f:gm=%f:bm=%f",
			float64(getRandomInt(-30, 30))/100.0,
			float64(getRandomInt(-30, 30))/100.0,
			float64(getRandomInt(-30, 30))/100.0)
	case 3:
		videoFilter += ",vignette=angle=PI/4"
	}

	args := []string{
		"-f", "lavfi",
		"-i", videoFilter,
		"-f", "lavfi",
		"-i", fmt.Sprintf("sine=frequency=%d:duration=%d", getRandomInt(300, 1000), vg.Duration),
		"-c:v", "libx264",
		"-c:a", "aac",
		"-pix_fmt", "yuv420p",
		"-shortest",
		"-y", outputPath,
	}

	return vg.runFFmpeg(args, outputPath, "MP4")
}

func (vg *VideoGenerator) CreateWebM(outputFolder string) (string, error) {

	filename := fmt.Sprintf("video_%s_%s.webm", vg.uniqueID, generateUniqueID())
	outputPath := filepath.Join(outputFolder, filename)

	colors := []string{"red", "blue", "green", "yellow", "purple", "orange", "pink", "cyan",
		"magenta", "lime", "teal", "coral", "gold", "silver"}
	texts := []string{"WebM Video", "VP9 Test", "Generated Content", "WebM Sample",
		"Streaming Video", "Web Content", "VP9 Demo"}
	patterns := []string{"spiral", "grid", "waves", "particles", "plasma"}

	color := getRandomElement(colors)
	text := getRandomElement(texts)
	pattern := getRandomElement(patterns)
	fontSize := getRandomInt(28, 56)
	textColor := getRandomElement([]string{"white", "black", "yellow", "red", "cyan"})

	var videoFilter string

	switch pattern {
	case "spiral":
		videoFilter = fmt.Sprintf("color=%s:duration=%d:size=%dx%d:rate=%d",
			color, vg.Duration, vg.Width, vg.Height, vg.Framerate)
		videoFilter += fmt.Sprintf(",drawtext=text='%s %s':fontcolor=%s:fontsize=%d:x=(w-text_w)/2+cos(t)*w/4:y=(h-text_h)/2+sin(t)*h/4",
			text, vg.uniqueID[:8], textColor, fontSize)

	case "grid":
		videoFilter = fmt.Sprintf("color=%s:duration=%d:size=%dx%d:rate=%d,drawgrid=width=40:height=40:thickness=2:color=white@0.3",
			color, vg.Duration, vg.Width, vg.Height, vg.Framerate)
		videoFilter += fmt.Sprintf(",drawtext=text='%s %s':fontcolor=%s:fontsize=%d+sin(t*3)*10:x=(w-text_w)/2:y=(h-text_h)/2",
			text, vg.uniqueID[:8], textColor, fontSize)

	case "waves":
		videoFilter = fmt.Sprintf("color=%s:duration=%d:size=%dx%d:rate=%d",
			color, vg.Duration, vg.Width, vg.Height, vg.Framerate)
		videoFilter += fmt.Sprintf(",drawtext=text='%s %s':fontcolor=%s:fontsize=%d:x=w/2+sin(t*2)*w/3:y=h/2+cos(t*3)*h/3",
			text, vg.uniqueID[:8], textColor, fontSize)

	case "particles":
		videoFilter = fmt.Sprintf("color=%s:duration=%d:size=%dx%d:rate=%d",
			color, vg.Duration, vg.Width, vg.Height, vg.Framerate)
		videoFilter += fmt.Sprintf(",drawtext=text='%s':fontcolor=%s:fontsize=%d:x=(w-text_w)/2:y=(h-text_h)/2",
			text, textColor, fontSize)
		videoFilter += fmt.Sprintf(",drawtext=text='%s':fontcolor=%s@0.7:fontsize=%d:x=w/2+cos(t*4)*w/3:y=h/2+sin(t*4)*h/3",
			vg.uniqueID[:8], textColor, fontSize/2)

	case "plasma":
		videoFilter = fmt.Sprintf("color=%s:duration=%d:size=%dx%d:rate=%d",
			color, vg.Duration, vg.Width, vg.Height, vg.Framerate)
		videoFilter += fmt.Sprintf(",drawtext=text='%s %s':fontcolor=%s:fontsize=%d:x=(w-text_w)/2:y=(h-text_h)/2",
			text, vg.uniqueID[:8], textColor, fontSize)
		videoFilter += ",hue=h=t*60"
	}

	effectType := getRandomInt(1, 5)
	switch effectType {
	case 1:
		videoFilter += ",eq=brightness=sin(t*0.5)*0.3"
	case 2:
		videoFilter += ",eq=contrast=1+sin(t)*0.5"
	case 3:
		videoFilter += ",eq=saturation=1+cos(t*2)*0.5"
	case 4:
		videoFilter += ",eq=gamma=1+sin(t*0.8)*0.3"
	}

	args := []string{
		"-f", "lavfi",
		"-i", videoFilter,
		"-f", "lavfi",
		"-i", fmt.Sprintf("sine=frequency=%d:duration=%d", getRandomInt(200, 600), vg.Duration),
		"-c:v", "libvpx-vp9",
		"-c:a", "libopus",
		"-crf", fmt.Sprintf("%d", getRandomInt(25, 35)),
		"-b:v", "0",
		"-row-mt", "1",
		"-shortest",
		"-y", outputPath,
	}

	return vg.runFFmpeg(args, outputPath, "WebM")
}

func (vg *VideoGenerator) Create3GP(outputFolder string) (string, error) {

	filename := fmt.Sprintf("video_%s_%s.3gp", vg.uniqueID, generateUniqueID())
	outputPath := filepath.Join(outputFolder, filename)

	colors := []string{"red", "blue", "green", "yellow", "orange", "purple", "pink", "white"}
	texts := []string{"3GP Mobile", "Phone Video", "Mobile Test", "3GP Demo", "Cell Video", "Portable"}
	mobileEffects := []string{"simple", "ticker", "flash", "pulse", "zoom"}

	color := getRandomElement(colors)
	text := getRandomElement(texts)
	effect := getRandomElement(mobileEffects)
	fontSize := getRandomInt(12, 20)
	textColor := getRandomElement([]string{"white", "black", "yellow"})

	mobileWidth := 176
	mobileHeight := 144
	mobileRate := 15

	var videoFilter string

	switch effect {
	case "simple":

		videoFilter = fmt.Sprintf("color=%s:duration=%d:size=%dx%d:rate=%d,drawtext=text='%s %s':fontcolor=%s:fontsize=%d:x=(w-text_w)/2:y=(h-text_h)/2",
			color, vg.Duration, mobileWidth, mobileHeight, mobileRate, text, vg.uniqueID[:6], textColor, fontSize)

	case "ticker":

		videoFilter = fmt.Sprintf("color=%s:duration=%d:size=%dx%d:rate=%d,drawtext=text='%s %s':fontcolor=%s:fontsize=%d:x=w-mod(t*30\\,w+tw):y=(h-text_h)/2",
			color, vg.Duration, mobileWidth, mobileHeight, mobileRate, text, vg.uniqueID[:6], textColor, fontSize)

	case "flash":

		videoFilter = fmt.Sprintf("color=%s:duration=%d:size=%dx%d:rate=%d,drawtext=text='%s %s':fontcolor=%s:fontsize=%d:x=(w-text_w)/2:y=(h-text_h)/2:alpha='if(mod(floor(t*2)\\,2)\\,1\\,0.3)'",
			color, vg.Duration, mobileWidth, mobileHeight, mobileRate, text, vg.uniqueID[:6], textColor, fontSize)

	case "pulse":

		videoFilter = fmt.Sprintf("color=%s:duration=%d:size=%dx%d:rate=%d,drawtext=text='%s %s':fontcolor=%s:fontsize=%d+sin(t*3)*5:x=(w-text_w)/2:y=(h-text_h)/2",
			color, vg.Duration, mobileWidth, mobileHeight, mobileRate, text, vg.uniqueID[:6], textColor, fontSize)

	case "zoom":

		videoFilter = fmt.Sprintf("color=%s:duration=%d:size=%dx%d:rate=%d,drawtext=text='%s %s':fontcolor=%s:fontsize=%d:x=(w-text_w)/2:y=(h-text_h)/2",
			color, vg.Duration, mobileWidth, mobileHeight, mobileRate, text, vg.uniqueID[:6], textColor, fontSize)
		videoFilter += ",scale=iw*(1+sin(t*0.5)*0.2):ih*(1+sin(t*0.5)*0.2)"
	}

	if getRandomInt(1, 4) == 1 {
		videoFilter += fmt.Sprintf(",eq=brightness=%f", float64(getRandomInt(-20, 20))/100.0)
	}

	videoFilter += ",drawtext=text='%{pts\\:hms}':fontcolor=white@0.8:fontsize=8:x=2:y=2"

	args := []string{
		"-f", "lavfi",
		"-i", videoFilter,
		"-f", "lavfi",
		"-i", fmt.Sprintf("sine=frequency=%d:duration=%d", getRandomInt(500, 1200), vg.Duration),
		"-c:v", "libx264",
		"-c:a", "aac",
		"-s", "176x144",
		"-r", "15",
		"-b:v", "64k",
		"-b:a", "32k",
		"-pix_fmt", "yuv420p",
		"-profile:v", "baseline",
		"-level", "1.3",
		"-shortest",
		"-y", outputPath,
	}

	return vg.runFFmpeg(args, outputPath, "3GP")
}

func (vg *VideoGenerator) runFFmpeg(args []string, outputPath, format string) (string, error) {
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Creating %s: %s\n", format, filepath.Base(outputPath))

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		return "", fmt.Errorf("ffmpeg execution failed: %w", err)
	}

	fmt.Printf("✓ %s created in %v: %s\n", format, duration, outputPath)
	return outputPath, nil
}

func (vg *VideoGenerator) SetDimensions(width, height int) {
	vg.Width = width
	vg.Height = height
}

func (vg *VideoGenerator) SetDuration(seconds int) {
	vg.Duration = seconds
}

func (vg *VideoGenerator) SetFramerate(fps int) {
	vg.Framerate = fps
}

func (vg *VideoGenerator) GetUniqueID() string {
	return vg.uniqueID
}
