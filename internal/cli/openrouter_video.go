package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"strings"
	"time"

	"github.com/kenrogers/openrouter-cli/internal/client"
	"github.com/kenrogers/openrouter-cli/internal/output"
	"github.com/spf13/cobra"
)

type openRouterVideoModelListResponse struct {
	Data []openRouterVideoModel `json:"data"`
}

type openRouterVideoModel struct {
	ID                           string   `json:"id"`
	Name                         string   `json:"name,omitempty"`
	SupportedAspectRatios        []string `json:"supported_aspect_ratios,omitempty"`
	SupportedDurations           []int64  `json:"supported_durations,omitempty"`
	SupportedFrameImages         []string `json:"supported_frame_images,omitempty"`
	SupportedInputReferences     []string `json:"supported_input_references,omitempty"`
	SupportedResolutions         []string `json:"supported_resolutions,omitempty"`
	SupportedSizes               []string `json:"supported_sizes,omitempty"`
	GenerateAudio                *bool    `json:"generate_audio,omitempty"`
	Seed                         *bool    `json:"seed,omitempty"`
	AllowedPassthroughParameters []string `json:"allowed_passthrough_parameters,omitempty"`
	PricingSKUs                  any      `json:"pricing_skus,omitempty"`
}

type openRouterVideoModelSummary struct {
	ID                           string   `json:"id"`
	Name                         string   `json:"name,omitempty"`
	SupportedAspectRatios        []string `json:"supported_aspect_ratios,omitempty"`
	SupportedDurations           []int64  `json:"supported_durations,omitempty"`
	SupportedFrameImages         []string `json:"supported_frame_images,omitempty"`
	SupportedInputReferences     []string `json:"supported_input_references,omitempty"`
	SupportedResolutions         []string `json:"supported_resolutions,omitempty"`
	SupportedSizes               []string `json:"supported_sizes,omitempty"`
	GenerateAudio                *bool    `json:"generate_audio,omitempty"`
	Seed                         *bool    `json:"seed,omitempty"`
	AllowedPassthroughParameters []string `json:"allowed_passthrough_parameters,omitempty"`
	PricingSKUs                  any      `json:"pricing_skus,omitempty"`
}

func newOpenRouterVideoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "video [prompt]",
		Short: "Generate video from text and images",
		Long: `Generate a video through OpenRouter's async video API.

By default the command submits the job, waits for completion, downloads the
first video, and returns the saved path. Use --no-wait to only submit and return
the job ID.`,
		Example: `  openrouter video "a calm product shot of a glass keyboard on a walnut desk"
  openrouter video --model google/veo-3.1-lite --duration 4 --resolution 720p --aspect-ratio 16:9 "clouds over Denver"
  openrouter video --first-frame start.png --last-frame end.png --output clip.mp4 "animate this transition"
  openrouter video models veo`,
		RunE: runOpenRouterVideoCommand,
	}
	cmd.Flags().StringP("prompt", "p", "", "Video prompt (can also be provided as positional args)")
	cmd.Flags().StringP("file", "f", "", "Read prompt text from a file")
	cmd.Flags().Bool("stdin", false, "Read prompt text from stdin")
	cmd.Flags().StringP("model", "m", "auto", "Video model ID, or auto to choose a current video model")
	cmd.Flags().Int64("duration", 0, "Duration in seconds")
	cmd.Flags().String("resolution", "", "Resolution such as 480p, 720p, 1080p, 1K, 2K, or 4K")
	cmd.Flags().String("aspect-ratio", "", "Aspect ratio such as 16:9, 9:16, or 1:1")
	cmd.Flags().String("size", "", "Exact size such as 1280x720")
	cmd.Flags().Bool("generate-audio", false, "Generate audio when the selected model supports it")
	cmd.Flags().Int64("seed", 0, "Deterministic seed for models that support it")
	cmd.Flags().String("callback-url", "", "HTTPS callback URL for completion notification")
	cmd.Flags().String("first-frame", "", "First-frame image path, URL, or data URL")
	cmd.Flags().String("last-frame", "", "Last-frame image path, URL, or data URL")
	cmd.Flags().StringArray("reference-image", nil, "Reference image path, URL, or data URL; repeat for multiple references")
	cmd.Flags().String("provider", "", "Provider routing JSON object")
	cmd.Flags().Bool("no-wait", false, "Submit the job and return without polling")
	cmd.Flags().Duration("poll-interval", 5*time.Second, "Polling interval while waiting for completion")
	cmd.Flags().Duration("wait-timeout", 15*time.Minute, "Maximum time to wait for completion")
	cmd.Flags().String("output", "", "Output file path or directory (default: openrouter-video-<timestamp>.mp4)")
	cmd.Flags().Bool("force", false, "Overwrite output files if they already exist")
	cmd.AddCommand(newOpenRouterVideoModelsCommand())
	return cmd
}

func newOpenRouterVideoModelsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models [query]",
		Short: "List video generation models",
		Long:  "List OpenRouter video generation models, optionally filtered by a query.",
		RunE:  runOpenRouterVideoModelsCommand,
	}
	cmd.Flags().Int("limit", 20, "Maximum number of models to show (0 for all)")
	return cmd
}

func runOpenRouterVideoCommand(cmd *cobra.Command, args []string) error {
	prompt, err := textPromptFromFlags(cmd, args, "video")
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
		model, err = selectWorkflowModel(cmd, "video", preferredVideoModels, defaultVideoModel)
		if err != nil {
			return err
		}
	}
	body, err := buildVideoRequestBody(cmd, model, prompt)
	if err != nil {
		return err
	}
	req, _, err := openRouterJSONRequest(cmd.Context(), cmd, http.MethodPost, "/videos", body, apiKey)
	if err != nil {
		return err
	}
	var submitted map[string]any
	if _, err := openRouterDoJSON(cmd, req, &submitted, 2*time.Minute); err != nil {
		return err
	}
	if client.IsDryRun(cmd) {
		return nil
	}
	jobID := videoJobID(submitted)
	result := map[string]any{
		"model":      model,
		"prompt":     prompt,
		"job_id":     jobID,
		"status":     stringValue(submitted, "status"),
		"key_source": keySource,
	}
	if pollingURL := stringValue(submitted, "polling_url"); pollingURL != "" {
		result["polling_url"] = pollingURL
	}

	noWait, _ := cmd.Flags().GetBool("no-wait")
	if noWait {
		return output.Result(cmd, result)
	}
	if jobID == "" {
		return output.AgentModeError(cmd,
			"video_job_missing",
			"OpenRouter did not return a video job ID",
			[]string{"Run with `--debug` to inspect the response", "Use `openrouter video-generation generate` for the raw endpoint"},
		)
	}
	completed, err := waitForVideoCompletion(cmd, apiKey, jobID, submitted)
	if err != nil {
		return err
	}
	result["status"] = stringValue(completed, "status")
	result["generation"] = completed
	videoData, headers, err := downloadVideoResult(cmd, apiKey, jobID, completed)
	if err != nil {
		return err
	}
	outputPath, _ := cmd.Flags().GetString("output")
	force, _ := cmd.Flags().GetBool("force")
	ext := extensionForContentType(headers.Get("Content-Type"), ".mp4")
	saved, err := writeWorkflowFile(outputPathForMedia(outputPath, "openrouter-video", ext), videoData, force)
	if err != nil {
		return err
	}
	result["video_saved"] = saved
	result["bytes"] = len(videoData)
	return output.Result(cmd, result)
}

func runOpenRouterVideoModelsCommand(cmd *cobra.Command, args []string) error {
	models, err := fetchVideoModels(cmd)
	if err != nil {
		return err
	}
	if client.IsDryRun(cmd) {
		return nil
	}
	query := strings.ToLower(strings.TrimSpace(strings.Join(args, " ")))
	summaries := make([]openRouterVideoModelSummary, 0, len(models))
	for _, model := range models {
		if query != "" && !strings.Contains(strings.ToLower(model.ID), query) && !strings.Contains(strings.ToLower(model.Name), query) {
			continue
		}
		summaries = append(summaries, summarizeVideoModel(model))
	}
	limit, _ := cmd.Flags().GetInt("limit")
	if limit > 0 && len(summaries) > limit {
		summaries = summaries[:limit]
	}
	return output.Result(cmd, map[string]any{
		"count":  len(summaries),
		"models": summaries,
	})
}

func buildVideoRequestBody(cmd *cobra.Command, model, prompt string) (map[string]any, error) {
	body := map[string]any{
		"model":  model,
		"prompt": prompt,
	}
	if cmd.Flags().Changed("duration") {
		duration, _ := cmd.Flags().GetInt64("duration")
		body["duration"] = duration
	}
	if resolution, _ := cmd.Flags().GetString("resolution"); strings.TrimSpace(resolution) != "" {
		body["resolution"] = strings.TrimSpace(resolution)
	}
	if aspectRatio, _ := cmd.Flags().GetString("aspect-ratio"); strings.TrimSpace(aspectRatio) != "" {
		body["aspect_ratio"] = strings.TrimSpace(aspectRatio)
	}
	if size, _ := cmd.Flags().GetString("size"); strings.TrimSpace(size) != "" {
		body["size"] = strings.TrimSpace(size)
	}
	if cmd.Flags().Changed("generate-audio") {
		generateAudio, _ := cmd.Flags().GetBool("generate-audio")
		body["generate_audio"] = generateAudio
	}
	if cmd.Flags().Changed("seed") {
		seed, _ := cmd.Flags().GetInt64("seed")
		body["seed"] = seed
	}
	if callbackURL, _ := cmd.Flags().GetString("callback-url"); strings.TrimSpace(callbackURL) != "" {
		body["callback_url"] = strings.TrimSpace(callbackURL)
	}
	frameImages, err := videoFrameImagesFromFlags(cmd)
	if err != nil {
		return nil, err
	}
	if len(frameImages) > 0 {
		body["frame_images"] = frameImages
	}
	references, err := videoReferencesFromFlags(cmd)
	if err != nil {
		return nil, err
	}
	if len(references) > 0 {
		body["input_references"] = references
	}
	if provider, _ := cmd.Flags().GetString("provider"); strings.TrimSpace(provider) != "" {
		var providerConfig map[string]any
		if err := json.Unmarshal([]byte(provider), &providerConfig); err != nil {
			return nil, fmt.Errorf("invalid --provider JSON: %w", err)
		}
		body["provider"] = providerConfig
	}
	return body, nil
}

func videoFrameImagesFromFlags(cmd *cobra.Command) ([]map[string]any, error) {
	var frames []map[string]any
	if firstFrame, _ := cmd.Flags().GetString("first-frame"); strings.TrimSpace(firstFrame) != "" {
		imageURL, err := imageInputToURL(firstFrame)
		if err != nil {
			return nil, err
		}
		frames = append(frames, map[string]any{"frame_type": "first_frame", "image_url": imageURL})
	}
	if lastFrame, _ := cmd.Flags().GetString("last-frame"); strings.TrimSpace(lastFrame) != "" {
		imageURL, err := imageInputToURL(lastFrame)
		if err != nil {
			return nil, err
		}
		frames = append(frames, map[string]any{"frame_type": "last_frame", "image_url": imageURL})
	}
	return frames, nil
}

func videoReferencesFromFlags(cmd *cobra.Command) ([]map[string]any, error) {
	inputs, _ := cmd.Flags().GetStringArray("reference-image")
	references := make([]map[string]any, 0, len(inputs))
	for _, input := range inputs {
		imageURL, err := imageInputToURL(input)
		if err != nil {
			return nil, err
		}
		references = append(references, map[string]any{"image_url": imageURL})
	}
	return references, nil
}

func waitForVideoCompletion(cmd *cobra.Command, apiKey, jobID string, initial map[string]any) (map[string]any, error) {
	status := strings.ToLower(stringValue(initial, "status"))
	if status == "completed" {
		return initial, nil
	}
	pollInterval, _ := cmd.Flags().GetDuration("poll-interval")
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}
	waitTimeout, _ := cmd.Flags().GetDuration("wait-timeout")
	if waitTimeout <= 0 {
		waitTimeout = 15 * time.Minute
	}
	deadline := time.Now().Add(waitTimeout)
	current := initial
	for {
		status = strings.ToLower(stringValue(current, "status"))
		switch status {
		case "completed":
			return current, nil
		case "failed", "cancelled", "expired":
			return nil, fmt.Errorf("video generation %s: %s", status, videoErrorMessage(current))
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out waiting for video job %s after %s", jobID, waitTimeout)
		}
		time.Sleep(pollInterval)
		polled, err := getVideoGeneration(cmd, apiKey, jobID)
		if err != nil {
			return nil, err
		}
		current = polled
	}
}

func getVideoGeneration(cmd *cobra.Command, apiKey, jobID string) (map[string]any, error) {
	req, _, err := openRouterJSONRequest(cmd.Context(), cmd, http.MethodGet, "/videos/"+urlpkg.PathEscape(jobID), nil, apiKey)
	if err != nil {
		return nil, err
	}
	var response map[string]any
	if _, err := openRouterDoJSON(cmd, req, &response, 30*time.Second); err != nil {
		return nil, err
	}
	return response, nil
}

func downloadVideoResult(cmd *cobra.Command, apiKey, jobID string, generation map[string]any) ([]byte, http.Header, error) {
	data, headers, err := downloadVideoContentEndpoint(cmd, apiKey, jobID)
	if err == nil {
		return data, headers, nil
	}
	for _, unsignedURL := range stringSliceValue(generation, "unsigned_urls") {
		data, headers, fallbackErr := downloadURL(cmd, unsignedURL, "")
		if fallbackErr == nil {
			return data, headers, nil
		}
	}
	return nil, nil, err
}

func downloadVideoContentEndpoint(cmd *cobra.Command, apiKey, jobID string) ([]byte, http.Header, error) {
	path := "/videos/" + urlpkg.PathEscape(jobID) + "/content?index=0"
	req, err := openRouterBinaryRequest(cmd, openRouterAPIBase(cmd)+path, apiKey)
	if err != nil {
		return nil, nil, err
	}
	data, headers, _, err := openRouterDo(cmd, req, 10*time.Minute)
	return data, headers, err
}

func downloadURL(cmd *cobra.Command, rawURL, apiKey string) ([]byte, http.Header, error) {
	req, err := openRouterBinaryRequest(cmd, rawURL, apiKey)
	if err != nil {
		return nil, nil, err
	}
	data, headers, _, err := openRouterDo(cmd, req, 10*time.Minute)
	return data, headers, err
}

func openRouterBinaryRequest(cmd *cobra.Command, rawURL, apiKey string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	if err := openRouterSetCommonHeaders(cmd, req, apiKey); err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "*/*")
	return req, nil
}

func fetchVideoModelsAsOpenRouterModels(cmd *cobra.Command) ([]openRouterModel, error) {
	videoModels, err := fetchVideoModels(cmd)
	if err != nil {
		return nil, err
	}
	models := make([]openRouterModel, 0, len(videoModels))
	for _, videoModel := range videoModels {
		var model openRouterModel
		model.ID = videoModel.ID
		model.Name = videoModel.Name
		model.SupportedParameters = videoModel.AllowedPassthroughParameters
		model.Architecture.InputModalities = []string{"text"}
		if len(videoModel.SupportedFrameImages) > 0 || len(videoModel.SupportedInputReferences) > 0 {
			model.Architecture.InputModalities = append(model.Architecture.InputModalities, "image")
		}
		model.Architecture.OutputModalities = []string{"video"}
		models = append(models, model)
	}
	return models, nil
}

func fetchVideoModels(cmd *cobra.Command) ([]openRouterVideoModel, error) {
	apiKey, _, _ := openRouterAPIKey(cmd, false)
	req, _, err := openRouterJSONRequest(cmd.Context(), cmd, http.MethodGet, "/videos/models", nil, apiKey)
	if err != nil {
		return nil, err
	}
	var response openRouterVideoModelListResponse
	if _, err := openRouterDoJSON(cmd, req, &response, 30*time.Second); err != nil {
		return nil, err
	}
	if client.IsDryRun(cmd) {
		return nil, nil
	}
	return response.Data, nil
}

func summarizeVideoModel(model openRouterVideoModel) openRouterVideoModelSummary {
	return openRouterVideoModelSummary{
		ID:                           model.ID,
		Name:                         model.Name,
		SupportedAspectRatios:        model.SupportedAspectRatios,
		SupportedDurations:           model.SupportedDurations,
		SupportedFrameImages:         model.SupportedFrameImages,
		SupportedInputReferences:     model.SupportedInputReferences,
		SupportedResolutions:         model.SupportedResolutions,
		SupportedSizes:               model.SupportedSizes,
		GenerateAudio:                model.GenerateAudio,
		Seed:                         model.Seed,
		AllowedPassthroughParameters: model.AllowedPassthroughParameters,
		PricingSKUs:                  model.PricingSKUs,
	}
}

func videoJobID(response map[string]any) string {
	for _, key := range []string{"id", "generation_id", "job_id"} {
		if value := stringValue(response, key); value != "" {
			return value
		}
	}
	return ""
}

func videoErrorMessage(response map[string]any) string {
	if errorValue, _ := response["error"].(string); strings.TrimSpace(errorValue) != "" {
		return errorValue
	}
	if errorMap, _ := response["error"].(map[string]any); errorMap != nil {
		if message, _ := errorMap["message"].(string); strings.TrimSpace(message) != "" {
			return message
		}
	}
	return "OpenRouter did not include an error message"
}

func stringValue(values map[string]any, key string) string {
	if value, _ := values[key].(string); strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return ""
}

func stringSliceValue(values map[string]any, key string) []string {
	raw, _ := values[key].([]any)
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if value, _ := item.(string); strings.TrimSpace(value) != "" {
			out = append(out, strings.TrimSpace(value))
		}
	}
	return out
}
