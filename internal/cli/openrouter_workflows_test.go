package cli

import (
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

func TestAskCommandUsesAutoTextModel(t *testing.T) {
	config.Reset()
	t.Setenv("OPENROUTER_API_KEY", "sk-or-v1-test")

	var chatBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/models":
			if r.URL.Query().Get("output_modalities") != "text" {
				t.Fatalf("output_modalities = %q, want text", r.URL.Query().Get("output_modalities"))
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data":[{"id":"google/gemini-3.1-flash-lite","name":"Gemini Flash Lite","architecture":{"input_modalities":["text"],"output_modalities":["text"]}}]}`))
		case "/chat/completions":
			if err := json.NewDecoder(r.Body).Decode(&chatBody); err != nil {
				t.Fatalf("decode chat body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"choices":[{"message":{"content":"Hello from OpenRouter."}}]}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, err := executeRootCommand(t, "--server-url", server.URL, "--json", "ask", "say hello")
	if err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode result: %v\nstdout: %s", err, stdout)
	}
	if result["model"] != defaultTextModel || result["text"] != "Hello from OpenRouter." {
		t.Fatalf("unexpected result: %#v", result)
	}
	if chatBody["model"] != defaultTextModel {
		t.Fatalf("chat model = %#v", chatBody["model"])
	}
}

func TestModelsResolveFindsFuzzyMatch(t *testing.T) {
	config.Reset()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"id":"anthropic/claude-sonnet-4.6","name":"Claude Sonnet 4.6","architecture":{"input_modalities":["text"],"output_modalities":["text"]}},{"id":"google/gemini-3.1-flash-lite","name":"Gemini Flash Lite","architecture":{"input_modalities":["text"],"output_modalities":["text"]}}]}`))
	}))
	defer server.Close()

	stdout, stderr, err := executeRootCommand(t, "--server-url", server.URL, "--json", "models", "resolve", "sonnet", "--modality", "text")
	if err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode result: %v\nstdout: %s", err, stdout)
	}
	if result["model"] != "anthropic/claude-sonnet-4.6" {
		t.Fatalf("resolved model = %#v", result["model"])
	}
}

func TestAudioSpeakSavesBinary(t *testing.T) {
	config.Reset()
	t.Setenv("OPENROUTER_API_KEY", "sk-or-v1-test")

	var speechBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/models":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data":[{"id":"openai/gpt-4o-mini-tts-2025-12-15","name":"OpenAI TTS","architecture":{"input_modalities":["text"],"output_modalities":["audio"]}}]}`))
		case "/audio/speech":
			if err := json.NewDecoder(r.Body).Decode(&speechBody); err != nil {
				t.Fatalf("decode speech body: %v", err)
			}
			w.Header().Set("Content-Type", "audio/mpeg")
			w.Write([]byte("mp3-data"))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "voice.mp3")
	stdout, stderr, err := executeRootCommand(t, "--server-url", server.URL, "--json", "audio", "speak", "--output", outputPath, "hello")
	if err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
	}
	if data, err := os.ReadFile(outputPath); err != nil || string(data) != "mp3-data" {
		t.Fatalf("saved audio = %q, err=%v", string(data), err)
	}
	if speechBody["model"] != defaultSpeechModel || speechBody["input"] != "hello" {
		t.Fatalf("speech body = %#v", speechBody)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode result: %v\nstdout: %s", err, stdout)
	}
	if result["audio_saved"] != outputPath {
		t.Fatalf("audio_saved = %#v", result["audio_saved"])
	}
}

func TestAudioTranscribeEncodesLocalFile(t *testing.T) {
	config.Reset()
	t.Setenv("OPENROUTER_API_KEY", "sk-or-v1-test")

	audioPath := filepath.Join(t.TempDir(), "meeting.wav")
	if err := os.WriteFile(audioPath, []byte("wav-data"), 0o600); err != nil {
		t.Fatal(err)
	}
	var transcriptionBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/models":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data":[{"id":"google/chirp-3","name":"Chirp 3","architecture":{"input_modalities":["audio"],"output_modalities":["text"]}}]}`))
		case "/audio/transcriptions":
			if err := json.NewDecoder(r.Body).Decode(&transcriptionBody); err != nil {
				t.Fatalf("decode transcription body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"text":"hello transcript"}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, err := executeRootCommand(t, "--server-url", server.URL, "--json", "audio", "transcribe", audioPath)
	if err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
	}
	inputAudio := transcriptionBody["input_audio"].(map[string]any)
	decoded, err := base64.StdEncoding.DecodeString(inputAudio["data"].(string))
	if err != nil {
		t.Fatal(err)
	}
	if string(decoded) != "wav-data" || inputAudio["format"] != "wav" {
		t.Fatalf("input_audio = %#v", inputAudio)
	}
	if !strings.Contains(stdout, "hello transcript") {
		t.Fatalf("stdout = %s", stdout)
	}
}

func TestEmbedCommandWritesResponse(t *testing.T) {
	config.Reset()
	t.Setenv("OPENROUTER_API_KEY", "sk-or-v1-test")

	var embedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/models":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data":[{"id":"google/gemini-embedding-2-preview","name":"Gemini Embedding","architecture":{"input_modalities":["text"],"output_modalities":["embeddings"]}}]}`))
		case "/embeddings":
			if err := json.NewDecoder(r.Body).Decode(&embedBody); err != nil {
				t.Fatalf("decode embed body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data":[{"embedding":[0.1,0.2,0.3],"index":0}],"usage":{"total_tokens":4}}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "embedding.json")
	stdout, stderr, err := executeRootCommand(t, "--server-url", server.URL, "--json", "embed", "--output", outputPath, "hello")
	if err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
	}
	if embedBody["model"] != defaultEmbeddingModel || embedBody["input"] != "hello" {
		t.Fatalf("embed body = %#v", embedBody)
	}
	if data, err := os.ReadFile(outputPath); err != nil || !strings.Contains(string(data), "embedding") {
		t.Fatalf("saved embedding = %q, err=%v", string(data), err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode result: %v\nstdout: %s", err, stdout)
	}
	if result["dimensions"] != float64(3) {
		t.Fatalf("dimensions = %#v", result["dimensions"])
	}
}

func TestRerankRootCommandPostsDocuments(t *testing.T) {
	config.Reset()
	t.Setenv("OPENROUTER_API_KEY", "sk-or-v1-test")

	var rerankBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rerank" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&rerankBody); err != nil {
			t.Fatalf("decode rerank body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[{"index":0,"relevance_score":0.99}]}`))
	}))
	defer server.Close()

	stdout, stderr, err := executeRootCommand(t, "--server-url", server.URL, "--json", "rerank", "--query", "capital", "--documents", "Paris", "--documents", "Berlin")
	if err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
	}
	if rerankBody["model"] != defaultRerankModel || rerankBody["query"] != "capital" {
		t.Fatalf("rerank body = %#v", rerankBody)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode result: %v\nstdout: %s", err, stdout)
	}
	if result["document_count"] != float64(2) {
		t.Fatalf("document_count = %#v", result["document_count"])
	}
}

func TestVideoCommandWaitsAndDownloadsContent(t *testing.T) {
	config.Reset()
	t.Setenv("OPENROUTER_API_KEY", "sk-or-v1-test")

	var videoBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/videos/models":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data":[{"id":"google/veo-3.1-lite","name":"Veo Lite","supported_durations":[4],"supported_resolutions":["720p"],"supported_aspect_ratios":["16:9"]}]}`))
		case "/videos":
			if err := json.NewDecoder(r.Body).Decode(&videoBody); err != nil {
				t.Fatalf("decode video body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id":"job-1","status":"pending"}`))
		case "/videos/job-1":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id":"job-1","status":"completed"}`))
		case "/videos/job-1/content":
			w.Header().Set("Content-Type", "video/mp4")
			w.Write([]byte("mp4-data"))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "clip.mp4")
	stdout, stderr, err := executeRootCommand(t,
		"--server-url", server.URL,
		"--json",
		"video",
		"--poll-interval", "1ms",
		"--wait-timeout", "1s",
		"--output", outputPath,
		"a cloud timelapse",
	)
	if err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
	}
	if videoBody["model"] != defaultVideoModel || videoBody["prompt"] != "a cloud timelapse" {
		t.Fatalf("video body = %#v", videoBody)
	}
	if data, err := os.ReadFile(outputPath); err != nil || string(data) != "mp4-data" {
		t.Fatalf("saved video = %q, err=%v", string(data), err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode result: %v\nstdout: %s", err, stdout)
	}
	if result["video_saved"] != outputPath {
		t.Fatalf("video_saved = %#v", result["video_saved"])
	}
}
