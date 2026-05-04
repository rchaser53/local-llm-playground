const axios = require('axios');
const cheerio = require('cheerio');

/**
 * 指定されたURLの記事を取得し、本文を抽出し、ローカルLLMに問い合わせて要約をコンソールに出力する関数。
 * コマンドライン引数からURLを受け取ることを想定しています。
 * @param {string} url - 要約対象のウェブサイトのURL
 */
async function summarizeArticle(url) {
    if (!url) {
        console.error("エラー: 要約対象のURLが指定されていません。");
        console.error("使用方法: node summarizeArticle.js <URL>");
        return;
    }
    console.log(`\n--- 記事の読み込みと要約を開始します: ${url} ---`);

    try {
        // 1. 記事のHTMLコンテンツを取得
        const response = await axios.get(url);
        const html = response.data;

        // 2. Cheerioを使用してHTMLをパースし、記事本文を抽出（簡易版）
        const $ = cheerio.load(html);
        
        // 記事本文を抽出するロジックはサイト構造に依存するため、
        // ここでは一般的な記事コンテナ要素やテキストコンテンツを集めるなど、
        // より高度なセレクタが必要です。デモとして、読み込んだテキストの先頭部分を取得します。
        let extractedText = $('body').text();

        // 長すぎる場合、先頭から一定量だけ切り出す（文字数制限のため）
        if (extractedText.length > 1500) {
            extractedText = extractedText.substring(0, 1500) + '... (本文の続きは省略されています)';
        }
        
        console.log(`[成功] 本文の抽出が完了しました。（文字数: 約${extractedText.length}文字）`);

        // 3. テキストの要約（ローカルLLM APIを呼び出す）
        let summary = "";
        
        try {
            // ★ローカルLLMを呼び出す関数を呼び出す
            summary = await callLlmApi(extractedText, "この記事を、一般の読者向けに、論理的かつ簡潔に3行で要約してください。専門用語は避け、最も伝えたい核となるメッセージに焦点を当ててください。");
        } catch (error) {
            console.warn("\n⚠️ LLMへの問い合わせに失敗しました。ローカルLLMが起動しているか、APIエンドポイントを確認してください。", error.message);
            // 失敗した場合のフォールバックメッセージ
            summary = `【ローカルLLMによる要約が失敗したため、以下のプレースホルダ情報を使用します。】\n` +
                      `1. 記事本文の抽出は成功しましたが、LLMへの接続に失敗しました。\n` +
                      `2. 抽出されたテキストの構造から、重要なトピックが存在すると想定されます。\n` +
                      `3. お手数ですが、ローカルLLMの起動とAPIエンドポイントの設定をご確認ください。`;
        }


        // 4. 結果の出力
        console.log("\n============================");
        console.log("✅ 記事の要約結果（3行）");
        console.log("===============");
        console.log(summary);
        console.log("====================================\n");


    } catch (error) {
        console.error("\n❌ エラーが発生しました。URLの確認、またはネットワーク接続を確認してください。", error.message);
    }
}

/**
 * ローカルLLM APIを呼び出してテキストを要約する（シミュレーション/実装部）
 * @param {string} text - 要約対象の本文
 * @param {string} prompt - LLMへの具体的な指示プロンプト
 * @returns {Promise<string>} - LLMから返された要約テキスト
 */
async function callLlmApi(text, prompt) {
    console.log("\n[LLM処理] ローカルLLM APIに問い合わせを行います...");
    
    // ★★★ ここを実際のローカルLLMのAPIエンドポイントに変更してください ★★★
    const LLM_API_URL = "http://localhost:11434/api/generate"; // 例: Ollamaのエンドポイント
    
    const payload = {
        model: config.modelName, // 使用モデル名
        prompt: `${prompt}\n\n--- 記事本文 ---\n${text}`,
        stream: false // ストリーミングを無効にする
    };

    try {
        const response = await axios.post(LLM_API_URL, payload, {
            headers: {
                'Content-Type': 'application/json'
            }
        });

        // LLMからの応答形式に合わせてパースする (Ollamaの場合、response.data.responseに結果が入ることが多い)
        const rawSummary = response.data.response; 
        console.log("[LLM処理] API呼び出し成功。結果を受け取りました。");
        return rawSummary.trim();

    } catch (error) {
        console.error("[LLM処理] API呼び出しエラー:", error.message);
        // エラーを投げることで、try-catchブロックでキャッチさせ、フォールバック処理を走らせる
        throw new Error("Could not connect to the local LLM API at " + LLM_API_URL);
    }
}





// --- メイン実行ロジック ---
const targetUrl = process.argv[2]; // コマンドライン引数からURLを取得

// スクリプト実行

summarizeArticle(targetUrl);
