//go:build !js

package main

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/zergon321/reisen"

	"bytes"
	"encoding/binary"
	"image"

	//_ "github.com/silbinarywolf/preferdiscretegpu"
)

// Video playback
type bgVideo struct {
	started                bool
	errs                   chan error
	frameBuffer            chan *image.RGBA
	sampleBuffer           chan [2]float64
	texture *Texture
}

const (
	frameBufferSize = 10
	sampleBufferSize = 8192
)

func (bgv *bgVideo) Open(filename string) (error) {
	// Open the media file.
	media, err := reisen.NewMedia(filename)
	if err != nil {
		return err
	}

	// Get the FPS for playing video frames.
	_, _ = media.Streams()[0].FrameRate()

	bgv.frameBuffer = make(chan *image.RGBA, frameBufferSize)
	bgv.sampleBuffer = make(chan [2]float64, sampleBufferSize)
	bgv.errs = make(chan error)

	err = media.OpenDecode()
	if err != nil {
		return err
	}

	videoStream := media.VideoStreams()[0]
	err = videoStream.Open()
	if err != nil {
		return err
	}

	audioStream := media.AudioStreams()[0]
	err = audioStream.Open()
	if err != nil {
		return err
	}

	speaker.Play(streamSamples(bgv.sampleBuffer))

	go func() {
		for {
			gotPacket := bgv.processPacket(media)
			if !gotPacket {
				break
			}
		}
		videoStream.Close()
		audioStream.Close()
		media.CloseDecode()
		close(bgv.frameBuffer)
		close(bgv.sampleBuffer)
		close(bgv.errs)
	}()

	return nil
}

// processPacket reads video and audio frames
// from the opened media and sends the decoded
// data to che channels to be played.
func (bgv *bgVideo) processPacket(media *reisen.Media) bool {
	packet, gotPacket, err := media.ReadPacket()
	if err != nil {
		bgv.errs <- err
	}

	if !gotPacket {
		return false
	}

	switch packet.Type() {
	case reisen.StreamVideo:
		s := media.Streams()[packet.StreamIndex()].(*reisen.VideoStream)
		videoFrame, gotFrame, err := s.ReadVideoFrame()

		if err != nil {
			bgv.errs <- err
		}

		if !gotFrame {
			return false
		}

		if videoFrame != nil {
println("Pushing video frame... ", len(bgv.frameBuffer), " elements for now")
			bgv.frameBuffer <- videoFrame.Image()
println("OK!v")
		}

	case reisen.StreamAudio:
		s := media.Streams()[packet.StreamIndex()].(*reisen.AudioStream)
		audioFrame, gotFrame, err := s.ReadAudioFrame()

		if err != nil {
			bgv.errs <- err
		}

		if !gotFrame {
			return false
		}

		if audioFrame != nil {
			reader := bytes.NewReader(audioFrame.Data())
println("Pushing audio frame... ", len(bgv.sampleBuffer), " elements for now")
			for reader.Len() >= 16 {
				sample := [2]float64{0, 0}
				err = binary.Read(reader, binary.LittleEndian, sample[:])
				if err != nil {
					bgv.errs <- err
				}
				bgv.sampleBuffer <- sample
			}
println("OK!a")
		}
	}

	return true
}

// streamSamples creates a new custom streamer for
// playing audio samples provided by the source channel.
//
// See https://github.com/faiface/beep/wiki/Making-own-streamers
// for reference.
func streamSamples(sampleBuffer <-chan [2]float64) beep.Streamer {
	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		for i := 0; i < len(samples); i++ {
			sample, ok := <-sampleBuffer

			if !ok {
				return i, false
			}

			samples[i] = sample
		}

		return len(samples), true
	})
}

func (bgv *bgVideo) Tick() error {
	// Check for incoming errors.
	select {
	case err, ok := <-bgv.errs:
		if ok {
			return err
		}

	default:
	}

	if !bgv.started {
		// Start playing audio samples.
		//speaker.Play(streamSamples(bgv.sampleBuffer))
		bgv.started = true
	}

	// Read video frames and draw them.
	if frame, ok := <-bgv.frameBuffer; ok {
		rect := frame.Bounds()
		width := int32(rect.Max.X - rect.Min.X)
		height := int32(rect.Max.Y - rect.Min.Y)
		if bgv.texture == nil || width != bgv.texture.width || height != bgv.texture.height {
			bgv.texture = newTexture(width, height, 32, true)
		}
		bgv.texture.SetData(frame.Pix)
	}

	return nil
}
