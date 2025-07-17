// WebAssembly関連の変数
let wasmInstance = null;
let selectedImage = null;
let pyodide = null;

// DOM要素の取得
const fileInput = document.getElementById('file-input');
const uploadBtn = document.getElementById('upload-btn');
const executeBtn = document.getElementById('execute-btn');
const previewSection = document.querySelector('.preview-section');
const resultSection = document.querySelector('.result-section');
const errorSection = document.querySelector('.error-section');
const previewImage = document.getElementById('preview-image');
const outputContent = document.getElementById('output-content');
const codeContent = document.getElementById('code-content');
const astContent = document.getElementById('ast-content');
const errorContent = document.getElementById('error-content');
const loading = document.getElementById('loading');

// サンプル画像のマッピング
const sampleImages = {
    'hello-world': 'static/samples/hello-world.png',
    'calculator': 'static/samples/calculator.png',
    'fibonacci': 'static/samples/fibonacci.png',
    'loop': 'static/samples/loop.png'
};

// WebAssemblyの初期化
async function initWasm() {
    try {
        // Goオブジェクトが利用可能か確認
        if (typeof Go === 'undefined') {
            console.error("Go is not defined. Checking window.Go:", window.Go);
            throw new Error("Go is not defined. Please ensure wasm_exec.js is loaded.");
        }
        
        console.log("Creating Go instance...");
        const go = new Go();
        
        console.log("Fetching WASM file from:", "static/wasm/grimoire.wasm");
        const result = await WebAssembly.instantiateStreaming(
            fetch("static/wasm/grimoire.wasm"), 
            go.importObject
        );
        go.run(result.instance);
        wasmInstance = result.instance;
        window.wasmInstance = result.instance; // E2Eテスト用にグローバルに公開
        window.processGrimoireImage = window.processGrimoireImage || processGrimoireImage; // E2Eテスト用
        console.log("WebAssembly initialized successfully");
        console.log("processGrimoireImage function available:", typeof window.processGrimoireImage);
    } catch (error) {
        console.error("Failed to initialize WebAssembly:", error);
        showError("WebAssemblyの初期化に失敗しました: " + error.message);
    }
}

// Pyodideの初期化
async function initPyodide() {
    try {
        pyodide = await loadPyodide();
        window.pyodide = pyodide; // E2Eテスト用にグローバルに公開
        console.log("Pyodide initialized successfully");
    } catch (error) {
        console.error("Failed to initialize Pyodide:", error);
        console.log("Python execution will not be available");
    }
}

// 画像をBase64に変換
async function imageToBase64(imageUrl) {
    try {
        const response = await fetch(imageUrl);
        const blob = await response.blob();
        
        // Blobを直接ArrayBufferに変換
        const arrayBuffer = await blob.arrayBuffer();
        const uint8Array = new Uint8Array(arrayBuffer);
        
        // Base64に手動でエンコード
        let binary = '';
        for (let i = 0; i < uint8Array.byteLength; i++) {
            binary += String.fromCharCode(uint8Array[i]);
        }
        return btoa(binary);
    } catch (error) {
        console.error('Failed to convert image to base64:', error);
        throw error;
    }
}

// ファイルをBase64に変換
async function fileToBase64(file) {
    try {
        const arrayBuffer = await file.arrayBuffer();
        const uint8Array = new Uint8Array(arrayBuffer);
        
        // Base64に手動でエンコード
        let binary = '';
        for (let i = 0; i < uint8Array.byteLength; i++) {
            binary += String.fromCharCode(uint8Array[i]);
        }
        return btoa(binary);
    } catch (error) {
        console.error('Failed to convert file to base64:', error);
        throw error;
    }
}

// エラー表示
function showError(message) {
    errorSection.style.display = 'block';
    resultSection.style.display = 'none';
    errorContent.textContent = message;
    loading.style.display = 'none';
}

// ローディング表示
function showLoading() {
    loading.style.display = 'flex';
}

function hideLoading() {
    loading.style.display = 'none';
}

// 画像プレビューの表示
function showPreview(imageUrl) {
    previewImage.src = imageUrl;
    previewSection.style.display = 'block';
    resultSection.style.display = 'none';
    errorSection.style.display = 'none';
}

// 実行結果の表示
async function showResult(result) {
    if (result.success) {
        resultSection.style.display = 'block';
        errorSection.style.display = 'none';
        
        // Pythonコードを表示
        codeContent.textContent = result.code || "// コードが生成されませんでした";
        
        // デバッグ情報とASTを表示
        let debugDisplay = "";
        if (result.debug) {
            // result.debugが文字列の場合はJSONパース
            let debugInfo = result.debug;
            if (typeof debugInfo === 'string') {
                try {
                    debugInfo = JSON.parse(debugInfo);
                } catch (e) {
                    console.error("Failed to parse debug info:", e);
                }
            }
            
            debugDisplay += "=== デバッグ情報 ===\n";
            debugDisplay += `検出されたシンボル数: ${debugInfo.symbolCount}\n`;
            if (debugInfo.symbols && debugInfo.symbols.length > 0) {
                debugDisplay += "シンボル一覧:\n";
                debugInfo.symbols.forEach((sym, i) => {
                    debugDisplay += `  ${i}: ${sym.type} at (${sym.position.x}, ${sym.position.y})`;
                    if (sym.pattern) debugDisplay += ` pattern: ${sym.pattern}`;
                    debugDisplay += "\n";
                });
            }
            debugDisplay += "\n=== AST ===\n";
        }
        debugDisplay += JSON.stringify(result.ast || {}, null, 2);
        astContent.textContent = debugDisplay;
        
        // Pyodideが利用可能な場合はPythonコードを実行
        if (pyodide && result.code) {
            try {
                console.log("Generated Python code:", result.code);
                
                // 出力をキャプチャするための設定
                pyodide.runPython(`
import sys
from io import StringIO
output_buffer = StringIO()
sys.stdout = output_buffer
                `);
                
                // Pythonコードを実行
                pyodide.runPython(result.code);
                
                // 出力を取得
                const output = pyodide.runPython(`
output_buffer.getvalue()
                `);
                
                console.log("Python output:", output);
                outputContent.textContent = output || "（出力なし）";
            } catch (error) {
                console.error("Python execution error:", error);
                outputContent.textContent = `実行エラー: ${error.message}`;
            }
        } else {
            outputContent.textContent = result.output || "Python execution in browser requires Pyodide integration";
            if (result.warning) {
                outputContent.textContent += "\n\n⚠️ " + result.warning;
            }
        }
    } else {
        showError(result.error || "不明なエラーが発生しました");
    }
}

// 画像の処理と実行
async function processImage() {
    if (!selectedImage) {
        showError("画像が選択されていません");
        return;
    }
    
    showLoading();
    
    try {
        let base64Image;
        
        if (typeof selectedImage === 'string') {
            // URLの場合
            base64Image = await imageToBase64(selectedImage);
        } else {
            // Fileオブジェクトの場合
            base64Image = await fileToBase64(selectedImage);
        }
        
        // WebAssemblyの関数を呼び出し
        const result = processGrimoireImage(base64Image);
        await showResult(result);
        
    } catch (error) {
        showError("画像の処理中にエラーが発生しました: " + error.message);
    } finally {
        hideLoading();
    }
}

// イベントリスナーの設定
uploadBtn.addEventListener('click', () => {
    fileInput.click();
});

fileInput.addEventListener('change', (event) => {
    const file = event.target.files[0];
    if (file) {
        selectedImage = file;
        const url = URL.createObjectURL(file);
        showPreview(url);
    }
});

executeBtn.addEventListener('click', processImage);

// サンプル画像のクリックイベント
document.querySelectorAll('.sample-item').forEach(item => {
    item.addEventListener('click', () => {
        const sampleName = item.dataset.sample;
        const imageUrl = sampleImages[sampleName];
        if (imageUrl) {
            selectedImage = imageUrl;
            showPreview(imageUrl);
        }
    });
});

// タブ切り替え
document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', () => {
        const tabName = btn.dataset.tab;
        
        // タブボタンのアクティブ状態を更新
        document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        
        // タブコンテンツの表示を更新
        document.querySelectorAll('.tab-pane').forEach(pane => pane.classList.remove('active'));
        document.getElementById(tabName + '-tab').classList.add('active');
    });
});

// wasm_exec.jsの読み込みを待つ
function waitForGo() {
    return new Promise((resolve) => {
        if (typeof Go !== 'undefined') {
            resolve();
        } else {
            // Goオブジェクトが定義されるまで待つ
            const checkInterval = setInterval(() => {
                if (typeof Go !== 'undefined') {
                    clearInterval(checkInterval);
                    resolve();
                }
            }, 100);
            
            // 5秒でタイムアウト
            setTimeout(() => {
                clearInterval(checkInterval);
                resolve(); // エラーを投げずに続行
            }, 5000);
        }
    });
}

// 初期化
document.addEventListener('DOMContentLoaded', async () => {
    showLoading();
    
    // wasm_exec.jsの読み込みを待つ
    await waitForGo();
    
    // 並列で初期化
    await Promise.all([
        initWasm(),
        initPyodide()
    ]);
    
    hideLoading();
});