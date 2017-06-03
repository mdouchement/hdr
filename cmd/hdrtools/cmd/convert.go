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

	toxyze bool
	torgbe bool
	tohdr  bool
	tocrad bool
)

func init() {
	ConvertCommand.Flags().BoolVarP(&toxyze, "to-xyze", "", false, "Converts to Radiance XYZE")
	ConvertCommand.Flags().BoolVarP(&torgbe, "to-rgbe", "", false, "Converts to Radiance RGBE")
	ConvertCommand.Flags().BoolVarP(&tohdr, "to-hdr", "", false, "Converts to Radiance RGBE/XYZE")
	ConvertCommand.Flags().BoolVarP(&tocrad, "to-crad", "", false, "Converts to CRAD")
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
	case toxyze:
		rgbe.Encode(fo, toXYZ(hdrm))
	case torgbe:
		rgbe.Encode(fo, toRGB(hdrm))
	case tohdr:
		rgbe.Encode(fo, hdrm)
	case tocrad:
		crad.Encode(fo, hdrm)
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
