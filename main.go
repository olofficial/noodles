package main

import (
	"fmt"
	"math/rand"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

	"github.com/james-bowman/sparse"
)

//initialize NxN adjacency matrix to symbolize connections
func initialization(N int) *sparse.CSR {
	indptr := make([]int, N+1)
	return sparse.NewCSR(N, N, indptr, nil, nil)
}

//shuffling and connecting end slices 
func randomMatching(noodles int) *sparse.CSR {
	matrix := initialization(noodles)

	//build the ends slice
	ends := make([]int, 0, 2*noodles)
	for i := 0; i < noodles; i++ {
		ends = append(ends, i, i)
	}

	//shuffling ends
	rand.Shuffle(len(ends), func(a, b int) { ends[a], ends[b] = ends[b], ends[a] })

	//pair and set adjacency
	for k := 0; k < len(ends); k += 2 {
		u, v := ends[k], ends[k+1]
		if u == v {
			//trivial self-loops on diagonals
			matrix.Set(u, u, 1)
		} else {
			//off-diagonal connections
			matrix.Set(u, v, matrix.At(u, v)+1)
			matrix.Set(v, u, matrix.At(v, u)+1)
		}
	}
	return matrix
}

//prints the matrix for prettiness
func printCSR(matrix *sparse.CSR) {
	rows, cols := matrix.Dims()
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			fmt.Printf("%4.0f ", matrix.At(i, j))
		}
		fmt.Println()
	}
}

//follows the loops
func countLoops(noodles int, matrix *sparse.CSR) int {
	visited := make([]bool, noodles)
	count := 0

	for i := 0; i < noodles; i++ {
		if visited[i] {
			continue
		}

		//trivial loop checking
		if matrix.At(i, i) > 0 {
			visited[i] = true
			count++
			continue
		}

		head := 0
		queue := make([]int, 0, noodles)
		queue = append(queue, i)
		visited[i] = true

		for head < len(queue) {
			u := queue[head]
			head++

			row := matrix.RowView(u).(*sparse.Vector)
			row.DoNonZero(func(j, _ int, _ float64) {
				if j == u {
					return
				}
				if !visited[j] {
					visited[j] = true
					queue = append(queue, j)
				}
			})
		}
		count++
	}

	return count
}

//separate function for sampling
func noodling(noodles int) int {
	matrix := randomMatching(noodles)
	loops := countLoops(noodles, matrix)
	return loops
}

func plotHistogram(data []int, title string, filename string) {
	//count the freqs, skip 0 (no loops are 0 long)
	freq := make(map[int]int)
	maxVal := 0
	for _, v := range data {
		if v == 0 {
			continue
		}
		freq[v]++
		if v > maxVal {
			maxVal = v
		}
	}

	//plot the histogram
	hist := make(plotter.Values, maxVal+1)
	for k, v := range freq {
		hist[k] = float64(v)
	}

	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = "Loop count"
	p.Y.Label.Text = "Frequency"

	bar, err := plotter.NewBarChart(hist[1:], vg.Points(10)) //skip index 0
	if err != nil {
		panic(err)
	}

	p.Add(bar)
	p.NominalX(makeXLabels(len(hist) - 1)[1:]...) //skip label 0

	if err := p.Save(6*vg.Inch, 4*vg.Inch, filename); err != nil {
		panic(err)
	}
}

//makes x labels
func makeXLabels(n int) []string {
	labels := make([]string, n)
	for i := range labels {
		labels[i] = fmt.Sprintf("%d", i)
	}
	return labels
}

//main function
func main() {
	var numSamples int = 1000000
	countsList := make([]int, numSamples)
	counts := 0
	noodles := 100

	for i := 0; i < numSamples; i++ {
		count := noodling(noodles)
		counts += count
		countsList[i] = count / numSamples
	}

	expected := float64(counts) / float64(numSamples)
	fmt.Printf("Expected loops: %.4f\n", expected)

	plotHistogram(countsList, fmt.Sprintf("Loop Count Distribution (noodles=%d)", noodles), "loop_histogram.png")
}
