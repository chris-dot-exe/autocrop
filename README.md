# autocrop

[![GoDoc](https://godoc.org/github.com/mandykoh/autocrop?status.svg)](https://godoc.org/github.com/mandykoh/autocrop)
[![Go Report Card](https://goreportcard.com/badge/github.com/mandykoh/autocrop)](https://goreportcard.com/report/github.com/mandykoh/autocrop)

`autocrop` is an automatic image cropping/region-of-interest finding implementation in Go.

`autocrop` uses an energy-based approach to determine regions of visual interest at which to stop cropping. Cropping is alpha channel aware and supports images with translucency.

`autocrop` currently assumes image data is sRGB encoded.

See the [API documentation](https://godoc.org/github.com/mandykoh/autocrop) for more details.

Some examples of cropping decisions made by `autocrop` can be seen here:

_Threshold 0.3_

![Example of textured background being cropped from around a pink square](screenshots/crop-example-1.png)

_Threshold 0.1_

![Example of space being cropped away around a character illustration](screenshots/crop-example-2.png)

_Threshold 0.15_

![Example of sky being cropped away around a bridge](screenshots/crop-example-3.png)


This software is made available under an [MIT license](LICENSE).


## Example usage

Find the autocropped bounds of an image using [`BoundsForThreshold`](https://godoc.org/github.com/mandykoh/autocrop#BoundsForThreshold):

```go
croppedBounds := autocrop.BoundsForThreshold(img, energyThreshold)
```

The `energyThreshold` is a value between 0.0 and 1.0 which determines at what energy level cropping stops, relative to the maximum energy of the image. Higher threshold values result in more aggressive cropping.

The [`ToThreshold`](https://godoc.org/github.com/mandykoh/autocrop#ToThreshold) function can be used to conveniently get a cropped image if you don’t need the bounds:

```go
croppedImg := autocrop.ToThreshold(img, energyThreshold)
```

## New
Option to add margin to the cropped image.
Margin follows the same rules as CSS's margin. 

* When one value is specified, it applies the same margin to all four sides.
* When two values are specified, the first margin applies to the top and bottom, the second to the left and right.
* When three values are specified, the first margin applies to the top, the second to the right and left, the third to the bottom.
* When four values are specified, the margins apply to the top, right, bottom, and left in that order (clockwise).
* 
```go
croppedImg := autocrop.ToThresholdWithMargin(img, energyThreshold, 10)
croppedImg =  autocrop.ToThresholdWithMargin(img, energyThreshold, 10, 5)
croppedImg =  autocrop.ToThresholdWithMargin(img, energyThreshold, 10, 5, 4)
croppedImg =  autocrop.ToThresholdWithMargin(img, energyThreshold, 10, 5, 4, 2)
```

## Command line tool

`autocrop` also provides a command line tool that can be installed as follows:

```
$ go install github.com/chris-dot-exe/autocrop/cmd/autocrop
```

and used to crop images like this:

```
$ autocrop <input_image> <output_png> [threshold]
```

The threshold is optional and defaults to 0.1.
