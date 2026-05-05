package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"gopkg.in/yaml.v3"
)

// Config はアプリケーションの設定を保持します。
type Config struct {
	ModelName string `yaml:"modelName"`
	LLMApiURL string `yaml:"llmApiUrl"`
}

// LLMRequest はローカルLLMへのリクエストペイロード構造体です。
type LLMRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// LLMResponse はローカルLLMからの応答ペイロード構造体（Ollamaを想定）です。
type LLMResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}

// loadConfig YAMLファイルから設定を読み込む関数。
func loadConfig(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("configファイルの読み込みに失敗しました: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("configYAMLのパースに失敗しました: %w", err)
	}
	return &cfg, nil
}

// extractArticleText はURLから記事本文を抽出する関数（簡易版）。
// 本来はサイト構造に依存しますが、ここでは基本的なテキスト抽出ロジックを再現します。
func extractArticleText(url string) (string, error) {
	// HTTPクライアントの設定
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// ウェブページを取得
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("URLへの接続に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	// ステータスコードチェック
	if resp.StatusCode != http.StatusOK {
		// 403 Forbidden やその他の非成功ステータスを明確に返す
		return "", fmt.Errorf("HTTPエラーが発生しました: ステータスコード %d. アクセス権限がないか、URLが存在しない可能性があります。", resp.StatusCode)
	}

	// ボディを全て読み込む
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ウェブページの読み取りに失敗しました: %w", err)
	}

	// HTMLをパースし、本文を抽出する（Cheerio/jsのロジックをGoで再現）
	doc, err := htmlquery.Parse(bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("HTMLのパースに失敗しました: %w", err)
	}

	// bodyタグ内の全てのテキストを取得する
	nodes := htmlquery.Find(doc, "//body")
	if len(nodes) == 0 {
		return "", fmt.Errorf("HTML構造から<body要素>が見つかりませんでした。")
	}

	// bodyの内容全体をテキストとして抽出
	textNodes := htmlquery.Find(nodes[0], "*")
	var builder strings.Builder
	for _, node := range textNodes {
		// 要素タグの判定と除外: スクリプトやスタイルタグは処理から除外する
		if node.FirstChild != nil && (node.Data == "script" || node.Data == "style") {
			continue
		}

		// テキストコンテンツとして結合
		builder.WriteString(htmlquery.InnerText(node))
		builder.WriteString(" ")
	}
	extractedText := strings.TrimSpace(builder.String())

	// 文字数制限を設ける
	const maxChars = 1500
	if len(extractedText) > maxChars {
		extractedText = extractedText[:maxChars] + " ... (本文の続きは省略されています)"
	}

	return extractedText, nil
}

// callLLMApi はローカルLLM APIに問い合わせてテキストを要約する。
func callLLMApi(text string, prompt string, cfg *Config) (string, error) {
	fmt.Println("\n[LLM処理] ローカルLLM APIに問い合わせを行います...")

	// リクエストペイロードの作成
	payload := LLMRequest{
		Model:  cfg.ModelName,
		Prompt: fmt.Sprintf("%s\n\n--- 記事本文 ---\n%s", prompt, text),
		Stream: false,
	}
	payloadBytes, _ := json.Marshal(payload)

	// HTTPリクエストの構築
	req, err := http.NewRequest("POST", cfg.LLMApiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("リクエスト構築失敗: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// API呼び出し
	client := &http.Client{
		// タイムアウトを60秒に延長し、ネットワーク遅延による失敗に対応します。
		Timeout: 600 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		// タイムアウトや接続エラーの場合の具体的なメッセージを返す
		if strings.Contains(err.Error(), "timeout") {
			return "", fmt.Errorf("LLM APIへの接続がタイムアウトしました。LLMサーバーが起動しているか、ネットワーク接続を確認してください。タイムアウトをさらに延ばす必要があるか、LLMサーバーの状態を確認してください。詳細エラー: %w", err)
		}
		return "", fmt.Errorf("API呼び出し失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("APIからの応答が予期せぬステータスコード %d で終了しました。ボディ: %s", resp.StatusCode, string(bodyBytes))
	}

	// 応答の読み込みとパース
	var llmResp LLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return "", fmt.Errorf("API応答JSONのパースに失敗しました: %w", err)
	}

	if llmResp.Response == "" {
		return "", fmt.Errorf("LLMからの応答本文が見つかりませんでした。")
	}

	fmt.Println("[LLM処理] API呼び出し成功。結果を受け取りました。")
	return llmResp.Response, nil
}

// summarizeArticle はメインの記事要約ロジックです。
func summarizeArticle(url string, cfg *Config) {
	if url == "" {
		fmt.Println("エラー: 要約対象のURLが指定されていません。")
		fmt.Println("使用方法: ./local-llm-playground <URL>")
		return
	}

	fmt.Printf("\n--- 記事の読み込みと要約を開始します: %s ---\n", url)

	// 1. 記事本文の抽出
	extractedText, err := extractArticleText(url)
	if err != nil {
		log.Fatalf("記事本文抽出失敗: %v", err)
	}

	fmt.Printf("[成功] 本文の抽出が完了しました。（文字数: 約%d文字）\n", len(extractedText))

	// 2. テキストの要約（ローカルLLM APIを呼び出す）
	const prompt = "この記事を、一般の読者向けに、論理的かつ簡潔に3行で要約してください。専門用語は避け、最も伝えたい核となるメッセージに焦点を当ててください。"
	var summary string

	summary, err = callLLMApi(extractedText, prompt, cfg)
	if err != nil {
		fmt.Printf("\n⚠️ LLMへの問い合わせに失敗しました。ローカルLLMが起動しているか、APIエンドポイントを確認してください。エラー: %v\n", err)
		// 失敗した場合のフォールバックメッセージ
		summary = `【ローカルLLMによる要約が失敗したため、以下のプレースホルダ情報を使用します。】
1. 記事本文の抽出は成功しましたが、LLMへの接続に失敗しました。
2. 抽出されたテキストの構造から、重要なトピックが存在すると想定されます。
3. お手数ですが、ローカルLLMの起動とAPIエンドポイントの設定をご確認ください。`
	}

	// 3. 結果の出力
	fmt.Println("\n=============== ===============")
	fmt.Println("✅ 記事の要約結果（3行）")
	fmt.Println("===============")
	fmt.Println(summary)
	fmt.Println("==============================================")
}

func main() {
	// コマンドライン引数からURLを取得。os.Args[1]がURLとなります。
	url := ""
	if len(os.Args) > 1 {
		url = os.Args[1]
	}

	// --- URLバリデーションの追加 ---
	// 引数がURL形式でない場合（例: --help, -h など）、適切な使い方を表示して終了する
	if url != "" && !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		fmt.Println("エラー: 無効なURL形式の引数が指定されました。")
		fmt.Printf("使用方法: %s <URL>\n", os.Args[0])
		fmt.Println("例: ./local-llm-playground https://www.example.com")
		os.Exit(1)
	}
	// -------------------------------

	// 設定ファイルのロード
	cfg, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("設定ファイルのロードに失敗しました: %v", err)
	}

	// 実行
	summarizeArticle(url, cfg)
}
