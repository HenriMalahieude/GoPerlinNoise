package main

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

func generateRandomGradients(x, y int, extreme bool) {
	x += 1
	y += 1

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
				val := Vector2{rand.Float64()*2 - 1, rand.Float64()*2 - 1}.Unit()
				scale := 1.0
				if extreme {
					scale = rand.Float64()
				}

				gradient_vectors[loc] = append(gradient_vectors[loc], val.Scale(scale)) //No race conditions since this will generate
				mu.Unlock()
			}

		}(iX, y, &wg)
	}

	wg.Wait()
}

// Resolution (Rez x Rez) represents the amount pixels/characters per square in grid
func generateDepthValues(resolution int, terrain bool) {
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
				var DistanceVectors []Vector2
				if terrain {
					DistanceVectors = []Vector2{
						/*GradientPositions[0].Sub(pos).Unit(),
						GradientPositions[1].Sub(pos).Unit(),
						GradientPositions[2].Sub(pos).Unit(),
						GradientPositions[3].Sub(pos).Unit(),*/
						pos.Sub(GradientPositions[0]).Unit(),
						pos.Sub(GradientPositions[1]).Unit(),
						pos.Sub(GradientPositions[2]).Unit(),
						pos.Sub(GradientPositions[3]).Unit(),
					}
				} else {
					DistanceVectors = []Vector2{
						pos.Sub(GradientPositions[0]),
						pos.Sub(GradientPositions[1]),
						pos.Sub(GradientPositions[2]),
						pos.Sub(GradientPositions[3]),
					}
				}

				//Calculate the Dot Products
				Dots := []float64{
					DistanceVectors[0].Dot(Gradients[0]),
					DistanceVectors[1].Dot(Gradients[1]),
					DistanceVectors[2].Dot(Gradients[2]),
					DistanceVectors[3].Dot(Gradients[3]),
				}

				//Interpolate between the two.............. yeah this is fun, could've made this a function but I decided to copy my old code
				ab := interpolate(Dots[0], Dots[1], (pos.x-GradientPositions[0].x)/(GradientPositions[1].x-GradientPositions[0].x))
				cd := interpolate(Dots[2], Dots[3], (pos.x-GradientPositions[0].x)/(GradientPositions[1].x-GradientPositions[0].x))
				finalValue := interpolate(ab, cd, (pos.y-GradientPositions[0].y)/(GradientPositions[2].y-GradientPositions[0].y))
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

func interpolate(a, b, alpha float64) float64 {
	return a + alpha*(b-a)
}

/*func outputToConsole() {
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

func outputGradients(x, y int) {
	for iX := 0; iX < x; iX++ {
		for iY := 0; iY < y; iY++ {
			vec := gradient_vectors[iX][iY]

			fmt.Print("(" + fmt.Sprintf("%.2f", vec.x) + ", " + fmt.Sprintf("%.2f", vec.y) + ") ")
		}

		fmt.Println()
	}
}*/
