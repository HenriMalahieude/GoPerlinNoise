package main

import (
	"fmt"
	"math"
	"strconv"
	"syscall/js"
)

func main() {
	fmt.Println("Go Web Assembly Is Now Running!")
	js.Global().Set("genPerlin", wasmGetPerlin())

	//Indefinite Wait
	<-make(chan bool)
}

func wasmGetPerlin() js.Func {
	perlFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 5 {
			return "Invalid no of arguments"
		}
		grid, err1 := strconv.Atoi(args[0].String())
		res, err2 := strconv.Atoi(args[1].String())
		smooth := args[2].Bool()
		extreme := args[3].Bool()
		terrain := args[4].Bool()

		if grid == 0 || res == 0 {
			return "Cannot accept zero"
		}

		if err1 != nil || err2 != nil {
			return "Couldn't convert to numbers"
		}

		generateRandomGradients(grid, grid, extreme)
		generateDepthValues(res, terrain)

		jsDoc := js.Global().Get("document")
		if !jsDoc.Truthy() {
			return "Unable to get document object"
		}

		width, height := float64(330), float64(200) //I have no idea why I can't get these normally, but fuck it whatever
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
				if smooth {
					val += 1
					val /= 2 //now scaled from 0 -> 1

					r := val
					g := (-(math.Pow(val, 2))+val)*5 - 0.25 //Tuned to perfection
					b := 1 - val

					str := "#"
					str += valToHex(r)
					str += valToHex(g)
					str += valToHex(b)

					ctx.Set("fillStyle", str)
				} else {
					if val > 0.9 {
						ctx.Set("fillStyle", "#000000")
					} else if val > 0.75 {
						ctx.Set("fillStyle", "#FF0000")
					} else if val > 0.4 {
						ctx.Set("fillStyle", "#AAFF00")
					} else if val > 0 {
						ctx.Set("fillStyle", "#00FF00")
					} else if val > -0.6 {
						ctx.Set("fillStyle", "#00AAFF")
					} else {
						ctx.Set("fillStyle", "#0000FF")
					}
				}

				ctx.Call("fillRect", float64(iX)*widthStep, float64(iY)*heightStep, widthStep*2, heightStep*2)
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
