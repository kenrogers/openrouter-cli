package cli

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/kenrogers/openrouter-cli/internal/client"
	"github.com/kenrogers/openrouter-cli/internal/output"
	"github.com/spf13/cobra"
)

const (
	defaultTextModel          = "google/gemini-3.1-flash-lite"
	defaultSpeechModel        = "openai/gpt-4o-mini-tts-2025-12-15"
	defaultTranscriptionModel = "google/chirp-3"
	defaultEmbeddingModel     = "google/gemini-embedding-2-preview"
	defaultRerankModel        = "cohere/rerank-v3.5"
	defaultVideoModel         = "google/veo-3.1-lite"
)

var preferredTextModels = []string{
	"google/gemini-3.1-flash-lite",
	"google/gemini-3.5-flash",
	"openai/gpt-5.4-mini",
	"anthropic/claude-sonnet-4.6",
}

var preferredSpeechModels = []string{
	"openai/gpt-4o-mini-tts-2025-12-15",
	"openai/gpt-4o-mini-tts",
	"elevenlabs/eleven-flash-v2.5",
	"elevenlabs/eleven-turbo-v2.5",
}

var preferredTranscriptionModels = []string{
	"google/chirp-3",
	"openai/whisper-large-v3",
	"openai/whisper-1",
}

var preferredEmbeddingModels = []string{
	"google/gemini-embedding-2-preview",
	"perplexity/pplx-embed-v1-4b",
	"perplexity/pplx-embed-v1-0.6b",
}

var preferredVideoModels = []string{
	"google/veo-3.1-lite",
	"google/veo-3.1",
	"bytedance/seedance-2.0-fast",
	"alibaba/wan-2.7",
}

type openRouterModelSummary struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name,omitempty"`
	InputModalities     []string          `json:"input_modalities,omitempty"`
	OutputModalities    []string          `json:"output_modalities,omitempty"`
	ContextLength       int64             `json:"context_length,omitempty"`
	SupportedParameters []string          `json:"supported_parameters,omitempty"`
	SupportedVoices     []string          `json:"supported_voices,omitempty"`
	Pricing             map[string]string `json:"pricing,omitempty"`
	Score               int               `json:"score,omitempty"`
}

func enhanceOpenRouterModelsCommand(parent *cobra.Command) {
	modelsCmd := findChildCommand(parent, "models")
	if modelsCmd == nil {
		return
	}
	modelsCmd.AddCommand(newOpenRouterModelsSearchCommand())
	modelsCmd.AddCommand(newOpenRouterModelsResolveCommand())
}

func newOpenRouterModelsSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search OpenRouter models by name, ID, and modality",
		Long:  "Search OpenRouter models and return compact, agent-friendly model summaries.",
		Example: `  openrouter models search sonnet --modality text
  openrouter models search veo --modality video
  openrouter models search embed --modality embeddings --json`,
		RunE: runOpenRouterModelsSearchCommand,
	}
	cmd.Flags().String("modality", "all", "Output modality to search: all, text, image, audio, speech, transcription, embeddings, or video")
	cmd.Flags().Int("limit", 20, "Maximum number of matches to return (0 for all)")
	return cmd
}

func newOpenRouterModelsResolveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve [query]",
		Short: "Resolve a fuzzy model query to a concrete model ID",
		Long:  "Resolve a human model query to the best matching OpenRouter model ID and a few alternatives.",
		Example: `  openrouter models resolve sonnet --modality text
  openrouter models resolve nano banana --modality image
  openrouter models resolve veo --modality video`,
		RunE: runOpenRouterModelsResolveCommand,
	}
	cmd.Flags().String("modality", "text", "Output modality to resolve: all, text, image, audio, speech, transcription, embeddings, or video")
	cmd.Flags().Int("limit", 5, "Number of candidate matches to include")
	return cmd
}

func runOpenRouterModelsSearchCommand(cmd *cobra.Command, args []string) error {
	query := strings.TrimSpace(strings.Join(args, " "))
	modality, _ := cmd.Flags().GetString("modality")
	models, err := fetchModelsForWorkflow(cmd, modality)
	if err != nil {
		return err
	}
	if client.IsDryRun(cmd) {
		return nil
	}
	matches := rankModelMatches(models, query, modality)
	limit, _ := cmd.Flags().GetInt("limit")
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}
	return output.Result(cmd, map[string]any{
		"query":    query,
		"modality": normalizeModelModality(modality),
		"count":    len(matches),
		"models":   matches,
	})
}

func runOpenRouterModelsResolveCommand(cmd *cobra.Command, args []string) error {
	query := strings.TrimSpace(strings.Join(args, " "))
	if query == "" {
		return fmt.Errorf("missing query: run `openrouter models resolve sonnet --modality text`")
	}
	modality, _ := cmd.Flags().GetString("modality")
	models, err := fetchModelsForWorkflow(cmd, modality)
	if err != nil {
		return err
	}
	if client.IsDryRun(cmd) {
		return nil
	}
	matches := rankModelMatches(models, query, modality)
	if len(matches) == 0 {
		return output.AgentModeError(cmd,
			"model_not_found",
			fmt.Sprintf("No OpenRouter model matched %q", query),
			[]string{
				"Run `openrouter models search --modality all` to inspect available models",
				"Use a concrete model ID with the workflow command's --model flag",
			},
		)
	}
	limit, _ := cmd.Flags().GetInt("limit")
	if limit <= 0 {
		limit = 5
	}
	if len(matches) > limit {
		matches = matches[:limit]
	}
	return output.Result(cmd, map[string]any{
		"query":       query,
		"modality":    normalizeModelModality(modality),
		"model":       matches[0].ID,
		"name":        matches[0].Name,
		"confidence":  modelConfidence(matches[0].Score),
		"score":       matches[0].Score,
		"alternates":  matches[1:],
		"recommended": fmt.Sprintf("--model %s", matches[0].ID),
	})
}

func fetchModelsForWorkflow(cmd *cobra.Command, modality string) ([]openRouterModel, error) {
	normalized := normalizeModelModality(modality)
	if normalized == "video" {
		return fetchVideoModelsAsOpenRouterModels(cmd)
	}
	apiKey, _, _ := openRouterAPIKey(cmd, false)
	path := "/models"
	if normalized != "" && normalized != "all" {
		queryModality := normalized
		switch normalized {
		case "speech", "transcription":
			queryModality = "audio"
		}
		path += "?" + url.Values{"output_modalities": []string{queryModality}}.Encode()
	} else if normalized == "all" {
		path += "?" + url.Values{"output_modalities": []string{"all"}}.Encode()
	}
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
	return filterModelsByWorkflowModality(response.Data, normalized), nil
}

func selectWorkflowModel(cmd *cobra.Command, modality string, preferred []string, fallback string) (string, error) {
	if client.IsDryRun(cmd) {
		return fallback, nil
	}
	models, err := fetchModelsForWorkflow(cmd, modality)
	if err != nil {
		return "", err
	}
	for _, preferredID := range preferred {
		for _, model := range models {
			if model.ID == preferredID {
				return model.ID, nil
			}
		}
	}
	if len(models) > 0 {
		return models[0].ID, nil
	}
	if fallback != "" {
		return fallback, nil
	}
	return "", fmt.Errorf("no %s models found", normalizeModelModality(modality))
}

func filterModelsByWorkflowModality(models []openRouterModel, modality string) []openRouterModel {
	if modality == "" || modality == "all" {
		return models
	}
	filtered := make([]openRouterModel, 0, len(models))
	for _, model := range models {
		if modelMatchesWorkflowModality(model, modality) {
			filtered = append(filtered, model)
		}
	}
	return filtered
}

func modelMatchesWorkflowModality(model openRouterModel, modality string) bool {
	switch normalizeModelModality(modality) {
	case "", "all":
		return true
	case "speech":
		return hasModality(model.Architecture.OutputModalities, "audio") || strings.Contains(strings.ToLower(model.ID+" "+model.Name), "tts")
	case "transcription":
		return hasModality(model.Architecture.InputModalities, "audio") || strings.Contains(strings.ToLower(model.ID+" "+model.Name), "whisper") || strings.Contains(strings.ToLower(model.ID+" "+model.Name), "chirp")
	default:
		return hasModality(model.Architecture.OutputModalities, normalizeModelModality(modality))
	}
}

func normalizeModelModality(modality string) string {
	value := strings.ToLower(strings.TrimSpace(modality))
	switch value {
	case "", "any", "*":
		return "all"
	case "embed", "embedding":
		return "embeddings"
	case "tts":
		return "speech"
	case "stt":
		return "transcription"
	case "images":
		return "image"
	case "videos":
		return "video"
	default:
		return value
	}
}

func rankModelMatches(models []openRouterModel, query, modality string) []openRouterModelSummary {
	query = strings.ToLower(strings.TrimSpace(query))
	summaries := make([]openRouterModelSummary, 0, len(models))
	for _, model := range models {
		if !modelMatchesWorkflowModality(model, modality) {
			continue
		}
		score := scoreModelMatch(model, query)
		if query != "" && score == 0 {
			continue
		}
		score += preferredModelBonus(model.ID, modality)
		summary := summarizeOpenRouterModel(model)
		summary.Score = score
		summaries = append(summaries, summary)
	}
	sort.SliceStable(summaries, func(i, j int) bool {
		if summaries[i].Score != summaries[j].Score {
			return summaries[i].Score > summaries[j].Score
		}
		return summaries[i].ID < summaries[j].ID
	})
	return summaries
}

func scoreModelMatch(model openRouterModel, query string) int {
	if query == "" {
		return 1
	}
	haystack := strings.ToLower(model.ID + " " + model.Name + " " + strings.Join(model.Architecture.InputModalities, " ") + " " + strings.Join(model.Architecture.OutputModalities, " "))
	id := strings.ToLower(model.ID)
	name := strings.ToLower(model.Name)
	if query == id {
		return 1000
	}
	if query == name {
		return 950
	}
	score := 0
	if strings.Contains(id, query) {
		score += 700
	}
	if strings.Contains(name, query) {
		score += 500
	}
	for _, token := range strings.FieldsFunc(query, func(r rune) bool {
		return r == ' ' || r == '-' || r == '_' || r == '/' || r == '.'
	}) {
		if token == "" {
			continue
		}
		if strings.Contains(haystack, token) {
			score += 100
		}
	}
	return score
}

func preferredModelBonus(modelID, modality string) int {
	preferred := preferredModelsForModality(modality)
	for index, preferredID := range preferred {
		if modelID == preferredID {
			return 250 - index
		}
	}
	return 0
}

func preferredModelsForModality(modality string) []string {
	switch normalizeModelModality(modality) {
	case "text":
		return preferredTextModels
	case "image":
		return preferredImageModels
	case "speech":
		return preferredSpeechModels
	case "transcription":
		return preferredTranscriptionModels
	case "embeddings":
		return preferredEmbeddingModels
	case "video":
		return preferredVideoModels
	default:
		return nil
	}
}

func modelConfidence(score int) string {
	switch {
	case score >= 500:
		return "high"
	case score >= 200:
		return "medium"
	default:
		return "low"
	}
}

func summarizeOpenRouterModel(model openRouterModel) openRouterModelSummary {
	return openRouterModelSummary{
		ID:                  model.ID,
		Name:                model.Name,
		InputModalities:     model.Architecture.InputModalities,
		OutputModalities:    model.Architecture.OutputModalities,
		ContextLength:       model.ContextLength,
		SupportedParameters: model.SupportedParameters,
		SupportedVoices:     model.SupportedVoices,
		Pricing:             model.Pricing,
	}
}
