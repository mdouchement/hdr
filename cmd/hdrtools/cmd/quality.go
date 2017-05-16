package cmd

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/mdouchement/hdr"
	// Import RGBE decoder
	_ "github.com/mdouchement/hdr/rgbe"
	"github.com/mdouchement/hdr/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	// QualityCommand defines the command for compute quality check.
	QualityCommand = &cobra.Command{
		Use:   "quality [flags] source_file target_file",
		Short: "Performs PSNR and SSIM",
		Long:  "Performs PSNR and SSIM",
		RunE:  qualityAction,
	}
)

func qualityAction(c *cobra.Command, args []string) error {
	if len(args) != 2 {
		return errors.New("quality: Invalid number of arguments")
	}
	f1, err := os.Open(args[0])
	if err != nil {
		return errors.Wrap(err, "quality:")
	}
	defer f1.Close()

	m1, fname, err := image.Decode(f1)
	if err != nil {
		return errors.Wrap(err, "quality:")
	}
	fmt.Printf("Read image (%dx%dp - %s) %s\n", m1.Bounds().Dx(), m1.Bounds().Dy(), fname, filepath.Base(args[0]))

	f2, err := os.Open(args[1])
	if err != nil {
		return errors.Wrap(err, "quality:")
	}
	defer f2.Close()

	m2, fname, err := image.Decode(f2)
	if err != nil {
		return errors.Wrap(err, "quality:")
	}
	fmt.Printf("Read image (%dx%dp - %s) %s\n", m2.Bounds().Dx(), m2.Bounds().Dy(), fname, filepath.Base(args[1]))

	hdrm1 := m1.(hdr.Image)
	hdrm2 := m2.(hdr.Image)
	mse, snr, psnr, peak := util.PSNR(hdrm1, hdrm2)
	ssim := util.SSIM(hdrm1, hdrm2)

	fmt.Printf("MSE: %.8f\n", mse)
	fmt.Printf("SNR: %.2f dB\n", snr)
	fmt.Printf("PSNR(max=%.4f): %.2f dB\n", peak, psnr)
	fmt.Printf("SSIM: %.8f\n", ssim)
	return nil
}
