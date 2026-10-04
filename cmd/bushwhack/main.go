// Command bushwhack turns an SVG into a router template from the command line.
//
//	bushwhack -bushing 5/16 -bit 1/8 -size 6 face.svg
//
// writes face-template.svg and face-reference.svg next to the input.
package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/brennanMKE/Bushwhack/internal/template"
	"github.com/brennanMKE/Bushwhack/internal/units"
)

func main() {
	def := template.DefaultOptions()
	bushing := flag.String("bushing", "5/16", "guide bushing outer diameter (5/16, 0.3125, 8mm)")
	bushingID := flag.String("bushing-id", "", "guide bushing inside diameter; default: the standard one for -bushing")
	bit := flag.String("bit", "1/8", "router bit diameter")
	mode := flag.String("mode", def.Mode, "hole: the opening matches the drawing; piece: the part that falls out matches")
	fit := flag.String("fit", def.Fit, "canvas, artwork or document")
	size := flag.String("size", "6", "longest side in inches for canvas/artwork fit")
	scale := flag.Float64("scale", 0, "explicit inches per SVG unit (overrides -fit)")
	minBridge := flag.String("min-bridge", "1/4", "warn when template material between openings is thinner than this")
	outDir := flag.String("o", "", "output directory (default: next to the input)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: bushwhack [flags] input.svg\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	in := flag.Arg(0)

	o := def
	o.Mode, o.Fit, o.Scale = *mode, *fit, *scale
	for _, f := range []struct {
		name, v string
		dst     *float64
	}{{"bushing", *bushing, &o.BushingOD}, {"bushing-id", *bushingID, &o.BushingID}, {"bit", *bit, &o.BitDia}, {"size", *size, &o.Size}, {"min-bridge", *minBridge, &o.MinBridge}} {
		if f.v == "" {
			continue
		}
		v, err := units.ParseLength(f.v)
		if err != nil {
			die("-%s: %v", f.name, err)
		}
		*f.dst = v
	}

	data, err := os.ReadFile(in)
	if err != nil {
		die("%v", err)
	}
	res, err := template.ProcessWithLimits(context.Background(), data, o, template.Limits{})
	if err != nil {
		die("%v", err)
	}

	dir := *outDir
	if dir == "" {
		dir = filepath.Dir(in)
	}
	base := strings.TrimSuffix(filepath.Base(in), filepath.Ext(in))
	tPath := filepath.Join(dir, base+"-template.svg")
	pPath := filepath.Join(dir, base+"-reference.svg")
	if err := os.WriteFile(tPath, res.TemplateSVG, 0o644); err != nil {
		die("%v", err)
	}
	if err := os.WriteFile(pPath, res.PreviewSVG, 0o644); err != nil {
		die("%v", err)
	}

	fmt.Printf("scale %.5f in/unit, offset %.4f in (%s), page %.3f x %.3f in\n",
		res.ScaleInPerUnit, res.Offset, units.Fraction(res.Offset), res.PageWidthIn, res.PageHeightIn)
	if math.IsInf(res.MinBridgeIn, 1) {
		fmt.Println("min bridge: n/a (one opening)")
	} else {
		fmt.Printf("min bridge %.3f in\n", res.MinBridgeIn)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "shape\twidth\theight\ttmpl w\ttmpl h\tlost\tovercut\t")
	for _, s := range res.Shapes {
		fmt.Fprintf(tw, "%d\t%.3f\t%.3f\t%.3f\t%.3f\t%.1f%%\t%.1f%%\t\n", s.Index, s.WidthIn, s.HeightIn,
			s.TemplateWidthIn, s.TemplateHeightIn, s.LostAreaPct, s.OvercutPct)
	}
	tw.Flush()
	for _, w := range res.Warnings {
		fmt.Println("warning:", w)
	}
	fmt.Println("wrote", tPath)
	fmt.Println("wrote", pPath)
}

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "bushwhack: "+format+"\n", a...)
	os.Exit(1)
}
