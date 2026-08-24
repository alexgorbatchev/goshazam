package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/alexgorbatchev/goshazam"
	"github.com/alexgorbatchev/goshazam/pkg/audio"
)

const version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "goshazam",
		Short: "goshazam is a fast Go library and CLI for Shazam music recognition",
	}

	rootCmd.SetVersionTemplate("{{.Version}}\n")
	rootCmd.Version = version

	var jsonOutput bool
	var language string
	var country string
	var proxy string

	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output raw JSON")
	rootCmd.PersistentFlags().StringVar(&language, "language", "en-US", "Preferred language (e.g. en-US)")
	rootCmd.PersistentFlags().StringVar(&country, "country", "GB", "Catalog country code (e.g. US, GB)")
	rootCmd.PersistentFlags().StringVar(&proxy, "proxy", "", "HTTP/HTTPS/SOCKS5 proxy URL")

	newClient := func() *goshazam.Shazam {
		var opts []goshazam.Option
		opts = append(opts, goshazam.WithLanguage(language))
		opts = append(opts, goshazam.WithEndpointCountry(country))
		if proxy != "" {
			opts = append(opts, goshazam.WithProxy(proxy))
		}
		return goshazam.New(opts...)
	}

	recognizeCmd := &cobra.Command{
		Use:   "recognize <audio-file>",
		Short: "Recognize music from an audio file (WAV, MP3, OGG, FLAC, M4A, etc.)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			client := newClient()
			ctx := context.Background()

			res, err := client.RecognizeFile(ctx, filePath)
			if err != nil {
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(res)
			}

			if len(res.Matches) == 0 || res.Track == nil {
				fmt.Println("No matches found.")
				return nil
			}

			track := res.Track
			fmt.Printf("Track:       %s\n", track.Title)
			fmt.Printf("Artist:      %s\n", track.Subtitle)
			fmt.Printf("Shazam Key:  %s\n", track.Key)
			if track.PhotoURL != "" {
				fmt.Printf("Cover Art:   %s\n", track.PhotoURL)
			}
			if track.AppleMusicURL != "" {
				fmt.Printf("Apple Music: %s\n", track.AppleMusicURL)
			}
			if track.SpotifyURL != "" {
				fmt.Printf("Spotify:     %s\n", track.SpotifyURL)
			}
			if track.YouTubeLink != "" {
				fmt.Printf("YouTube:     %s\n", track.YouTubeLink)
			}
			fmt.Printf("Matches:     %d\n", len(res.Matches))
			return nil
		},
	}

	signatureCmd := &cobra.Command{
		Use:   "signature <audio-file>",
		Short: "Generate Shazam Data URI signature from audio file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			ctx := context.Background()

			seg, err := audio.DecodeAudioFile(ctx, filePath)
			if err != nil {
				return err
			}

			sg := audio.CreateSignatureGenerator(seg)
			sig := sg.GetNextSignature()
			if sig == nil {
				return fmt.Errorf("insufficient audio data to generate signature")
			}

			if jsonOutput {
				out := map[string]any{
					"uri":            sig.EncodeToURI(),
					"sample_rate_hz": sig.SampleRateHz,
					"number_samples": sig.NumberSamples,
					"peaks":          sig.TotalPeaks(),
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Println(sig.EncodeToURI())
			return nil
		},
	}

	relatedCmd := &cobra.Command{
		Use:   "related <track-id>",
		Short: "Find similar/related tracks for a Shazam track ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			trackID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid track ID %q: %w", args[0], err)
			}

			client := newClient()
			ctx := context.Background()

			related, err := client.RelatedTracks(ctx, trackID, 10, 0)
			if err != nil {
				return err
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(related)
		},
	}

	rootCmd.AddCommand(recognizeCmd, signatureCmd, relatedCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
