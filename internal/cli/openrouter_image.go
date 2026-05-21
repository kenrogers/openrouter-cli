package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kenrogers/openrouter-cli/internal/client"
	"github.com/kenrogers/openrouter-cli/internal/output"
	"github.com/spf13/cobra"
)

const defaultImageModel = "google/gemini-3.1-flash-image-preview"

var preferredImageModels = []string{
	"google/gemini-3.1-flash-image-preview",
	"google/gemini-2.5-flash-image",
	"google/gemini-2.5-flash-image-preview",
	"black-forest-labs/flux.2-pro",
	"black-forest-labs/flux.2-flex",
}

type openRouterModelListResponse struct {
	Data []openRouterModel `json:"data"`
}

type openRouterModel struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	ContextLength       int64    `json:"context_length,omitempty"`
	SupportedParameters []string `json:"supported_parameters,omitempty"`
	SupportedVoices     []string `json:"supported_voices,omitempty"`
	Created             int64    `json:"created,omitempty"`
	Architecture        struct {
		InputModalities  []string `json:"input_modalities,omitempty"`
		OutputModalities []string `json:"output_modalities,omitempty"`
	} `json:"architecture,omitempty"`
	Pricing map[string]string `json:"pricing,omitempty"`
}

type imageModelSummary struct {
	ID               string            `json:"id"`
	Name             string            `json:"name,omitempty"`
	InputModalities  []string          `json:"input_modalities,omitempty"`
	OutputModalities []string          `json:"output_modalities,omitempty"`
	ContextLength    int64             `json:"context_length,omitempty"`
	Pricing          map[string]string `json:"pricing,omitempty"`
}

type imageGenerateResult struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	Mode        string   `json:"mode"`
	ImagesSaved []string `json:"images_saved"`
	Count       int      `json:"count"`
	Text        string   `json:"text,omitempty"`
	KeySource   string   `json:"key_source,omitempty"`
}

func newOpenRouterImageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image [prompt]",
		Short: "Generate or edit images",
		Long: `Generate or edit images through OpenRouter image-output models.

This is the friendly path for agents and developers. It wraps the underlying
chat-completions image workflow, saves returned base64 image data to files, and
returns machine-readable output without requiring hand-built JSON.`,
		Example: `  openrouter image "a tiny red robot, product photo style"
  openrouter image --model google/gemini-3.1-flash-image-preview --aspect-ratio 16:9 --output hero.png "a cinematic mountain sunrise"
  openrouter image --input-image avatar.png --output avatar-watercolor.png "turn this into a watercolor portrait"
  openrouter image models nano banana`,
		RunE: runOpenRouterImageCommand,
	}
	cmd.Flags().StringP("prompt", "p", "", "Image generation or edit prompt (can also be provided as positional args)")
	cmd.Flags().StringP("model", "m", "auto", "Image-capable model ID, or auto to choose a current image-output model")
	cmd.Flags().String("output", "", "Output file path or directory (default: openrouter-image-<timestamp>.<ext>)")
	cmd.Flags().StringArray("input-image", nil, "Input image path, URL, or data URL for image editing; repeat for multiple images")
	cmd.Flags().String("aspect-ratio", "", "Requested aspect ratio, such as 1:1, 16:9, 9:16, or 4:1")
	cmd.Flags().String("image-size", "", "Requested image size, such as 0.5K, 1K, 2K, or 4K")
	cmd.Flags().Float64("strength", -1, "Image edit strength for models that support it (0.0 to 1.0)")
	cmd.Flags().String("image-config", "", "Advanced image_config JSON object; explicit image flags override matching keys")
	cmd.Flags().Bool("force", false, "Overwrite output files if they already exist")
	cmd.AddCommand(newOpenRouterImageModelsCommand())
	return cmd
}

func newOpenRouterImageModelsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models [query]",
		Short: "List image generation models",
		Long:  "List OpenRouter models whose output modalities include image, optionally filtered by a search query.",
		Example: `  openrouter image models
  openrouter image models flux
  openrouter image models --limit 5 --json`,
		RunE: runOpenRouterImageModelsCommand,
	}
	cmd.Flags().Int("limit", 20, "Maximum number of models to show (0 for all)")
	return cmd
}

func runOpenRouterImageCommand(cmd *cobra.Command, args []string) error {
	prompt, err := imagePromptFromFlags(cmd, args)
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
		model, err = selectImageModel(cmd)
		if err != nil {
			return err
		}
	}

	inputs, _ := cmd.Flags().GetStringArray("input-image")
	messageContent, err := buildImageMessageContent(prompt, inputs)
	if err != nil {
		return err
	}
	imageConfig, err := buildImageConfig(cmd)
	if err != nil {
		return err
	}

	body := map[string]any{
		"model": model,
		"messages": []map[string]any{
			{
				"role":    "user",
				"content": messageContent,
			},
		},
		"modalities": []string{"image", "text"},
		"stream":     false,
	}
	if len(imageConfig) > 0 {
		body["image_config"] = imageConfig
	}

	req, _, err := openRouterJSONRequest(cmd.Context(), cmd, http.MethodPost, "/chat/completions", body, apiKey)
	if err != nil {
		return err
	}
	var response map[string]any
	if _, err := openRouterDoJSON(cmd, req, &response, 5*time.Minute); err != nil {
		return err
	}
	if client.IsDryRun(cmd) {
		return nil
	}

	imageURLs, text := extractChatImageURLs(response)
	if len(imageURLs) == 0 {
		return output.AgentModeError(cmd,
			"image_generation_empty",
			"OpenRouter response did not include generated images",
			[]string{
				"Run `openrouter image models` and choose a model with image in output_modalities",
				"Use `--debug` to inspect the response",
			},
		)
	}
	force, _ := cmd.Flags().GetBool("force")
	outputPath, _ := cmd.Flags().GetString("output")
	saved, err := saveGeneratedImages(imageURLs, outputPath, force)
	if err != nil {
		return err
	}

	mode := "generate"
	if len(inputs) > 0 {
		mode = "edit"
	}
	result := map[string]any{
		"model":        model,
		"prompt":       prompt,
		"mode":         mode,
		"images_saved": saved,
		"count":        len(saved),
		"key_source":   keySource,
	}
	if text != "" {
		result["text"] = text
	}
	return output.Result(cmd, result)
}

func runOpenRouterImageModelsCommand(cmd *cobra.Command, args []string) error {
	query := strings.ToLower(strings.TrimSpace(strings.Join(args, " ")))
	models, err := fetchImageModels(cmd)
	if err != nil {
		return err
	}
	if client.IsDryRun(cmd) {
		return nil
	}
	summaries := make([]imageModelSummary, 0, len(models))
	for _, model := range models {
		if query != "" && !strings.Contains(strings.ToLower(model.ID), query) && !strings.Contains(strings.ToLower(model.Name), query) {
			continue
		}
		summaries = append(summaries, imageModelSummary{
			ID:               model.ID,
			Name:             model.Name,
			InputModalities:  model.Architecture.InputModalities,
			OutputModalities: model.Architecture.OutputModalities,
			ContextLength:    model.ContextLength,
			Pricing:          model.Pricing,
		})
	}
	total := len(summaries)
	limit, _ := cmd.Flags().GetInt("limit")
	if limit > 0 && len(summaries) > limit {
		summaries = summaries[:limit]
	}
	return output.Result(cmd, map[string]any{
		"count":  len(summaries),
		"total":  total,
		"models": summaries,
	})
}

func imagePromptFromFlags(cmd *cobra.Command, args []string) (string, error) {
	flagPrompt, _ := cmd.Flags().GetString("prompt")
	flagPrompt = strings.TrimSpace(flagPrompt)
	argPrompt := strings.TrimSpace(strings.Join(args, " "))
	if flagPrompt != "" && argPrompt != "" {
		return "", fmt.Errorf("provide the prompt either as positional text or --prompt, not both")
	}
	prompt := firstNonEmpty(flagPrompt, argPrompt)
	if prompt == "" {
		return "", fmt.Errorf("missing prompt: run `openrouter image \"your prompt\"` or use --prompt")
	}
	return prompt, nil
}

func buildImageMessageContent(prompt string, inputs []string) (any, error) {
	if len(inputs) == 0 {
		return prompt, nil
	}
	parts := []map[string]any{{"type": "text", "text": prompt}}
	for _, input := range inputs {
		imageURL, err := imageInputToURL(strings.TrimSpace(input))
		if err != nil {
			return nil, err
		}
		parts = append(parts, map[string]any{
			"type": "image_url",
			"image_url": map[string]any{
				"url": imageURL,
			},
		})
	}
	return parts, nil
}

func buildImageConfig(cmd *cobra.Command) (map[string]any, error) {
	config := map[string]any{}
	raw, _ := cmd.Flags().GetString("image-config")
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &config); err != nil {
			return nil, fmt.Errorf("invalid --image-config JSON: %w", err)
		}
	}
	if aspectRatio, _ := cmd.Flags().GetString("aspect-ratio"); strings.TrimSpace(aspectRatio) != "" {
		config["aspect_ratio"] = strings.TrimSpace(aspectRatio)
	}
	if imageSize, _ := cmd.Flags().GetString("image-size"); strings.TrimSpace(imageSize) != "" {
		config["image_size"] = strings.TrimSpace(imageSize)
	}
	if strength, _ := cmd.Flags().GetFloat64("strength"); strength >= 0 {
		if strength > 1 {
			return nil, fmt.Errorf("--strength must be between 0.0 and 1.0")
		}
		config["strength"] = strength
	}
	return config, nil
}

func selectImageModel(cmd *cobra.Command) (string, error) {
	if client.IsDryRun(cmd) {
		return defaultImageModel, nil
	}
	models, err := fetchImageModels(cmd)
	if err != nil {
		return "", err
	}
	for _, preferred := range preferredImageModels {
		for _, model := range models {
			if model.ID == preferred && hasModality(model.Architecture.OutputModalities, "image") {
				return model.ID, nil
			}
		}
	}
	for _, model := range models {
		if hasModality(model.Architecture.OutputModalities, "image") {
			return model.ID, nil
		}
	}
	return "", fmt.Errorf("no image-output models found")
}

func fetchImageModels(cmd *cobra.Command) ([]openRouterModel, error) {
	apiKey, _, _ := openRouterAPIKey(cmd, false)
	path := "/models?" + url.Values{"output_modalities": []string{"image"}}.Encode()
	req, _, err := openRouterJSONRequest(cmd.Context(), cmd, http.MethodGet, path, nil, apiKey)
	if err != nil {
		return nil, err
	}
	var response openRouterModelListResponse
	if _, err := openRouterDoJSON(cmd, req, &response, 30*time.Second); err != nil {
		return nil, err
	}
	if client.IsDryRun(cmd) {
		return nil, nil
	}
	return response.Data, nil
}

func hasModality(values []string, want string) bool {
	for _, value := range values {
		if strings.EqualFold(value, want) {
			return true
		}
	}
	return false
}

func imageInputToURL(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("empty --input-image value")
	}
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") || strings.HasPrefix(input, "data:") {
		return input, nil
	}
	data, err := os.ReadFile(input)
	if err != nil {
		return "", fmt.Errorf("read input image %s: %w", input, err)
	}
	mediaType := mime.TypeByExtension(strings.ToLower(filepath.Ext(input)))
	if mediaType == "" {
		mediaType = http.DetectContentType(data)
	}
	return "data:" + mediaType + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}

func extractChatImageURLs(response map[string]any) ([]string, string) {
	var imageURLs []string
	var textParts []string
	choices, _ := response["choices"].([]any)
	for _, choice := range choices {
		choiceMap, _ := choice.(map[string]any)
		message, _ := choiceMap["message"].(map[string]any)
		if contentText := textFromChatContent(message["content"]); contentText != "" {
			textParts = append(textParts, contentText)
		}
		if images, ok := message["images"].([]any); ok {
			for _, image := range images {
				if imageURL := imageURLFromValue(image); imageURL != "" {
					imageURLs = append(imageURLs, imageURL)
				}
			}
		}
		if contentImages := imageURLsFromChatContent(message["content"]); len(contentImages) > 0 {
			imageURLs = append(imageURLs, contentImages...)
		}
	}
	return imageURLs, strings.TrimSpace(strings.Join(textParts, "\n"))
}

func textFromChatContent(content any) string {
	switch value := content.(type) {
	case string:
		return strings.TrimSpace(value)
	case []any:
		var parts []string
		for _, part := range value {
			partMap, _ := part.(map[string]any)
			if text, _ := partMap["text"].(string); strings.TrimSpace(text) != "" {
				parts = append(parts, strings.TrimSpace(text))
			}
		}
		return strings.Join(parts, "\n")
	default:
		return ""
	}
}

func imageURLsFromChatContent(content any) []string {
	var urls []string
	parts, ok := content.([]any)
	if !ok {
		return urls
	}
	for _, part := range parts {
		if imageURL := imageURLFromValue(part); imageURL != "" {
			urls = append(urls, imageURL)
		}
	}
	return urls
}

func imageURLFromValue(value any) string {
	item, _ := value.(map[string]any)
	if direct, _ := item["url"].(string); direct != "" {
		return direct
	}
	imageURL, _ := item["image_url"].(map[string]any)
	if urlValue, _ := imageURL["url"].(string); urlValue != "" {
		return urlValue
	}
	return ""
}

func saveGeneratedImages(imageURLs []string, outputPath string, force bool) ([]string, error) {
	saved := make([]string, 0, len(imageURLs))
	for i, imageURL := range imageURLs {
		data, mediaType, err := decodeImageDataURL(imageURL)
		if err != nil {
			return nil, err
		}
		path := outputPathForImage(outputPath, mediaType, i, len(imageURLs))
		if err := writeImageFile(path, data, force); err != nil {
			return nil, err
		}
		absolute, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		saved = append(saved, absolute)
	}
	return saved, nil
}

func decodeImageDataURL(dataURL string) ([]byte, string, error) {
	if !strings.HasPrefix(dataURL, "data:") {
		return nil, "", fmt.Errorf("generated image URL was not a data URL")
	}
	header, payload, ok := strings.Cut(dataURL, ",")
	if !ok || !strings.Contains(header, ";base64") {
		return nil, "", fmt.Errorf("generated image data URL was malformed")
	}
	mediaType := strings.TrimPrefix(strings.TrimSuffix(header, ";base64"), "data:")
	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, "", fmt.Errorf("decode generated image: %w", err)
	}
	return data, mediaType, nil
}

func outputPathForImage(outputPath, mediaType string, index, total int) string {
	ext := extensionForImageMediaType(mediaType)
	if strings.TrimSpace(outputPath) == "" {
		name := "openrouter-image-" + time.Now().UTC().Format("20060102-150405")
		if total > 1 {
			name = fmt.Sprintf("%s-%d", name, index+1)
		}
		return name + ext
	}
	info, err := os.Stat(outputPath)
	if err == nil && info.IsDir() {
		name := "openrouter-image-" + time.Now().UTC().Format("20060102-150405")
		if total > 1 {
			name = fmt.Sprintf("%s-%d", name, index+1)
		}
		return filepath.Join(outputPath, name+ext)
	}
	dir := filepath.Dir(outputPath)
	base := strings.TrimSuffix(filepath.Base(outputPath), filepath.Ext(outputPath))
	fileExt := filepath.Ext(outputPath)
	if fileExt == "" {
		fileExt = ext
	}
	if total > 1 {
		return filepath.Join(dir, fmt.Sprintf("%s-%d%s", base, index+1, fileExt))
	}
	return filepath.Join(dir, base+fileExt)
}

func extensionForImageMediaType(mediaType string) string {
	switch strings.ToLower(mediaType) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".png"
	}
}

func writeImageFile(path string, data []byte, force bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists; pass --force to overwrite", path)
		}
	}
	return os.WriteFile(path, data, 0o600)
}
