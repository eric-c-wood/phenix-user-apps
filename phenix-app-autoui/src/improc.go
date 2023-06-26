package main

import (
	"fmt"

	"gocv.io/x/gocv"
)

func imageChanged(src, ref string) bool {

	if !pathExists(src) {
		fmt.Printf("Path %v does not exist\n", src)
		return false
	}

	if !pathExists(ref) {
		fmt.Printf("Path %v does not exist\n", ref)
		return false
	}

	imgSrc := gocv.IMRead(src, gocv.IMReadGrayScale)

	if imgSrc.Empty() {
		fmt.Printf("%v is empty", src)
		return false
	}

	defer imgSrc.Close()

	imgRef := gocv.IMRead(ref, gocv.IMReadGrayScale)

	if imgRef.Empty() {
		fmt.Printf("%v is empty", ref)
		return false

	}

	diff := gocv.NewMat()
	defer diff.Close()

	gocv.AbsDiff(imgSrc, imgRef, &diff)

	fmt.Printf("Zeros:%v\n", gocv.CountNonZero(diff))

	if gocv.CountNonZero(diff) > 0 {
		return true
	}

	return false

}
func getXY(src, ref string) (int, int, error) {

	if !pathExists(src) {
		return -1, -1, fmt.Errorf("Path %s does not exist", src)
	}

	if !pathExists(ref) {
		return -1, -1, fmt.Errorf("Path %s does not exist", ref)
	}

	imgSrc := gocv.IMRead(src, gocv.IMReadGrayScale)

	if imgSrc.Empty() {
		return -1, -1, fmt.Errorf("Cannot read image at %s", src)
	}

	defer imgSrc.Close()

	imgRef := gocv.IMRead(ref, gocv.IMReadGrayScale)

	if imgRef.Empty() {
		return -1, -1, fmt.Errorf("Cannot read image at %s", ref)
	}

	defer imgRef.Close()

	result := gocv.NewMat()
	defer result.Close()

	mask := gocv.NewMat()
	defer mask.Close()

	fmt.Printf("Path:%s Source:w:%d:h:%d:r:%f\n", src, imgSrc.Cols(), imgSrc.Rows(), float32(imgSrc.Cols())/float32(imgRef.Cols()))
	fmt.Printf("Path:%s Ref:w:%d:h:%d:r:%f\n", ref, imgRef.Cols(), imgRef.Rows(), float32(imgSrc.Rows())/float32(imgRef.Rows()))

	gocv.MatchTemplate(imgSrc, imgRef, &result, gocv.TmCcoeffNormed, mask)

	_, maxValue, _, maxCoord := gocv.MinMaxLoc(result)

	if maxValue < 0.8 {
		return -1, -1, fmt.Errorf("Image %s not found in reference image %s\n", src, ref)
	}

	srcWidth := imgSrc.Cols()
	srcHeight := imgSrc.Rows()

	centerX := maxCoord.X + srcWidth/2
	centerY := maxCoord.Y + srcHeight/2

	return centerX, centerY, nil

}
