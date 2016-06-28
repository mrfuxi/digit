package gridgen

import (
	"image"
	"math"
	"math/rand"

	"github.com/llgcode/draw2d/draw2dimg"
)

type FragmentType uint8

const (
	FragmentTypeEmpty FragmentType = iota
	FragmentTypeCornerNW
	FragmentTypeCornerNE
	FragmentTypeCornerSE
	FragmentTypeCornerSW
	FragmentTypeEdgeN
	FragmentTypeEdgeE
	FragmentTypeEdgeS
	FragmentTypeEdgeW
	FragmentTypeCross
)

type FragmentSuperType uint8

const (
	FragmentSuperTypeEmpty FragmentSuperType = iota
	FragmentSuperTypeCorner
	FragmentSuperTypeEdge
	FragmentSuperTypeCross
)

func IsCorner(fragment FragmentType) bool {
	switch fragment {
	case FragmentTypeCornerNW:
		return true
	case FragmentTypeCornerNE:
		return true
	case FragmentTypeCornerSE:
		return true
	case FragmentTypeCornerSW:
		return true
	}
	return false
}

func IsEdge(fragment FragmentType) bool {
	switch fragment {
	case FragmentTypeEdgeN:
		return true
	case FragmentTypeEdgeE:
		return true
	case FragmentTypeEdgeS:
		return true
	case FragmentTypeEdgeW:
		return true
	}
	return false
}

func IsCross(fragment FragmentType) bool {
	return fragment == FragmentTypeCross
}

func IsEmpty(fragment FragmentType) bool {
	if fragment == FragmentTypeEmpty {
		return true
	}

	return false
}

func FragmentTypeToSuper(fragment FragmentType) FragmentSuperType {
	switch {
	case IsCorner(fragment):
		return FragmentSuperTypeCorner
	case IsEdge(fragment):
		return FragmentSuperTypeEdge
	case IsCross(fragment):
		return FragmentSuperTypeCross
	case IsEmpty(fragment):
		return FragmentSuperTypeEmpty
	default:
		panic("What is this?")
	}
}

type Drawer interface {
	Draw(images chan<- Image)
	Count() int
}

type contextDrawFunc func(gc *draw2dimg.GraphicContext)

func drawBase(dr contextDrawFunc) image.Image {
	center := ImageSize / 2.0

	canvas := image.NewRGBA(image.Rect(0, 0, ImageSize, ImageSize))

	gc := draw2dimg.NewGraphicContext(canvas)

	gc.DrawImage(image.Black)      // Background color
	gc.SetStrokeColor(image.White) // Line color
	gc.SetLineWidth(2)

	gc.Translate(center, center)

	dr(gc)

	return canvas
}

type cornerDrawer struct {
	Movements []float64
	Angles    []float64
}

func (c *cornerDrawer) Draw(images chan<- Image) {
	for _, fragment := range []FragmentType{FragmentTypeCornerNW, FragmentTypeCornerNE, FragmentTypeCornerSE, FragmentTypeCornerSW} {
		for _, ds := range c.Angles {
			for _, dd := range c.Angles {
				for _, dx := range c.Movements {
					for _, dy := range c.Movements {
						images <- Image{
							GridInfo: GridInfo{
								Fragment:      fragment,
								FragmentSuper: FragmentTypeToSuper(fragment),
								Train:         rand.Intn(100) >= 5,
							},
							Image: drawBase(c.drawFragment(fragment, dx, dy, ds, dd)),
						}
					}
				}
			}
		}
	}
}

func (c *cornerDrawer) Count() int {
	return 4 * len(c.Angles) * len(c.Angles) * len(c.Movements) * len(c.Movements)
}

func (c *cornerDrawer) drawFragment(fragment FragmentType, dx, dy, dStartAngle, dDiffAngle float64) contextDrawFunc {
	return func(gc *draw2dimg.GraphicContext) {
		var startAngle float64
		var diffAngle float64 = 90

		switch fragment {
		case FragmentTypeCornerNW:
			startAngle = 0
		case FragmentTypeCornerNE:
			startAngle = 90
		case FragmentTypeCornerSE:
			startAngle = 180
		case FragmentTypeCornerSW:
			startAngle = 270
		default:
			panic("invalid type")
		}

		startAngle += dStartAngle
		diffAngle += dDiffAngle

		gc.Translate(dx, dy)
		gc.Rotate(startAngle * math.Pi / 180.0)

		gc.MoveTo(0, 0)
		gc.LineTo(ImageSize, 0)
		gc.Close()
		gc.FillStroke()

		gc.Rotate(diffAngle * math.Pi / 180.0)

		gc.MoveTo(0, 0)
		gc.LineTo(ImageSize, 0)
		gc.Close()
		gc.FillStroke()
	}
}

type edgeDrawer struct {
	Movements []float64
	Angles    []float64
}

func (e *edgeDrawer) Draw(images chan<- Image) {
	for _, fragment := range []FragmentType{FragmentTypeEdgeN, FragmentTypeEdgeE, FragmentTypeEdgeS, FragmentTypeEdgeW} {
		for _, ds := range e.Angles {
			for _, dd := range e.Angles {
				for _, dx := range e.Movements {
					for _, dy := range e.Movements {
						images <- Image{
							GridInfo: GridInfo{
								Fragment:      fragment,
								FragmentSuper: FragmentTypeToSuper(fragment),
								Train:         rand.Intn(100) >= 5,
							},
							Image: drawBase(e.drawFragment(fragment, dx, dy, ds, dd)),
						}
					}
				}
			}
		}
	}
}

func (e *edgeDrawer) Count() int {
	return 4 * len(e.Angles) * len(e.Angles) * len(e.Movements) * len(e.Movements)
}

func (e *edgeDrawer) drawFragment(fragment FragmentType, dx, dy, dStartAngle, dDiffAngle float64) contextDrawFunc {
	return func(gc *draw2dimg.GraphicContext) {
		var startAngle float64
		var diffAngle float64 = 90

		switch fragment {
		case FragmentTypeEdgeN:
			startAngle = 0
		case FragmentTypeEdgeE:
			startAngle = 90
		case FragmentTypeEdgeS:
			startAngle = 180
		case FragmentTypeEdgeW:
			startAngle = 270
		default:
			panic("invalid type")
		}

		startAngle += dStartAngle
		diffAngle += dDiffAngle

		gc.Translate(dx, dy)
		gc.Rotate(startAngle * math.Pi / 180.0)

		gc.MoveTo(-ImageSize, 0)
		gc.LineTo(ImageSize, 0)
		gc.Close()
		gc.FillStroke()

		gc.Rotate(diffAngle * math.Pi / 180.0)

		gc.MoveTo(0, 0)
		gc.LineTo(ImageSize, 0)
		gc.Close()
		gc.FillStroke()
	}
}

type crossDrawer struct {
	Movements []float64
	Angles    []float64
}

func (c *crossDrawer) Draw(images chan<- Image) {
	fragment := FragmentTypeCross
	for _, ds := range c.Angles {
		for _, dd := range c.Angles {
			for _, dx := range c.Movements {
				for _, dy := range c.Movements {
					images <- Image{
						GridInfo: GridInfo{
							Fragment:      fragment,
							FragmentSuper: FragmentTypeToSuper(fragment),
							Train:         rand.Intn(100) >= 5,
						},
						Image: drawBase(c.drawFragment(dx, dy, ds, dd)),
					}
				}
			}
		}
	}
}

func (c *crossDrawer) Count() int {
	return len(c.Angles) * len(c.Angles) * len(c.Movements) * len(c.Movements)
}

func (c *crossDrawer) drawFragment(dx, dy, dStartAngle, dDiffAngle float64) contextDrawFunc {
	return func(gc *draw2dimg.GraphicContext) {
		var startAngle float64
		var diffAngle float64 = 90

		startAngle += dStartAngle
		diffAngle += dDiffAngle

		gc.Translate(dx, dy)
		gc.Rotate(startAngle * math.Pi / 180.0)

		gc.MoveTo(-ImageSize, 0)
		gc.LineTo(ImageSize, 0)
		gc.Close()
		gc.FillStroke()

		gc.Rotate(diffAngle * math.Pi / 180.0)

		gc.MoveTo(-ImageSize, 0)
		gc.LineTo(ImageSize, 0)
		gc.Close()
		gc.FillStroke()
	}
}

type lineDrawer struct {
	Movements []float64
	Angles    []float64
}

func (l *lineDrawer) Draw(images chan<- Image) {
	for _, horizontal := range []bool{true, false} {
		for _, ds := range l.Angles {
			for _, move := range l.Movements {
				images <- Image{
					GridInfo: GridInfo{
						Fragment:      FragmentTypeEmpty,
						FragmentSuper: FragmentTypeToSuper(FragmentTypeEmpty),
						Train:         rand.Intn(100) >= 5,
					},
					Image: drawBase(l.drawFragment(horizontal, move, ds)),
				}
			}
		}
	}
}

func (l *lineDrawer) Count() int {
	return 2 * len(l.Angles) * len(l.Movements)
}

func (l *lineDrawer) drawFragment(horizontal bool, move, dStartAngle float64) contextDrawFunc {
	return func(gc *draw2dimg.GraphicContext) {
		var startAngle float64
		var dx float64
		var dy float64

		if horizontal {
			startAngle = 0
			dy = move
		} else {
			startAngle = 90
			dx = move
		}

		startAngle += dStartAngle

		gc.Translate(dx, dy)
		gc.Rotate(startAngle * math.Pi / 180.0)

		gc.MoveTo(-ImageSize, 0)
		gc.LineTo(ImageSize, 0)
		gc.Close()
		gc.FillStroke()
	}
}

type emptyDrawer struct {
	Samples int
	Noise   float64
}

func (e *emptyDrawer) Draw(images chan<- Image) {
	fragment := FragmentTypeEmpty

	images <- Image{
		GridInfo: GridInfo{
			Fragment:      fragment,
			FragmentSuper: FragmentTypeToSuper(fragment),
			Train:         true,
		},
		Image: drawBase(e.drawFragment(0)),
	}

	for i := 0; i < e.Samples; i++ {
		images <- Image{
			GridInfo: GridInfo{
				Fragment:      fragment,
				FragmentSuper: FragmentTypeToSuper(fragment),
				Train:         rand.Intn(100) >= 5,
			},
			Image: drawBase(e.drawFragment(e.Noise)),
		}
	}
}

func (e *emptyDrawer) Count() int {
	return e.Samples
}

func (e *emptyDrawer) drawFragment(noise float64) contextDrawFunc {
	return func(gc *draw2dimg.GraphicContext) {
		for i := 0; i < int(float64(ImageSize*ImageSize)*noise); i++ {
			x := (rand.Float64() - 0.5) * ImageSize
			y := (rand.Float64() - 0.5) * ImageSize

			gc.MoveTo(x, y)
			gc.LineTo(x+1, y)
			gc.Close()
			gc.FillStroke()
		}
	}
}

type incompleteEdgeDrawer struct {
	Movements []float64
	Angles    []float64
}

func (i *incompleteEdgeDrawer) Draw(images chan<- Image) {
	dOff := 4.0
	for _, fragment := range []FragmentType{FragmentTypeEdgeN, FragmentTypeEdgeE, FragmentTypeEdgeS, FragmentTypeEdgeW} {
		for _, ds := range i.Angles {
			for _, dd := range i.Angles {
				for _, dx := range i.Movements {
					for _, dy := range i.Movements {
						fr := FragmentTypeEmpty

						images <- Image{
							GridInfo: GridInfo{
								Fragment:      fr,
								FragmentSuper: FragmentTypeToSuper(fr),
								Train:         rand.Intn(100) >= 5,
							},
							Image: drawBase(i.drawFragment(fragment, dx, dy, dOff, ds, dd)),
						}

					}
				}
			}
		}
	}
}

func (i *incompleteEdgeDrawer) Count() int {
	return 4 * len(i.Angles) * len(i.Angles) * len(i.Movements) * len(i.Movements)
}

func (i *incompleteEdgeDrawer) drawFragment(fragment FragmentType, dx, dy, dOff, dStartAngle, dDiffAngle float64) contextDrawFunc {
	return func(gc *draw2dimg.GraphicContext) {
		var startAngle float64
		var diffAngle float64 = 90

		switch fragment {
		case FragmentTypeEdgeN:
			startAngle = 0
		case FragmentTypeEdgeE:
			startAngle = 90
		case FragmentTypeEdgeS:
			startAngle = 180
		case FragmentTypeEdgeW:
			startAngle = 270
		default:
			panic("invalid type")
		}

		startAngle += dStartAngle
		diffAngle += dDiffAngle

		gc.Translate(dx, dy)
		gc.Rotate(startAngle * math.Pi / 180.0)

		gc.MoveTo(-ImageSize, 0)
		gc.LineTo(ImageSize, 0)
		gc.Close()
		gc.FillStroke()

		gc.Rotate(diffAngle * math.Pi / 180.0)

		gc.MoveTo(dOff, 0)
		gc.LineTo(ImageSize, 0)
		gc.Close()
		gc.FillStroke()
	}
}
