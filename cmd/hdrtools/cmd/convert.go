package cmd

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/crad"
	"github.com/mdouchement/hdr/rgbe"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	// ConvertCommand defines the command for compute image conversion.
	ConvertCommand = &cobra.Command{
		Use:   "convert [flags] source_file target_file",
		Short: "Convert image",
		Long:  "Convert image",
		RunE:  convertAction,
	}

	rgbe2xyze bool
	xyze2rgbe bool
	hdr2crad  bool
	crad2hdr  bool
)

func init() {
	ConvertCommand.Flags().BoolVarP(&rgbe2xyze, "rgbe2xyze", "", false, "Radiance RGBE to Radiance XYZE")
	ConvertCommand.Flags().BoolVarP(&xyze2rgbe, "xyze2rgbe", "", false, "Radiance XYZE to Radiance RGBE")
	ConvertCommand.Flags().BoolVarP(&hdr2crad, "hdr2crad", "", false, "Radiance RGBE/XYZE to CRAD")
	ConvertCommand.Flags().BoolVarP(&crad2hdr, "crad2hdr", "", false, "CRAD to Radiance RGBE/XYZE")
}

func convertAction(c *cobra.Command, args []string) error {
	if len(args) != 2 {
		return errors.New("convert: Invalid number of arguments")
	}

	fi, err := os.Open(args[0])
	if err != nil {
		return errors.Wrap(err, "convert:")
	}
	defer fi.Close()

	m, fname, err := image.Decode(fi)
	if err != nil {
		return errors.Wrap(err, "convert:")
	}
	fmt.Printf("Read image (%dx%dp - %s) %s\n", m.Bounds().Dx(), m.Bounds().Dy(), fname, filepath.Base(args[0]))

	fo, err := os.Create(args[1])
	if err != nil {
		return errors.Wrap(err, "convert:")
	}
	defer fo.Close()
	fmt.Printf("Write image (%dx%dp - %s) %s\n", m.Bounds().Dx(), m.Bounds().Dy(), fname, filepath.Base(args[1]))

	hdrm := m.(hdr.Image)
	switch {
	case rgbe2xyze:
		rgbe.Encode(fo, toXYZ(hdrm))
	case xyze2rgbe:
		rgbe.Encode(fo, toRGB(hdrm))
	case hdr2crad:
		crad.Encode(fo, hdrm)
	case crad2hdr:
		rgbe.Encode(fo, hdrm)
	default:
		return errors.New("convert: No converion flage provided")
	}

	return nil
}

func toRGB(m hdr.Image) hdr.Image {
	d := m.Bounds()

	new := hdr.NewRGB(d)
	for y := 0; y < d.Dy(); y++ {
		for x := 0; x < d.Dx(); x++ {
			new.Set(x, y, m.HDRAt(x, y))
		}
	}

	return new
}

func toXYZ(m hdr.Image) hdr.Image {
	d := m.Bounds()

	new := hdr.NewXYZ(d)
	for y := 0; y < d.Dy(); y++ {
		for x := 0; x < d.Dx(); x++ {
			new.Set(x, y, m.HDRAt(x, y))
		}
	}

	return new
}
