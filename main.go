// Copyright 2022 Ruel Tmeizeh - All Rights Reserved
package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	// "github.com/Microsoft/cognitive-services-speech-sdk-go/audio"
	"github.com/Microsoft/cognitive-services-speech-sdk-go/common"
	"github.com/Microsoft/cognitive-services-speech-sdk-go/speech"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

type CommandlineOptions struct {
	ListVoices *bool    `json:"listvoices,omitempty"`
	Ssml       *bool    `json:"ssml,omitempty"`
	Output     *string  `json:"output,omitempty"`
	Input      *string  `json:"input,omitempty"`
	Language   *string  `json:"language,omitempty"`
	Gender     *string  `json:"gender,omitempty"`
	Voice      *string  `json:"voice,omitempty"`
	Format     *string  `json:"format,omitempty"`
	Speed      *float64 `json:"speed,omitempty"`
	Pitch      *float64 `json:"pitch,omitempty"`
	SampleRate *int     `json:"samplerate,omitempty"`
	VolumeGain *float64 `json:"volume,omitempty"`
	Engine     *string  `json:"format,omitempty"`
}

func main() {
	//check commandline args:
	opts := &CommandlineOptions{
		ListVoices: flag.Bool("listvoices", false, "List available voices, rather than generate TTS. Use in\ncombination with '-l ALL' to show voices from all languages."),
		Ssml:       flag.Bool("ssml", false, "Input is SSML format, rather than plain text."),
		Input:      flag.String("i", "-", "Input file path. Defaults to stdin."),
		Output:     flag.String("o", "./tts.mp3", "Output file path. Use '-' for stdout."),
		Language:   flag.String("l", "en-US", "Language selection. 'en-US', 'en-GB', 'en-AU', 'en-IN',\n'el-GR', 'ru-RU', etc."),
		Gender:     flag.String("g", "m", "Gender selection. [m,f,n] 'n' means neutral/don't care."),
		Format:     flag.String("f", "mp3", "Audio format selection. MP3 is 32k [mp3,opus,pcm,ulaw,alaw]"),
		Voice:      flag.String("v", "unspecified", "Voice. If specified, this overrides language & gender."),
		Speed:      flag.Float64("s", 1.0, "Speed. E.g. '1.0' is normal. '2.0' is double\nspeed, '0.25' is quarter speed, etc."),
		Pitch:      flag.Float64("p", 1.0, "Pitch. E.g. '0.0' is normal. '20.0' is highest,\n'-20.0' is lowest."),
		SampleRate: flag.Int("r", 24000, "Samplerate in Hz. [8000,11025,16000,22050,24000,32000,44100,48000]"),
		VolumeGain: flag.Float64("db", 0.0, "Volume gain in dB. [-96 to 16]"),
		Engine:     flag.String("e", "google", "TTS engine to use [google,azure]"),
	}
	flag.Parse()

	var fileExtension string
	switch *opts.Format {
	case "mp3":
		fileExtension = "mp3"
	case "opus":
		fileExtension = "ogg"
	case "ogg":
		fileExtension = "ogg"
	case "pcm":
		fileExtension = "pcm"
	case "ulaw":
		fileExtension = "ulaw"
	case "alaw":
		fileExtension = "alaw"
	default:
		fileExtension = "mp3"
	}

	filename := "tts." + fileExtension
	if *opts.Output != "./tts.mp3" {
		filename = *opts.Output
	}

	///////////////////////////////////////

	var inputFile *os.File
	if *opts.Input == "-" {
		//read input from stdin
		inputFile = os.Stdin
	} else {
		//read input from file
		var err error
		inputFile, err = os.Open(*opts.Input)
		if err != nil {
			log.Fatal(err)
		}
		defer inputFile.Close()
	}

	var input string

	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		//fmt.Println(scanner.Text())
		input = input + scanner.Text()
	}

	var audioContent []byte
	var err error
	if *opts.Engine == "google" {
		audioContent, err = ttsWithGoogleCloud(input, *opts.Ssml, *opts.Gender, *opts.Language, *opts.Voice, *opts.Format, *opts.Speed, *opts.Pitch, *opts.VolumeGain, int32(*opts.SampleRate))
	} else {
		audioContent, err = ttsWithAzureCloud(input, *opts.Ssml, *opts.Gender, *opts.Language, *opts.Voice, *opts.Format, *opts.Speed, *opts.Pitch, *opts.VolumeGain, int32(*opts.SampleRate))
	}

	if err != nil {
		log.Fatal(err)
	}

	if *opts.Output == "-" { //write to stdout
		//binary.Write(os.Stdout, binary.LittleEndian, resp.AudioContent)
		bufStdout := bufio.NewWriter(os.Stdout) //add a buffer
		defer bufStdout.Flush()
		binary.Write(bufStdout, binary.LittleEndian, audioContent)
	} else { //write to file
		err = ioutil.WriteFile(filename, audioContent, 0644)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Audio content written to file: %v\n", filename)
	}
}

func ttsWithAzureCloud(input string, ssml bool, genderCode string, lang string, voiceName string, format string, speed float64, pitch float64, db float64, rate int32) ([]byte, error) {
	// stream, err := audio.CreatePullAudioOutputStream()
	// if err != nil {
	// 	log.Panic(fmt.Sprintf("create pull audio output stream error: %v", err))
	// }
	// defer stream.Close()

	var apiKey = os.Getenv("AZURE_API_KEY")
	var apiRegion = os.Getenv("AZURE_REGION")
	log.Printf("Api Key: %s\nRegions: %s\n", apiKey, apiRegion)

	speechConfig, err := speech.NewSpeechConfigFromSubscription(apiKey, apiRegion)
	if err != nil {
		log.Panic(err)
	}

	var audioFormat = azureAudioFormatFromType(format, rate)
	err = speechConfig.SetSpeechSynthesisOutputFormat(audioFormat)
	log.Printf("Format: %d\n", audioFormat)
	if err != nil {
		log.Panic(err)
	}
	defer speechConfig.Close()

	if voiceName != "" {
		err := speechConfig.SetSpeechSynthesisVoiceName(voiceName)
		if err != nil {
			log.Panic(err)
		}
	}
	log.Printf("Voice: %s\n", voiceName)

	speechSynthesizer, err := speech.NewSpeechSynthesizerFromConfig(speechConfig, nil)
	if err != nil {
		log.Panic(err)
	}
	defer speechSynthesizer.Close()

	log.Printf("Synthesize: %s to %s\n", input, "stdout")
	task := speechSynthesizer.SpeakTextAsync(input)

	var outcome speech.SpeechSynthesisOutcome
	select {
	case outcome = <-task:
	case <-time.After(10 * time.Second):
		log.Panic("Synthesize timed out")
	}
	defer outcome.Close()

	if outcome.Error != nil {
		log.Panic(outcome.Error.Error())
	}

	if outcome.Result.Reason == common.SynthesizingAudioCompleted {
		return outcome.Result.AudioData, nil
	}

	cancellation, err := speech.NewCancellationDetailsFromSpeechSynthesisResult(outcome.Result)
	if err != nil {
		return []byte{}, err
	}

	return []byte{}, errors.New(cancellation.ErrorDetails)
}
func azureAudioFormatFromType(format string, sample_rate int32) common.SpeechSynthesisOutputFormat {

	format = fmt.Sprintf("%s_%d", format, sample_rate)
	switch format {
	case "mp3_16000":
		return common.Audio16Khz32KBitRateMonoMp3
	case "mp3_24000":
		return common.Audio24Khz48KBitRateMonoMp3
	case "mp3_48000":
		return common.Audio48Khz96KBitRateMonoMp3
	case "pcm_8000":
		return common.Riff8Khz16BitMonoPcm
	case "pcm_16000":
		return common.Riff16Khz16BitMonoPcm
	case "pcm_22050":
		return common.Riff22050Hz16BitMonoPcm
	case "pcm_24000":
		return common.Riff24Khz16BitMonoPcm
	case "pcm_44100":
		return common.Riff44100Hz16BitMonoPcm
	case "pcm_48000":
		return common.Riff48Khz16BitMonoPcm
	default:
		return common.Audio16Khz32KBitRateMonoMp3
	}
}

func ttsWithGoogleCloud(input string, ssml bool, genderCode string, lang string, voiceName string, format string, speed float64, pitch float64, db float64, rate int32) ([]byte, error) {
	//Instantiates a Google Cloud client
	ctx := context.Background()
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	//Start building TTS request things
	synthInput := &texttospeechpb.SynthesisInput{}
	synthInput.InputSource = &texttospeechpb.SynthesisInput_Text{Text: input}
	if ssml {
		synthInput.InputSource = &texttospeechpb.SynthesisInput_Ssml{Ssml: input}
	}

	// Voice Gender
	var gender texttospeechpb.SsmlVoiceGender
	switch genderCode {
	case "m":
		gender = texttospeechpb.SsmlVoiceGender_MALE
	case "f":
		gender = texttospeechpb.SsmlVoiceGender_FEMALE
	default:
		gender = texttospeechpb.SsmlVoiceGender_NEUTRAL
	}

	voice := &texttospeechpb.VoiceSelectionParams{
		LanguageCode: lang,
		SsmlGender:   gender,
	}
	if voiceName != "unspecified" {
		voice.Name = voiceName
	}

	var audioFormat = googleAudioFormatFromType(format)

	// the request parameters
	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: synthInput,
		Voice: voice,
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding:   audioFormat,
			SpeakingRate:    speed,
			SampleRateHertz: rate,
			Pitch:           pitch,
			VolumeGainDb:    db,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		return []byte{}, err
	}

	return resp.AudioContent, nil
}

func googleAudioFormatFromType(format string) texttospeechpb.AudioEncoding {
	switch format {
	case "mp3":
		return texttospeechpb.AudioEncoding_MP3
	case "opus":
		return texttospeechpb.AudioEncoding_OGG_OPUS
	case "ogg":
		return texttospeechpb.AudioEncoding_OGG_OPUS
	case "pcm":
		return texttospeechpb.AudioEncoding_LINEAR16
	case "ulaw":
		return texttospeechpb.AudioEncoding_MULAW
	case "alaw":
		return texttospeechpb.AudioEncoding_ALAW
	default:
		return texttospeechpb.AudioEncoding_MP3
	}
}
