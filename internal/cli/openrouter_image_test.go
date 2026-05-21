package cli

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kenrogers/openrouter-cli/internal/config"
)

func TestImageCommandGeneratesAndSavesImage(t *testing.T) {
	config.Reset()
	t.Setenv("OPENROUTER_API_KEY", "sk-or-v1-test")

	var chatBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/models":
			if r.URL.Query().Get("output_modalities") != "image" {
				t.Fatalf("output_modalities = %q, want image", r.URL.Query().Get("output_modalities"))
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data":[{"id":"google/gemini-3.1-flash-image-preview","name":"Gemini Image","architecture":{"input_modalities":["text","image"],"output_modalities":["image","text"]}}]}`))
		case "/chat/completions":
			if got := r.Header.Get("Authorization"); got != "Bearer sk-or-v1-test" {
				t.Fatalf("Authorization = %q", got)
			}
			if err := json.NewDecoder(r.Body).Decode(&chatBody); err != nil {
				t.Fatalf("decode chat body: %v", err)
			}
			dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("png-data"))
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{
					{
						"message": map[string]any{
							"content": "done",
							"images": []map[string]any{
								{"type": "image_url", "image_url": map[string]any{"url": dataURL}},
							},
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "robot.png")
	stdout, stderr, err := executeRootCommand(t,
		"--server-url", server.URL,
		"--json",
		"image",
		"--aspect-ratio", "16:9",
		"--output", outputPath,
		"a tiny red robot",
	)
	if err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read saved image: %v", err)
	}
	if string(data) != "png-data" {
		t.Fatalf("saved image = %q", string(data))
	}

	var result imageGenerateResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode result: %v\nstdout: %s", err, stdout)
	}
	if result.Model != defaultImageModel || result.Prompt != "a tiny red robot" || result.Count != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(result.ImagesSaved) != 1 || result.ImagesSaved[0] != outputPath {
		t.Fatalf("images_saved = %#v, want %s", result.ImagesSaved, outputPath)
	}
	if chatBody["model"] != defaultImageModel {
		t.Fatalf("model = %v", chatBody["model"])
	}
	if config, ok := chatBody["image_config"].(map[string]any); !ok || config["aspect_ratio"] != "16:9" {
		t.Fatalf("image_config = %#v", chatBody["image_config"])
	}
}

func TestImageCommandEncodesInputImageForEditing(t *testing.T) {
	config.Reset()
	t.Setenv("OPENROUTER_API_KEY", "sk-or-v1-test")

	inputPath := filepath.Join(t.TempDir(), "input.png")
	if err := os.WriteFile(inputPath, []byte("input-image"), 0o600); err != nil {
		t.Fatal(err)
	}

	var chatBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/chat/completions":
			if err := json.NewDecoder(r.Body).Decode(&chatBody); err != nil {
				t.Fatalf("decode chat body: %v", err)
			}
			dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("edited"))
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{{"message": map[string]any{"images": []map[string]any{{"image_url": map[string]any{"url": dataURL}}}}}},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "edited.png")
	_, stderr, err := executeRootCommand(t,
		"--server-url", server.URL,
		"--json",
		"image",
		"--model", "test/image-model",
		"--input-image", inputPath,
		"--output", outputPath,
		"make it watercolor",
	)
	if err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
	}

	messages := chatBody["messages"].([]any)
	content := messages[0].(map[string]any)["content"].([]any)
	if content[0].(map[string]any)["text"] != "make it watercolor" {
		t.Fatalf("text content = %#v", content[0])
	}
	imageURL := content[1].(map[string]any)["image_url"].(map[string]any)["url"].(string)
	if !strings.HasPrefix(imageURL, "data:image/png;base64,") {
		t.Fatalf("input image URL = %q", imageURL)
	}
	encoded := strings.TrimPrefix(imageURL, "data:image/png;base64,")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(decoded) != "input-image" {
		t.Fatalf("decoded input = %q", string(decoded))
	}
}

func TestImageModelsCommandFiltersImageModels(t *testing.T) {
	config.Reset()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"id":"black-forest-labs/flux.2-pro","name":"Flux Pro","architecture":{"input_modalities":["text"],"output_modalities":["image"]}},{"id":"google/gemini-3.1-flash-image-preview","name":"Gemini Image","architecture":{"input_modalities":["text"],"output_modalities":["image","text"]}}]}`))
	}))
	defer server.Close()

	stdout, stderr, err := executeRootCommand(t, "--server-url", server.URL, "--json", "image", "models", "flux")
	if err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
	}
	var result struct {
		Count  int                 `json:"count"`
		Total  int                 `json:"total"`
		Models []imageModelSummary `json:"models"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode result: %v\nstdout: %s", err, stdout)
	}
	if result.Count != 1 || result.Total != 1 || result.Models[0].ID != "black-forest-labs/flux.2-pro" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func executeRootCommand(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd, err := NewRootCommand()
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return stdout.String(), stderr.String(), err
}
