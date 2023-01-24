# Command Line TTS Tool

This is a command line tool that accepts a text input from stdin or a file 
and converts it into audio using Google Cloud TTS or Azure TTS.

This is a fork of [Google Squawk](https://github.com/ruhnet/google-squawk) 
by  [ruhnet](https://github.com/ruhnet). I just added support for Azure TTS.

## Build

Azure TTS support requires Microsoft Cognitive Speech Services SDK to be installed.

Follow the instructions for [installing the SDK](https://learn.microsoft.com/en-us/azure/cognitive-services/speech-service/quickstarts/setup-platform?pivots=programming-language-go&tabs=windows%2Cubuntu%2Cdotnet%2Cjre%2Cmaven%2Cbrowser%2Cmac%2Cpypi#install-the-speech-sdk) on your platform.

```
export SPEECHSDK_ROOT="$PWD/speech-sdk"
mkdir -p "$SPEECHSDK_ROOT"
```

### macOS

```
wget -O MicrosoftCognitiveServicesSpeech.xcframework.zip https://aka.ms/csspeech/macosbinary
unzip MicrosoftCognitiveServicesSpeech.xcframework.zip -d "$SPEECHSDK_ROOT"
export CGO_CFLAGS="-I$SPEECHSDK_ROOT/MicrosoftCognitiveServicesSpeech.xcframework/macos-arm64_x86_64/MicrosoftCognitiveServicesSpeech.framework/Headers"
export CGO_LDFLAGS="-Wl,-rpath,$SPEECHSDK_ROOT/MicrosoftCognitiveServicesSpeech.xcframework/macos-arm64_x86_64 -F$SPEECHSDK_ROOT/MicrosoftCognitiveServicesSpeech.xcframework/macos-arm64_x86_64 -framework MicrosoftCognitiveServicesSpeech"
```

### Linux 

```
wget -O SpeechSDK-Linux.tar.gz https://aka.ms/csspeech/linuxbinary
tar --strip 1 -xzf SpeechSDK-Linux.tar.gz -C "$SPEECHSDK_ROOT"
export CGO_CFLAGS="-I$SPEECHSDK_ROOT/include/c_api"
export CGO_LDFLAGS="-L$SPEECHSDK_ROOT/lib/x64 -lMicrosoft.CognitiveServices.Speech.core"
```

Now we are ready to build the application.

```
go build
```

## Authentication

### Google Cloud TTS

To use the program with Google Cloud TTS, specify the path to your Google cloud account credentials with the environment variable `GOOGLE_APPLICATION_CREDENTIALS`.
 
```
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/serviceaccount/credentials_file.json
```

### Azure Cloud TTS

To use Azure Cognitive service, specify your Azure API Key and region using environment varialbles `AZURE_API_KEY` and `AZURE_REGION`.

```
export AZURE_API_KEY="..."
export AZURE_REGION="centralindia"
```

## Usage

By default, input is supplied via standard input, but can also be specified from a file with the `-i` option.
Both plain text and SSML input are supported. Specify `-ssml` to tell the program to expect SSML input from your text source.

Output is to a file, but you can set the filename to `-` to send to stdout. This can be useful if you want to convert the file on the fly to a format that the program doesn't support, with **sox** or **ffmpeg** or the like.

Specify the path to the Speech SDK library using

Run ```tts -h``` to see the help:

```
Usage of ./tts:
  -db float
    	Volume gain in dB. [-96 to 16]
  -e string
        TTS Engine. ['google', 'azure'] (default "google")
  -f string
    	Audio format selection. MP3 is 32k [mp3,pcm] (default "mp3")
  -g string
    	Gender selection. [m,f,n] 'n' means neutral/don't care. (default "m")
  -i string
    	Input file path. Defaults to stdin. (default "-")
  -l string
    	Language selection. 'en-US', 'en-GB', 'en-AU', 'en-IN',
    	'el-GR', 'ru-RU', etc. (default "en-US")
  -o string
    	Output file path. Use '-' for stdout. (default "./tts.mp3")
  -p float
    	Pitch. E.g. '0.0' is normal. '20.0' is highest,
    	'-20.0' is lowest. (default 1)
  -r int
    	Samplerate in Hz. [8000,11025,16000,22050,24000,32000,44100,48000] (default 24000)
  -s float
    	Speed. E.g. '1.0' is normal. '2.0' is double
    	speed, '0.25' is quarter speed, etc. (default 1)
  -ssml
    	Input is SSML format, rather than plain text.
  -v string
    	Voice. If specified, this overrides language & gender. (default "")

```

## Troubleshooting

If you are getting the error 

> error while loading shared libraries: libMicrosoft.CognitiveServices.Speech.core.so: cannot open shared object file: No such file or directory

either copy the shared libraries to `/usr/lib` 
 
```
cp $SPEECHSDK_ROOT/lib/x64/*.so /usr/lib/
```

or add the library path to the `LD_LIBRARY_PATH` environment variable

```
export LD_LIBRARY_PATH="$SPEECHSDK_ROOT/lib/x64:$LD_LIBRARY_PATH"
```

