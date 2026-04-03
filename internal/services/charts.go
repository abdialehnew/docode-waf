package services

import (
	"bytes"
	"fmt"
	"image/color"
	"time"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func generateTrafficPieChart(allowed, blocked int) ([]byte, error) {
	pie := chart.PieChart{
		Width:  800,
		Height: 400,
		Values: []chart.Value{
			{Value: float64(allowed), Label: fmt.Sprintf("Allowed (%d)", allowed), Style: chart.Style{FillColor: drawing.Color{R: 76, G: 175, B: 80, A: 255}}}, // Green
			{Value: float64(blocked), Label: fmt.Sprintf("Blocked (%d)", blocked), Style: chart.Style{FillColor: drawing.Color{R: 244, G: 67, B: 54, A: 255}}}, // Red
		},
	}

	var buffer bytes.Buffer
	err := pie.Render(chart.PNG, &buffer)
	return buffer.Bytes(), err
}

func generateAttacksBarChart(stats []StatItem) ([]byte, error) {
	if len(stats) == 0 {
		return nil, fmt.Errorf("no data to render bar chart")
	}

	// Create a new plot
	p := plot.New()
	p.Title.Text = ""
	p.Y.Label.Text = ""
	p.X.Label.Text = ""

	// Create bar values
	values := make(plotter.Values, len(stats))
	labels := make([]string, len(stats))
	for i, s := range stats {
		values[i] = float64(s.Count)
		labels[i] = s.Name
	}

	// Create bars
	bars, err := plotter.NewBarChart(values, vg.Points(40))
	if err != nil {
		return nil, err
	}
	bars.Color = color.RGBA{R: 66, G: 133, B: 244, A: 255} // Blue

	// Add bars to the plot
	p.Add(bars)
	p.NominalX(labels...)

	// Set Y-axis to start from 0
	p.Y.Min = 0

	// Write to buffer as PNG
	writer, err := p.WriterTo(8*vg.Inch, 4*vg.Inch, "png")
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	_, err = writer.WriteTo(&buffer)
	return buffer.Bytes(), err
}

func generateTrafficLineChart(dailyStats []DailyStat) ([]byte, error) {
	var xValues []float64
	var yTotal []float64
	var yBlocked []float64
	var xTicks []chart.Tick

	for i, s := range dailyStats {
		xValues = append(xValues, float64(i))
		yTotal = append(yTotal, float64(s.Total))
		yBlocked = append(yBlocked, float64(s.Blocked))

		// Simplify label: "01-Oct"
		label := s.Date
		if len(s.Date) >= 10 {
			t, _ := time.Parse("2006-01-02", s.Date)
			label = t.Format("02-Jan")
		}

		xTicks = append(xTicks, chart.Tick{
			Value: float64(i),
			Label: label,
		})
	}

	graph := chart.Chart{
		Width:  800,
		Height: 400,
		Background: chart.Style{
			Padding: chart.Box{Top: 20, Left: 20, Right: 20, Bottom: 20},
		},
		XAxis: chart.XAxis{
			Ticks: xTicks,
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				Hidden: false,
			},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Name:    "Total Requests",
				XValues: xValues,
				YValues: yTotal,
				Style: chart.Style{
					StrokeColor: drawing.Color{R: 0, G: 0, B: 0, A: 255}, // Dark
					StrokeWidth: 2,
					DotWidth:    4,
				},
			},
			chart.ContinuousSeries{
				Name:    "Blocked Requests",
				XValues: xValues,
				YValues: yBlocked,
				Style: chart.Style{
					StrokeColor: drawing.Color{R: 244, G: 67, B: 54, A: 255}, // Red
					StrokeWidth: 2,
					DotWidth:    4,
				},
			},
		},
	}

	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	var buffer bytes.Buffer
	err := graph.Render(chart.PNG, &buffer)
	return buffer.Bytes(), err
}

func generateAttackVectorsBarChart(stats []StatItem) ([]byte, error) {
	if len(stats) == 0 {
		return nil, fmt.Errorf("no data to render attack vectors chart")
	}

	// Create a new plot
	p := plot.New()
	p.Title.Text = ""
	p.Y.Label.Text = ""
	p.X.Label.Text = ""

	// Create bar values
	values := make(plotter.Values, len(stats))
	labels := make([]string, len(stats))
	for i, s := range stats {
		values[i] = float64(s.Count)
		labels[i] = s.Name
	}

	// Create bars with orange color
	bars, err := plotter.NewBarChart(values, vg.Points(50))
	if err != nil {
		return nil, err
	}
	bars.Color = color.RGBA{R: 255, G: 165, B: 0, A: 255} // Orange

	// Add bars to the plot
	p.Add(bars)
	p.NominalX(labels...)

	// Set Y-axis to start from 0
	p.Y.Min = 0

	// Write to buffer as PNG
	writer, err := p.WriterTo(8*vg.Inch, 4*vg.Inch, "png")
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	_, err = writer.WriteTo(&buffer)
	return buffer.Bytes(), err
}

func generateTrafficDonutChart(allowed, blocked int) ([]byte, error) {
	// Donut chart using go-chart PieChart
	pie := chart.PieChart{
		Width:  500,
		Height: 500,
		Values: []chart.Value{
			{Value: float64(allowed), Label: "Allowed", Style: chart.Style{FillColor: drawing.Color{R: 76, G: 175, B: 80, A: 255}}}, // Green
			{Value: float64(blocked), Label: "Blocked", Style: chart.Style{FillColor: drawing.Color{R: 244, G: 67, B: 54, A: 255}}}, // Red
		},
	}

	var buffer bytes.Buffer
	err := pie.Render(chart.PNG, &buffer)
	return buffer.Bytes(), err
}
