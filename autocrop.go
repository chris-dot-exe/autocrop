package autocrop

import (
	"github.com/mandykoh/prism/linear"
	"github.com/mandykoh/prism/srgb"
	"image"
	"image/draw"
	"log"
	"math"
)

const bufferPixels = 2

type Margin struct {
	Top    int
	Right  int
	Left   int
	Bottom int
}

// BoundsForThreshold returns the bounds for a crop of the specified image up to
// the given energy threshold.
//
// energyThreshold is a value between 0.0 and 1.0 representing the maximum
// energy to allow to be cropped away before stopping, relative to the maximum
// energy of the image.
func BoundsForThreshold(img *image.NRGBA, energyThreshold float32) image.Rectangle {
	rect, _ := boundsForThreshold(img, energyThreshold, false, nil)
	return rect
}

func BoundsForThresholdWithMargin(img *image.NRGBA, energyThreshold float32, extend bool, margins ...int) (image.Rectangle, srgb.Color) {

	margin := getMargin(margins)
	log.Printf("%+v\n", margin)

	return boundsForThreshold(img, energyThreshold, extend, margin)
}

func boundsForThreshold(img *image.NRGBA, energyThreshold float32, extend bool, margin *Margin) (image.Rectangle, srgb.Color) {

	crop := img.Bounds()

	energyCrop := crop
	energyCrop.Min.X++
	energyCrop.Min.Y++
	energyCrop.Max.X--
	energyCrop.Max.Y--

	if energyCrop.Empty() {
		return img.Bounds(), srgb.Color{
			RGB: linear.RGB{
				R: 255,
				G: 255,
				B: 255,
			},
		}
	}

	colEnergies, rowEnergies, dominantColor := Energies(img, energyCrop)

	// Find left and right high energy jumps
	maxEnergy := findMaxEnergy(colEnergies)
	cropLeft := findFirstEnergyBound(colEnergies, maxEnergy, energyThreshold)
	cropRight := findLastEnergyBound(colEnergies, maxEnergy, energyThreshold, cropLeft)

	// Find top and bottom high energy jumps
	maxEnergy = findMaxEnergy(rowEnergies)
	cropTop := findFirstEnergyBound(rowEnergies, maxEnergy, energyThreshold)
	cropBottom := findLastEnergyBound(rowEnergies, maxEnergy, energyThreshold, cropTop)

	// Apply the crop
	crop.Min.X += cropLeft
	crop.Min.Y += cropTop
	crop.Max.X -= cropRight
	crop.Max.Y -= cropBottom
	if margin != nil {

		if !extend {
			if (crop.Min.X - margin.Left) >= 0 {
				crop.Min.X -= margin.Left
			} else {
				crop.Min.X = 0
			}
			if (crop.Min.Y - margin.Top) >= 0 {
				crop.Min.Y -= margin.Top
			} else {
				crop.Min.Y = 0
			}
			if (crop.Max.X + margin.Right) <= img.Bounds().Max.X {
				crop.Max.X += margin.Right
			} else {
				crop.Max.X = img.Bounds().Max.X
			}
			if (crop.Max.Y + margin.Bottom) <= img.Bounds().Max.Y {
				crop.Max.Y += margin.Bottom
			} else {
				crop.Max.Y = img.Bounds().Max.Y
			}
		} else {
			crop.Min.X -= margin.Left
			crop.Min.Y -= margin.Top
			crop.Max.X += margin.Right
			crop.Max.Y += margin.Bottom
		}
	}

	return crop, dominantColor
}

// Energies returns the total row and column energies for the specified region
// of an image.
func Energies(img *image.NRGBA, r image.Rectangle) (cols, rows []float32, color2 srgb.Color) {

	// Need 1 pixel more luminance data on each side so that all energy
	// calculations are for pixels with a full set of neighbours.
	luminanceBounds := r
	luminanceBounds.Min.X--
	luminanceBounds.Min.Y--
	luminanceBounds.Max.X++
	luminanceBounds.Max.Y++

	luminances, alphas, dominantColor := luminancesAndAlphas(img, luminanceBounds)

	cols = make([]float32, r.Dx(), r.Dx())
	rows = make([]float32, r.Dy(), r.Dy())
	// Calculate total column and row energies
	for i, row := r.Min.Y, 0; i < r.Max.Y; i, row = i+1, row+1 {
		for j, col := r.Min.X, 0; j < r.Max.X; j, col = j+1, col+1 {
			e := energy(luminances, alphas, luminanceBounds.Dx(), col+1, row+1)
			cols[col] += e
			rows[row] += e
		}
	}

	return cols, rows, dominantColor
}

// ToThreshold returns an image cropped using the bounds provided by
// BoundsForThreshold.
//
// energyThreshold is a value between 0.0 and 1.0 representing the maximum
// energy to allow to be cropped away before stopping, relative to the maximum
// energy of the image.
func ToThreshold(img *image.NRGBA, energyThreshold float32) *image.NRGBA {
	return toThreshold(img, energyThreshold, false)
}

func ToThresholdWithMargin(img *image.NRGBA, energyThreshold float32, extend bool, margins ...int) *image.NRGBA {
	return toThreshold(img, energyThreshold, extend, margins...)
}

func toThreshold(img *image.NRGBA, energyThreshold float32, extend bool, margins ...int) *image.NRGBA {
	crop, dominantColor := BoundsForThresholdWithMargin(img, energyThreshold, extend, margins...)
	resultImg := image.NewNRGBA(image.Rect(0, 0, crop.Dx(), crop.Dy()))
	draw.Draw(resultImg, resultImg.Bounds(), &image.Uniform{C: dominantColor.ToRGBA(1.0)}, image.Point{}, draw.Src)
	draw.Draw(resultImg, resultImg.Bounds(), img, crop.Min, draw.Src)
	return resultImg
}

func colourAt(img *image.NRGBA, x, y int) (col srgb.Color, alpha float32) {
	return srgb.ColorFromNRGBA(img.NRGBAAt(x, y))
}

func energy(luminances, alphas []float32, width int, x, y int) float32 {
	center := y*width + x

	// North west + west + south west - north east - east - south east
	eX := luminances[center-width-1] + luminances[center-1] + luminances[center+width-1] - luminances[center-width+1] - luminances[center+1] - luminances[center+width+1]

	// North west + north + north east - south west - south - south east
	eY := luminances[center-width-1] + luminances[center-width] + luminances[center-width+1] - luminances[center+width-1] - luminances[center+width] - luminances[center+width+1]

	return float32((math.Abs(float64(eX)) + math.Abs(float64(eY))) * float64(alphas[center]))
}

func findFirstEnergyBound(energies []float32, maxEnergy, threshold float32) (bound int) {
	for i := 0; i < len(energies); i++ {
		if energies[i]/maxEnergy > threshold {
			bound = i
			if i > 0 {
				bound += bufferPixels
			}
			break
		}
		bound++
	}

	return bound
}

func findLastEnergyBound(energies []float32, maxEnergy, threshold float32, firstBound int) (bound int) {
	for i := len(energies) - 1; i > firstBound+bufferPixels; i-- {
		if energies[i]/maxEnergy > threshold {
			bound = len(energies) - i - 1
			if i < len(energies)-1 {
				bound += bufferPixels
			}
			break
		}
		bound++
	}

	return bound
}

func findMaxEnergy(energies []float32) float32 {
	max := energies[0]
	for i := 1; i < len(energies); i++ {
		if energies[i] > max {
			max = energies[i]
		}
	}

	return max
}

func luminance(c srgb.Color, alpha float32) float32 {
	return c.Luminance() + alpha
}

func luminancesAndAlphas(img *image.NRGBA, r image.Rectangle) (luminances, alphas []float32, color2 srgb.Color) {

	luminances = make([]float32, r.Dx()*r.Dy(), r.Dx()*r.Dy())
	alphas = make([]float32, r.Dx()*r.Dy(), r.Dx()*r.Dy())

	index := 0

	colorCount := make(map[srgb.Color]int)

	// Get the luminances and alphas for all pixels
	for i := r.Min.Y; i < r.Max.Y; i++ {
		for j := r.Min.X; j < r.Max.X; j++ {
			c, a := colourAt(img, j, i)
			colorCount[c]++
			luminances[index] = luminance(c, a)
			alphas[index] = a
			index++
		}
	}

	// Find the most frequent color
	var dominantColor srgb.Color
	maxCount := 0
	for c, count := range colorCount {
		if count > maxCount {
			dominantColor = c
			maxCount = count
		}
	}

	return luminances, alphas, dominantColor
}

func getMargin(margins []int) *Margin {
	var m *Margin

	switch len(margins) {
	case 0:
		m = nil
	case 1:
		m = &Margin{
			Top:    margins[0],
			Right:  margins[0],
			Left:   margins[0],
			Bottom: margins[0],
		}
	case 2:
		m = &Margin{
			Top:    margins[0],
			Right:  margins[1],
			Left:   margins[1],
			Bottom: margins[0],
		}
	case 3:
		m = &Margin{
			Top:    margins[0],
			Right:  margins[1],
			Left:   margins[1],
			Bottom: margins[2],
		}
	case 4:
		m = &Margin{
			Top:    margins[0],
			Right:  margins[1],
			Bottom: margins[2],
			Left:   margins[3],
		}
	}
	return m
}
