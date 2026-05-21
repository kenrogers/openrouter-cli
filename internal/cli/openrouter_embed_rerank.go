package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kenrogers/openrouter-cli/internal/client"
	"github.com/kenrogers/openrouter-cli/internal/output"
	"github.com/spf13/cobra"
)

func newOpenRouterEmbedCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "embed [text]",
		Short: "Generate embeddings",
		Long:  "Generate embeddings for text or JSON input through OpenRouter embedding models.",
		Example: `  openrouter embed "The quick brown fox"
  openrouter embed --model google/gemini-embedding-2-preview --file notes.txt --output embedding.json
  openrouter embed --input-json '["query one","query two"]' --json`,
		RunE: runOpenRouterEmbedCommand,
	}
	cmd.Flags().StringP("input", "i", "", "Input text (can also be provided as positional args)")
	cmd.Flags().String("input-json", "", "Raw JSON input value for embedding arrays or token arrays")
	cmd.Flags().StringP("file", "f", "", "Read input text from a file")
	cmd.Flags().Bool("stdin", false, "Read input text from stdin")
	cmd.Flags().StringP("model", "m", "auto", "Embedding model ID, or auto to choose a current embedding model")
	cmd.Flags().Int64("dimensions", 0, "Number of embedding dimensions")
	cmd.Flags().String("encoding-format", "", "Embedding encoding format: float or base64")
	cmd.Flags().String("input-type", "", "Input type, such as search_query or search_document")
	cmd.Flags().String("provider", "", "Provider routing JSON object")
	cmd.Flags().String("output", "", "Optional file to save the full embedding response as JSON")
	cmd.Flags().Bool("force", false, "Overwrite output files if they already exist")
	return cmd
}

func enhanceOpenRouterRerankCommand(parent *cobra.Command) {
	rerankCmd := findChildCommand(parent, "rerank")
	if rerankCmd == nil {
		return
	}
	rerankCmd.Short = "Rerank documents"
	rerankCmd.Long = "Rerank documents through OpenRouter. The generated endpoint remains available as `openrouter rerank rerank`."
	rerankCmd.Example = `  openrouter rerank --query "capital of France" --documents "Paris is in France" --documents "Berlin is in Germany"
  openrouter rerank "capital of France" --documents-file docs.txt --top-n 3`
	rerankCmd.RunE = runOpenRouterRerankWorkflowCommand
	rerankCmd.Flags().String("query", "", "Search query to rank documents against (can also be positional)")
	rerankCmd.Flags().StringArray("documents", nil, "Document text; repeat for multiple documents")
	rerankCmd.Flags().String("documents-file", "", "Read documents from a JSON array file or newline-delimited text file")
	rerankCmd.Flags().StringP("model", "m", defaultRerankModel, "Rerank model ID")
	rerankCmd.Flags().Int64("top-n", 0, "Number of most relevant documents to return")
	rerankCmd.Flags().String("provider", "", "Provider routing JSON object")
}

func runOpenRouterEmbedCommand(cmd *cobra.Command, args []string) error {
	input, err := embeddingInputFromFlags(cmd, args)
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
		model, err = selectWorkflowModel(cmd, "embeddings", preferredEmbeddingModels, defaultEmbeddingModel)
		if err != nil {
			return err
		}
	}
	body := map[string]any{
		"model": model,
		"input": input,
	}
	if cmd.Flags().Changed("dimensions") {
		dimensions, _ := cmd.Flags().GetInt64("dimensions")
		body["dimensions"] = dimensions
	}
	if encodingFormat, _ := cmd.Flags().GetString("encoding-format"); strings.TrimSpace(encodingFormat) != "" {
		body["encoding_format"] = strings.TrimSpace(encodingFormat)
	}
	if inputType, _ := cmd.Flags().GetString("input-type"); strings.TrimSpace(inputType) != "" {
		body["input_type"] = strings.TrimSpace(inputType)
	}
	if provider, _ := cmd.Flags().GetString("provider"); strings.TrimSpace(provider) != "" {
		var providerConfig map[string]any
		if err := json.Unmarshal([]byte(provider), &providerConfig); err != nil {
			return fmt.Errorf("invalid --provider JSON: %w", err)
		}
		body["provider"] = providerConfig
	}

	req, _, err := openRouterJSONRequest(cmd.Context(), cmd, http.MethodPost, "/embeddings", body, apiKey)
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
	if outputPath, _ := cmd.Flags().GetString("output"); strings.TrimSpace(outputPath) != "" {
		force, _ := cmd.Flags().GetBool("force")
		saved, err := writeWorkflowFile(outputPathForMedia(outputPath, "openrouter-embedding", ".json"), raw, force)
		if err != nil {
			return err
		}
		response["embedding_saved"] = saved
	}
	response["model"] = model
	response["key_source"] = keySource
	if dimensions := firstEmbeddingDimensions(response); dimensions > 0 {
		response["dimensions"] = dimensions
	}
	return output.Result(cmd, response)
}

func runOpenRouterRerankWorkflowCommand(cmd *cobra.Command, args []string) error {
	query, _ := cmd.Flags().GetString("query")
	query = strings.TrimSpace(query)
	if query == "" {
		query = strings.TrimSpace(strings.Join(args, " "))
	}
	if query == "" {
		return fmt.Errorf("missing query: run `openrouter rerank --query \"your query\" --documents \"doc\"`")
	}
	documents, err := rerankDocumentsFromFlags(cmd)
	if err != nil {
		return err
	}
	if len(documents) == 0 {
		return fmt.Errorf("missing documents: pass --documents repeatedly or --documents-file")
	}
	apiKey, keySource, err := openRouterAPIKey(cmd, true)
	if err != nil {
		return err
	}
	model, _ := cmd.Flags().GetString("model")
	model = strings.TrimSpace(model)
	if model == "" || strings.EqualFold(model, "auto") {
		model = defaultRerankModel
	}
	body := map[string]any{
		"model":     model,
		"query":     query,
		"documents": documents,
	}
	if cmd.Flags().Changed("top-n") {
		topN, _ := cmd.Flags().GetInt64("top-n")
		body["top_n"] = topN
	}
	if provider, _ := cmd.Flags().GetString("provider"); strings.TrimSpace(provider) != "" {
		var providerConfig map[string]any
		if err := json.Unmarshal([]byte(provider), &providerConfig); err != nil {
			return fmt.Errorf("invalid --provider JSON: %w", err)
		}
		body["provider"] = providerConfig
	}

	req, _, err := openRouterJSONRequest(cmd.Context(), cmd, http.MethodPost, "/rerank", body, apiKey)
	if err != nil {
		return err
	}
	var response map[string]any
	if _, err := openRouterDoJSON(cmd, req, &response, 2*time.Minute); err != nil {
		return err
	}
	if client.IsDryRun(cmd) {
		return nil
	}
	response["model"] = model
	response["query"] = query
	response["document_count"] = len(documents)
	response["key_source"] = keySource
	return output.Result(cmd, response)
}

func embeddingInputFromFlags(cmd *cobra.Command, args []string) (any, error) {
	if rawJSON, _ := cmd.Flags().GetString("input-json"); strings.TrimSpace(rawJSON) != "" {
		var value any
		if err := json.Unmarshal([]byte(rawJSON), &value); err != nil {
			return nil, fmt.Errorf("invalid --input-json: %w", err)
		}
		return value, nil
	}
	values := []string{}
	if input, _ := cmd.Flags().GetString("input"); strings.TrimSpace(input) != "" {
		values = append(values, strings.TrimSpace(input))
	}
	if argInput := strings.TrimSpace(strings.Join(args, " ")); argInput != "" {
		values = append(values, argInput)
	}
	if filePath, _ := cmd.Flags().GetString("file"); strings.TrimSpace(filePath) != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read input file %s: %w", filePath, err)
		}
		values = append(values, strings.TrimSpace(string(data)))
	}
	if useStdin, _ := cmd.Flags().GetBool("stdin"); useStdin {
		data, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
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
		return nil, fmt.Errorf("provide input through only one input source")
	}
	if len(nonEmpty) == 0 {
		return nil, fmt.Errorf("missing input: run `openrouter embed \"text\"`, use --input, --file, --stdin, or --input-json")
	}
	return nonEmpty[0], nil
}

func rerankDocumentsFromFlags(cmd *cobra.Command) ([]string, error) {
	documents, _ := cmd.Flags().GetStringArray("documents")
	if filePath, _ := cmd.Flags().GetString("documents-file"); strings.TrimSpace(filePath) != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read documents file %s: %w", filePath, err)
		}
		var parsed []string
		if err := json.Unmarshal(data, &parsed); err == nil {
			documents = append(documents, parsed...)
		} else {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if line != "" {
					documents = append(documents, line)
				}
			}
		}
	}
	cleaned := make([]string, 0, len(documents))
	for _, document := range documents {
		if strings.TrimSpace(document) != "" {
			cleaned = append(cleaned, strings.TrimSpace(document))
		}
	}
	return cleaned, nil
}

func firstEmbeddingDimensions(response map[string]any) int {
	data, _ := response["data"].([]any)
	for _, item := range data {
		itemMap, _ := item.(map[string]any)
		embedding, _ := itemMap["embedding"].([]any)
		if len(embedding) > 0 {
			return len(embedding)
		}
		if encoded, _ := itemMap["embedding"].(string); encoded != "" {
			return len(encoded)
		}
	}
	return 0
}
