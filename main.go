package main

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"syscall/js"
	"time"
)

func main() {
	fmt.Println("Go Web Assembly Is Now Running!")
	js.Global().Set("genPerlin", wasmGetPerlin())

	//Indefinite Wait
	<-make(chan bool)
}

func wasmGetPerlin() js.Func {
	perlFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 2 {
			return "Invalid no of arguments"
		}
		grid, err1 := strconv.Atoi(args[0].String())
		res, err2 := strconv.Atoi(args[1].String())

		if grid == 0 || res == 0 {
			return "Cannot accept zero"
		}

		if err1 != nil || err2 != nil {
			return "Couldn't convert to numbers"
		}

		generateRandomGradients(grid, grid)
		generateDepthValues(res)

		jsDoc := js.Global().Get("document")
		if !jsDoc.Truthy() {
			return "Unable to get document object"
		}

		width, height := float64(350), float64(200) //I have no idea why I can't get these normally, but fuck it whatever
		totalWidthHeight := float64(grid * res)

		widthStep := width / totalWidthHeight
		heightStep := height / totalWidthHeight

		canv := jsDoc.Call("getElementById", "window")
		if !canv.Truthy() {
			return "Unable to get canvas"
		}

		ctx := canv.Call("getContext", "2d")
		if !ctx.Truthy() {
			return "Unable to get context"
		}

		//BTW, the only reason I'm drawing from within here is because I can't send an array through for SOME FUCKING REASON
		for iX := 0; iX < len(landscape); iX++ {
			for iY := 0; iY < len(landscape[0]); iY++ {
				val := landscape[iX][iY]
				/*val += 1
				val /= 2 //now scaled from 0 -> 1

				r := val
				g := -(math.Pow(2*val, 2)) + 4*val
				b := 1 - val*/

				if val > 0.85 {
					ctx.Set("fillStyle", "#FF0000")
				} else if val > 0.5 {
					ctx.Set("fillStyle", "#AAFF00")
				} else if val > 0 {
					ctx.Set("fillStyle", "#00FF00")
				} else if val > -0.5 {
					ctx.Set("fillStyle", "#00AAFF")
				} else {
					ctx.Set("fillStyle", "#0000FF")
				}

				ctx.Call("fillRect", float64(iX)*widthStep, float64(iY)*heightStep, widthStep, heightStep)
			}
		}

		return "Generate New"
	})

	return perlFunc
}

func valToHex(v float64) (out string) {
	let := [6]string{
		"A",
		"B",
		"C",
		"D",
		"E",
		"F",
	}
	var nVal int = int(v * 255.0)

	secDig := 0

	for nVal >= 16 {
		secDig++
		nVal -= 16
	}

	if secDig > 9 {
		out = let[secDig-10]
	} else {
		out = strconv.Itoa(secDig)
	}

	if nVal > 9 {
		out += let[nVal-10]
	} else {
		out += strconv.Itoa(nVal)
	}

	//fmt.Println(out)

	return
}

func outputToConsole() {
	for iX := 0; iX < len(landscape); iX++ {
		for iY := 0; iY < len(landscape[0]); iY++ {
			val := landscape[iX][iY]
			if val >= 0.85 {
				fmt.Print("# ")
			} else if val >= 0.5 {
				fmt.Print("X ")
			} else if val >= 0 {
				fmt.Print("+ ")
			} else if val >= -0.5 {
				fmt.Print("- ")
			} else {
				fmt.Print("  ")
			}
		}

		fmt.Println()
	}
}

// Resolution (Rez x Rez) represents the amount pixels/characters per square in grid
func generateDepthValues(resolution int) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	//Subtract 1 to avoid trying to create landscape on edge (where there are no gradient vectors)
	length := (len(gradient_vectors) - 1) * resolution
	height := (len(gradient_vectors[0]) - 1) * resolution

	landscape = nil

	for x := 0; x < length; x++ {
		mu.Lock()
		landscape = append(landscape, []float64{})
		mu.Unlock()

		//There ain't no way I ain't making this multi-threaded
		wg.Add(1)
		go func(cX, h int) {
			defer wg.Done()
			gridX := (cX - (cX % resolution)) / resolution

			for y := 0; y < h; y++ {
				gridY := (y - (y % resolution)) / resolution

				//Get Corresponding Gradient Vectors
				Gradients := []Vector2{
					gradient_vectors[gridX][gridY],
					gradient_vectors[gridX+1][gridY],
					gradient_vectors[gridX][gridY+1],
					gradient_vectors[gridX+1][gridY+1],
				}

				GradientPositions := []Vector2{
					{float64(gridX), float64(gridY)},
					{float64(gridX + 1), float64(gridY)},
					{float64(gridX), float64(gridY + 1)},
					{float64(gridX + 1), float64(gridY + 1)},
				}

				//Calculate "Theoretical Position"
				pos := Vector2{
					float64(gridX) + ((float64((cX % resolution)) + 0.5) / float64(resolution)),
					float64(gridY) + ((float64((y % resolution)) + 0.5) / float64(resolution)),
				}

				//Calculate the Distance Vectors
				DistanceVectors := []Vector2{
					pos.Sub(GradientPositions[0]).Unit(),
					pos.Sub(GradientPositions[1]).Unit(),
					pos.Sub(GradientPositions[2]).Unit(),
					pos.Sub(GradientPositions[3]).Unit(),
				}

				//Calculate the Dot Products
				Dots := []float64{
					DistanceVectors[0].Dot(Gradients[0]),
					DistanceVectors[1].Dot(Gradients[1]),
					DistanceVectors[2].Dot(Gradients[2]),
					DistanceVectors[3].Dot(Gradients[3]),
				}

				//Interpolate between the two.............. yeah this is fun, could've made this a function but I decided to copy my old code
				ab := Dots[0] + ((pos.x-GradientPositions[0].x)/GradientPositions[1].Distance(GradientPositions[0]))*(Dots[1]-Dots[0])
				cd := Dots[2] + ((pos.x-GradientPositions[2].x)/GradientPositions[3].Distance(GradientPositions[2]))*(Dots[3]-Dots[2])
				finalValue := ab + ((pos.y-GradientPositions[0].y)/GradientPositions[2].Distance(GradientPositions[0]))*(cd-ab)
				if math.IsNaN(finalValue) {
					finalValue = 0
				}

				mu.Lock()
				landscape[cX] = append(landscape[cX], finalValue)
				mu.Unlock()
			}
		}(x, height)
	}

	wg.Wait()
}

/*func outputGradients(x, y int) {
	for iX := 0; iX < x; iX++ {
		for iY := 0; iY < y; iY++ {
			vec := gradient_vectors[iX][iY]

			fmt.Print("(" + fmt.Sprintf("%.2f", vec.x) + ", " + fmt.Sprintf("%.2f", vec.y) + ") ")
		}

		fmt.Println()
	}
}*/

func generateRandomGradients(x, y int) {
	rand.Seed(int64(time.Now().Unix()))

	var wg sync.WaitGroup
	var mu sync.Mutex

	//Quickly Generate the gradients
	gradient_vectors = nil
	wg.Add(x)
	for iX := 0; iX < x; iX++ {
		mu.Lock()
		gradient_vectors = append(gradient_vectors, []Vector2{})
		mu.Unlock()

		go func(loc, lim int, w *sync.WaitGroup) { //Instead of Sequentially generating, we generate in parallel
			defer w.Done()

			for iY := 0; iY < lim; iY++ {
				mu.Lock()
				val := &Vector2{rand.Float64()*2 - 1, rand.Float64()*2 - 1}
				val.PUnit()

				gradient_vectors[loc] = append(gradient_vectors[loc], *val) //No race conditions since this will generate
				mu.Unlock()
			}

		}(iX, y, &wg)
	}

	wg.Wait()
}
