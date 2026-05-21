package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kenrogers/openrouter-cli/internal/client"
	"github.com/kenrogers/openrouter-cli/internal/output"
	"github.com/spf13/cobra"
)

func newOpenRouterAudioCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audio",
		Short: "Generate speech and transcribe audio",
		Long:  "Friendly audio workflows for OpenRouter text-to-speech and speech-to-text models.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newOpenRouterAudioSpeakCommand())
	cmd.AddCommand(newOpenRouterAudioTranscribeCommand())
	cmd.AddCommand(newOpenRouterAudioModelsCommand())
	return cmd
}

func newOpenRouterAudioSpeakCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "speak [text]",
		Short: "Generate speech audio from text",
		Long:  "Generate speech audio from text and save it to an audio file.",
		Example: `  openrouter audio speak "Ship it, but make it calm." --output voice.mp3
  openrouter audio speak --model openai/gpt-4o-mini-tts-2025-12-15 --voice alloy --format mp3 "Hello from OpenRouter"`,
		RunE: runOpenRouterAudioSpeakCommand,
	}
	cmd.Flags().StringP("text", "t", "", "Text to synthesize (can also be provided as positional args)")
	cmd.Flags().StringP("file", "f", "", "Read text from a file")
	cmd.Flags().Bool("stdin", false, "Read text from stdin")
	cmd.Flags().StringP("model", "m", "auto", "Speech model ID, or auto to choose a current speech model")
	cmd.Flags().String("voice", "alloy", "Voice identifier")
	cmd.Flags().String("format", "mp3", "Audio format: mp3 or pcm")
	cmd.Flags().Float64("speed", 0, "Playback speed multiplier for models that support it")
	cmd.Flags().String("provider", "", "Provider routing JSON object")
	cmd.Flags().String("output", "", "Output file path or directory (default: openrouter-audio-<timestamp>.<format>)")
	cmd.Flags().Bool("force", false, "Overwrite output files if they already exist")
	return cmd
}

func newOpenRouterAudioTranscribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transcribe <audio-file>",
		Short: "Transcribe speech to text",
		Long:  "Transcribe a local audio file through OpenRouter speech-to-text models.",
		Args:  cobra.ExactArgs(1),
		Example: `  openrouter audio transcribe meeting.mp3
  openrouter audio transcribe meeting.wav --model google/chirp-3 --language en --json`,
		RunE: runOpenRouterAudioTranscribeCommand,
	}
	cmd.Flags().StringP("model", "m", "auto", "Transcription model ID, or auto to choose a current STT model")
	cmd.Flags().String("format", "auto", "Input audio format, or auto to infer from file extension")
	cmd.Flags().StringP("language", "l", "", "ISO-639-1 language code, such as en or ja")
	cmd.Flags().Float64("temperature", -1, "Sampling temperature")
	cmd.Flags().String("provider", "", "Provider routing JSON object")
	cmd.Flags().String("output", "", "Optional transcript output file")
	cmd.Flags().Bool("raw", false, "Print only the transcript text")
	cmd.Flags().Bool("force", false, "Overwrite output files if they already exist")
	return cmd
}

func newOpenRouterAudioModelsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "List audio models",
		Long:  "List speech or transcription models available through OpenRouter.",
		RunE:  runOpenRouterAudioModelsCommand,
	}
	cmd.Flags().String("kind", "speech", "Audio model kind: speech or transcription")
	cmd.Flags().Int("limit", 20, "Maximum number of models to show (0 for all)")
	return cmd
}

func runOpenRouterAudioSpeakCommand(cmd *cobra.Command, args []string) error {
	text, err := audioSpeakTextFromFlags(cmd, args)
	if err != nil {
		return err
	}
	apiKey, keySource, err := openRouterAPIKey(cmd, true)
	if err != nil {
		return err
	}
	model, _ := cmd.Flags().GetString("model")
	model = strings.TrimSpace(model)
	if model == "" || strings.EqualFold(model, "auto") {
		model, err = selectWorkflowModel(cmd, "speech", preferredSpeechModels, defaultSpeechModel)
		if err != nil {
			return err
		}
	}
	format, _ := cmd.Flags().GetString("format")
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "mp3"
	}
	if format != "mp3" && format != "pcm" {
		return fmt.Errorf("invalid --format %q: expected mp3 or pcm", format)
	}
	voice, _ := cmd.Flags().GetString("voice")
	body := map[string]any{
		"model":           model,
		"input":           text,
		"voice":           strings.TrimSpace(voice),
		"response_format": format,
	}
	if cmd.Flags().Changed("speed") {
		speed, _ := cmd.Flags().GetFloat64("speed")
		body["speed"] = speed
	}
	if provider, _ := cmd.Flags().GetString("provider"); strings.TrimSpace(provider) != "" {
		var providerConfig map[string]any
		if err := json.Unmarshal([]byte(provider), &providerConfig); err != nil {
			return fmt.Errorf("invalid --provider JSON: %w", err)
		}
		body["provider"] = providerConfig
	}

	req, _, err := openRouterJSONRequest(cmd.Context(), cmd, http.MethodPost, "/audio/speech", body, apiKey)
	if err != nil {
		return err
	}
	data, headers, _, err := openRouterDo(cmd, req, 5*time.Minute)
	if err != nil {
		return err
	}
	if client.IsDryRun(cmd) {
		return nil
	}
	ext := "." + format
	if format == "mp3" {
		ext = extensionForContentType(headers.Get("Content-Type"), ".mp3")
	}
	outputPath, _ := cmd.Flags().GetString("output")
	force, _ := cmd.Flags().GetBool("force")
	saved, err := writeWorkflowFile(outputPathForMedia(outputPath, "openrouter-audio", ext), data, force)
	if err != nil {
		return err
	}
	return output.Result(cmd, map[string]any{
		"model":       model,
		"voice":       voice,
		"format":      format,
		"audio_saved": saved,
		"bytes":       len(data),
		"key_source":  keySource,
	})
}

func runOpenRouterAudioTranscribeCommand(cmd *cobra.Command, args []string) error {
	apiKey, keySource, err := openRouterAPIKey(cmd, true)
	if err != nil {
		return err
	}
	model, _ := cmd.Flags().GetString("model")
	model = strings.TrimSpace(model)
	if model == "" || strings.EqualFold(model, "auto") {
		model, err = selectWorkflowModel(cmd, "transcription", preferredTranscriptionModels, defaultTranscriptionModel)
		if err != nil {
			return err
		}
	}
	audioPath := args[0]
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return fmt.Errorf("read audio file %s: %w", audioPath, err)
	}
	audioFormat, _ := cmd.Flags().GetString("format")
	audioFormat = strings.TrimSpace(audioFormat)
	if audioFormat == "" || strings.EqualFold(audioFormat, "auto") {
		audioFormat = strings.TrimPrefix(strings.ToLower(filepath.Ext(audioPath)), ".")
	}
	if audioFormat == "" {
		return fmt.Errorf("could not infer audio format; pass --format mp3, wav, m4a, or another supported format")
	}

	body := map[string]any{
		"model": model,
		"input_audio": map[string]any{
			"data":   base64.StdEncoding.EncodeToString(audioData),
			"format": audioFormat,
		},
	}
	if language, _ := cmd.Flags().GetString("language"); strings.TrimSpace(language) != "" {
		body["language"] = strings.TrimSpace(language)
	}
	if cmd.Flags().Changed("temperature") {
		temperature, _ := cmd.Flags().GetFloat64("temperature")
		body["temperature"] = temperature
	}
	if provider, _ := cmd.Flags().GetString("provider"); strings.TrimSpace(provider) != "" {
		var providerConfig map[string]any
		if err := json.Unmarshal([]byte(provider), &providerConfig); err != nil {
			return fmt.Errorf("invalid --provider JSON: %w", err)
		}
		body["provider"] = providerConfig
	}

	req, _, err := openRouterJSONRequest(cmd.Context(), cmd, http.MethodPost, "/audio/transcriptions", body, apiKey)
	if err != nil {
		return err
	}
	var response map[string]any
	raw, err := openRouterDoJSON(cmd, req, &response, 2*time.Minute)
	if err != nil {
		return err
	}
	if client.IsDryRun(cmd) {
		return nil
	}
	transcript := transcriptText(response)
	outputPath, _ := cmd.Flags().GetString("output")
	if strings.TrimSpace(outputPath) != "" {
		force, _ := cmd.Flags().GetBool("force")
		data := raw
		ext := ".json"
		if transcript != "" {
			data = []byte(transcript + "\n")
			ext = ".txt"
		}
		saved, err := writeWorkflowFile(outputPathForMedia(outputPath, "openrouter-transcript", ext), data, force)
		if err != nil {
			return err
		}
		response["transcript_saved"] = saved
	}
	rawOnly, _ := cmd.Flags().GetBool("raw")
	if rawOnly {
		if transcript == "" {
			return fmt.Errorf("transcription response did not include a text field")
		}
		_, err := fmt.Fprintln(cmd.OutOrStdout(), transcript)
		return err
	}
	response["model"] = model
	response["input_file"] = audioPath
	response["format"] = audioFormat
	response["key_source"] = keySource
	if transcript != "" {
		response["text"] = transcript
	}
	return output.Result(cmd, response)
}

func runOpenRouterAudioModelsCommand(cmd *cobra.Command, args []string) error {
	kind, _ := cmd.Flags().GetString("kind")
	kind = normalizeModelModality(kind)
	if kind != "speech" && kind != "transcription" {
		return fmt.Errorf("invalid --kind %q: expected speech or transcription", kind)
	}
	models, err := fetchModelsForWorkflow(cmd, kind)
	if err != nil {
		return err
	}
	if client.IsDryRun(cmd) {
		return nil
	}
	summaries := rankModelMatches(models, strings.TrimSpace(strings.Join(args, " ")), kind)
	limit, _ := cmd.Flags().GetInt("limit")
	if limit > 0 && len(summaries) > limit {
		summaries = summaries[:limit]
	}
	return output.Result(cmd, map[string]any{
		"kind":   kind,
		"count":  len(summaries),
		"models": summaries,
	})
}

func audioSpeakTextFromFlags(cmd *cobra.Command, args []string) (string, error) {
	values := []string{}
	if text, _ := cmd.Flags().GetString("text"); strings.TrimSpace(text) != "" {
		values = append(values, strings.TrimSpace(text))
	}
	if argText := strings.TrimSpace(strings.Join(args, " ")); argText != "" {
		values = append(values, argText)
	}
	if filePath, _ := cmd.Flags().GetString("file"); strings.TrimSpace(filePath) != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("read text file %s: %w", filePath, err)
		}
		values = append(values, strings.TrimSpace(string(data)))
	}
	if useStdin, _ := cmd.Flags().GetBool("stdin"); useStdin {
		data, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		values = append(values, strings.TrimSpace(string(data)))
	}
	nonEmpty := []string{}
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			nonEmpty = append(nonEmpty, strings.TrimSpace(value))
		}
	}
	if len(nonEmpty) > 1 {
		return "", fmt.Errorf("provide text through only one input source")
	}
	if len(nonEmpty) == 0 {
		return "", fmt.Errorf("missing text: run `openrouter audio speak \"hello\"`, use --text, --file, or --stdin")
	}
	return nonEmpty[0], nil
}

func transcriptText(response map[string]any) string {
	for _, key := range []string{"text", "transcript"} {
		if value, _ := response[key].(string); strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	if data, _ := response["data"].(map[string]any); data != nil {
		for _, key := range []string{"text", "transcript"} {
			if value, _ := data[key].(string); strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
	}
	return ""
}
