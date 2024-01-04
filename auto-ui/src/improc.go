package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"gocv.io/x/gocv"
)

func getRefResolution(imgPath string) (int, int, error) {

	filename := strings.Split(filepath.Base(imgPath), ".")[0]

	resolution := strings.Split(filename, "_")

	fmt.Println(resolution)

	tmp := strings.Split(resolution[1], "x")

	width, err := strconv.Atoi(tmp[0])

	if err != nil {

		return 0, 0, fmt.Errorf("getRefResolution:%v", err)
	}

	height, err := strconv.Atoi(tmp[1])

	if err != nil {

		return 0, 0, fmt.Errorf("getRefResolution:%v", err)
	}

	return width, height, nil

}

func getImgScale(img *gocv.Mat, tmpImgPath string) (float32, float32, error) {

	refWidth, refHeight, err := getRefResolution(tmpImgPath)

	if err != nil {
		return 0, 0, fmt.Errorf("getImgScale:%v", err)
	}

	imgWidth := img.Cols()
	imgHeight := img.Rows()

	aspectDiff := (float32(refWidth) / float32(refHeight)) / (float32(img.Cols()) / float32(img.Rows()))
	InfoLogger.Printf("AspectDiff:%v", aspectDiff)
	widthDiff := (float32(refWidth-imgWidth) / float32(refWidth))
	heightDiff := (float32(refHeight-imgHeight) / float32(refHeight))

	InfoLogger.Printf("widthDiff:%v heightDiff:%v", widthDiff, heightDiff)

	return widthDiff * aspectDiff, heightDiff * aspectDiff, nil

}

func imageChanged(src, ref string, options ...int) bool {

	if !pathExists(src) {
		ErrorLogger.Printf("Path %v does not exist", src)
		return false
	}

	if !pathExists(ref) {
		ErrorLogger.Printf("Path %v does not exist", ref)
		return false
	}

	imgSrc := gocv.IMRead(src, gocv.IMReadGrayScale)

	if imgSrc.Empty() {
		ErrorLogger.Printf("%v is empty", src)
		return false
	}

	var threshold int
	defer imgSrc.Close()

	imgRef := gocv.IMRead(ref, gocv.IMReadGrayScale)

	if imgRef.Empty() {
		ErrorLogger.Printf("%v is empty", ref)
		return false

	}

	diff := gocv.NewMat()
	defer diff.Close()

	gocv.AbsDiff(imgSrc, imgRef, &diff)

	InfoLogger.Printf("Zeros:%v", gocv.CountNonZero(diff))

	if len(options) > 0 {
		threshold = options[0]
	} else {
		threshold = 100
	}

	if gocv.CountNonZero(diff) >= threshold {
		return true
	}

	return false

}
func getXY(src, ref string, options ...float32) (int, int, error) {

	if !pathExists(src) {
		ErrorLogger.Printf("Path %s does not exist", src)
		return -1, -1, fmt.Errorf("Path %s does not exist\n", src)
	}

	if !pathExists(ref) {
		ErrorLogger.Printf("Path %s does not exist", ref)
		return -1, -1, fmt.Errorf("Path %s does not exist\n", ref)
	}

	var threshold float32

	imgSrc := gocv.IMRead(src, gocv.IMReadGrayScale)

	if imgSrc.Empty() {
		ErrorLogger.Printf("Cannot read image at %s", src)
		return -1, -1, fmt.Errorf("Cannot read image at %s\n", src)
	}

	defer imgSrc.Close()

	imgRef := gocv.IMRead(ref, gocv.IMReadGrayScale)

	if imgRef.Empty() {
		ErrorLogger.Printf("Cannot read image at %s", ref)
		return -1, -1, fmt.Errorf("Cannot read image at %s", ref)
	}

	defer imgRef.Close()

	result := gocv.NewMat()
	defer result.Close()

	mask := gocv.NewMat()
	defer mask.Close()

	srcWidth := imgSrc.Cols()
	srcHeight := imgSrc.Rows()

	/*
		scaleWidth, scaleHeight, err := getImgScale(&imgRef, src)

		fmt.Printf("x:%f y:%f\n", scaleWidth, scaleHeight)

		if err != nil {
			return -1, -1, fmt.Errorf("getXY:%v", err)
		}

		if int(scaleWidth) == 100 {

			gocv.IMWrite(filepath.Join(refImageDirectory, fmt.Sprintf("before_size_%d_%d.png", srcWidth, srcHeight)), imgSrc)
			gocv.Resize(imgSrc, &imgSrc, image.Point{0, 0}, float64(scaleWidth), float64(scaleHeight), gocv.InterpolationArea)
			gocv.IMWrite(filepath.Join(refImageDirectory, fmt.Sprintf("resized_%d_%d.png", imgSrc.Cols(), imgSrc.Rows())), imgSrc)
			gocv.IMWrite(filepath.Join(refImageDirectory, "ref_gray.png"), imgRef)
		}

	*/

	InfoLogger.Printf("Current Resolution: Width:%d Height:%d", imgRef.Cols(), imgRef.Rows())

	gocv.MatchTemplate(imgSrc, imgRef, &result, gocv.TmCcoeffNormed, mask)

	_, maxValue, _, maxCoord := gocv.MinMaxLoc(result)

	InfoLogger.Printf("Threshold:%f", maxValue)

	if len(options) > 0 {
		threshold = options[0]
	} else {
		threshold = defaultThreshold
	}

	if maxValue < threshold {
		ErrorLogger.Printf("Image %s not found in reference image %s", src, ref)
		return -1, -1, fmt.Errorf("Image %s not found in reference image %s\n", src, ref)
	}

	centerX := maxCoord.X + srcWidth/2
	centerY := maxCoord.Y + srcHeight/2

	return centerX, centerY, nil

}
