#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { JSDOM } = require('jsdom');

// テスト結果を記録
let testsPassed = 0;
let testsFailed = 0;

function assert(condition, message) {
    if (condition) {
        testsPassed++;
        console.log(`✓ ${message}`);
    } else {
        testsFailed++;
        console.error(`✗ ${message}`);
        throw new Error(message);
    }
}

async function runTests() {
    console.log('Starting WASM tests...\n');

    // 1. wasm_exec.jsの存在確認
    const wasmExecPath = path.join(__dirname, 'static', 'wasm_exec.js');
    assert(fs.existsSync(wasmExecPath), 'wasm_exec.js should exist');

    // 2. grimoire.wasmの存在確認
    const wasmPath = path.join(__dirname, 'static', 'wasm', 'grimoire.wasm');
    assert(fs.existsSync(wasmPath), 'grimoire.wasm should exist');

    // 3. DOM環境のセットアップ
    const dom = new JSDOM('<!DOCTYPE html><html><body></body></html>', {
        url: 'http://localhost:8080',
        runScripts: 'dangerously',
        resources: 'usable'
    });

    // Node.js環境でブラウザ環境をエミュレート
    global.window = dom.window;
    global.document = dom.window.document;
    global.fetch = require('node-fetch');
    global.global = global;
    
    // TextEncoderとTextDecoderのポリフィル
    const { TextEncoder, TextDecoder } = require('util');
    global.TextEncoder = TextEncoder;
    global.TextDecoder = TextDecoder;
    dom.window.TextEncoder = TextEncoder;
    dom.window.TextDecoder = TextDecoder;
    
    // cryptoのポリフィル
    const crypto = require('crypto');
    global.crypto = {
        getRandomValues: (buf) => {
            const bytes = crypto.randomBytes(buf.length);
            buf.set(bytes);
            return buf;
        }
    };
    dom.window.crypto = global.crypto;
    
    // performanceのポリフィル
    global.performance = {
        now: () => {
            const [sec, nsec] = process.hrtime();
            return sec * 1000 + nsec / 1000000;
        }
    };
    dom.window.performance = global.performance;
    
    // fsのモック
    global.fs = {
        constants: { O_WRONLY: -1, O_RDWR: -1, O_CREAT: -1, O_TRUNC: -1, O_APPEND: -1, O_EXCL: -1 },
        writeSync(fd, buf) {
            const text = new TextDecoder().decode(buf);
            process.stdout.write(text);
            return buf.length;
        },
        write(fd, buf, offset, length, position, callback) {
            const n = this.writeSync(fd, buf);
            callback(null, n);
        }
    };
    
    // processのモック
    global.process = {
        getuid() { return -1; },
        getgid() { return -1; },
        geteuid() { return -1; },
        getegid() { return -1; },
        getgroups() { throw new Error('not implemented'); },
        pid: -1,
        ppid: -1,
        umask() { throw new Error('not implemented'); },
        cwd() { throw new Error('not implemented'); },
        chdir() { throw new Error('not implemented'); },
        hrtime: process.hrtime
    };
    
    // WebAssembly.instantiateStreamingのポリフィル
    if (!WebAssembly.instantiateStreaming) {
        WebAssembly.instantiateStreaming = async (resp, importObject) => {
            const source = await (await resp).arrayBuffer();
            return await WebAssembly.instantiate(source, importObject);
        };
    }

    // 4. wasm_exec.jsを読み込む
    const wasmExecCode = fs.readFileSync(wasmExecPath, 'utf8');
    // wasm_exec.jsをevalで実行（Node.js環境で実行）
    eval(wasmExecCode);
    
    // Goオブジェクトが定義されているか確認
    assert(typeof global.Go !== 'undefined', 'Go object should be defined after loading wasm_exec.js');

    // 5. WASMの初期化テスト
    console.log('\nTesting WASM initialization...');
    
    const go = new global.Go();
    assert(go !== null, 'Go instance should be created');
    assert(typeof go.importObject === 'object', 'Go.importObject should be an object');

    // 6. WASMのロードとインスタンス化
    console.log('\nLoading and instantiating WASM...');
    
    try {
        const wasmBuffer = fs.readFileSync(wasmPath);
        const result = await WebAssembly.instantiate(wasmBuffer, go.importObject);
        
        assert(result.instance !== null, 'WASM instance should be created');
        
        // Goランタイムを実行（非同期で実行）
        go.run(result.instance);
        
        // WASMが初期化されるまで待つ
        await new Promise(resolve => setTimeout(resolve, 500));
        
        // グローバル関数が登録されているか確認
        assert(typeof global.processGrimoireImage === 'function', 'processGrimoireImage should be registered');
        assert(typeof global.validateGrimoireCode === 'function', 'validateGrimoireCode should be registered');
        assert(typeof global.formatGrimoireCode === 'function', 'formatGrimoireCode should be registered');
        
    } catch (error) {
        console.error('WASM initialization error:', error);
        testsFailed++;
        throw error;
    }

    // 7. processGrimoireImage関数のテスト
    console.log('\nTesting processGrimoireImage function...');
    
    try {
        // 無効な入力のテスト
        const result1 = global.processGrimoireImage();
        assert(result1 && result1.success === false, 'Should handle missing arguments');
        assert(result1.error === 'No image data provided', 'Should return correct error message');
        
        // 無効なBase64のテスト
        const result2 = global.processGrimoireImage('invalid-base64');
        assert(result2 && result2.success === false, 'Should handle invalid base64');
        assert(result2.error && result2.error.includes('Failed to decode image'), 'Should return decode error');
        
        // 有効なBase64（1x1の白いピクセル）のテスト
        const validBase64 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==';
        const result3 = global.processGrimoireImage(validBase64);
        assert(result3 && typeof result3 === 'object', 'Should return an object for valid input');
        assert(result3.success === true, 'Should succeed with valid image');
        assert(typeof result3.code === 'string', 'Should return Python code');
        assert(result3.code.includes('#!/usr/bin/env python3') || result3.code.includes('print'), 'Code should be valid Python');
        
        // nil値やundefinedが含まれていないか確認
        const jsonStr = JSON.stringify(result3);
        assert(!jsonStr.includes('null') || result3.ast === null, 'Should not have unexpected null values');
        
        // debugInfoがある場合、それがJSON文字列であることを確認
        if (result3.debug) {
            assert(typeof result3.debug === 'string', 'debug should be a JSON string');
            try {
                const debugObj = JSON.parse(result3.debug);
                assert(typeof debugObj === 'object', 'debug should be parseable as JSON');
            } catch (e) {
                assert(false, 'debug should be valid JSON: ' + e.message);
            }
        }
        
        // astがある場合、それがJSON文字列であることを確認
        if (result3.ast) {
            assert(typeof result3.ast === 'string', 'ast should be a JSON string');
            try {
                const astObj = JSON.parse(result3.ast);
                assert(typeof astObj === 'object', 'ast should be parseable as JSON');
            } catch (e) {
                assert(false, 'ast should be valid JSON: ' + e.message);
            }
        }
        
        // JavaScriptに渡せない値が含まれていないか確認（ValueOfエラーの原因となる値）
        function checkForInvalidValues(obj, path = '') {
            if (obj === null || obj === undefined) return;
            
            for (const key in obj) {
                const value = obj[key];
                const currentPath = path ? `${path}.${key}` : key;
                
                // 値の型をチェック
                const valueType = typeof value;
                assert(
                    valueType === 'string' || 
                    valueType === 'number' || 
                    valueType === 'boolean' || 
                    valueType === 'object',
                    `Invalid type at ${currentPath}: ${valueType}`
                );
                
                // オブジェクトの場合は再帰的にチェック
                if (valueType === 'object' && value !== null) {
                    checkForInvalidValues(value, currentPath);
                }
            }
        }
        
        checkForInvalidValues(result3);
        
    } catch (error) {
        console.error('Function test error:', error);
        testsFailed++;
        throw error;
    }

    // 8. より複雑な画像でのテスト（実際のGrimoireプログラムを模擬）
    console.log('\nTesting with complex image data...');
    
    try {
        // 100x100の白い画像（シンボルが検出される可能性のあるサイズ）
        const canvas = require('canvas');
        const createCanvas = canvas.createCanvas;
        const imageCanvas = createCanvas(100, 100);
        const ctx = imageCanvas.getContext('2d');
        
        // 白い背景
        ctx.fillStyle = 'white';
        ctx.fillRect(0, 0, 100, 100);
        
        // 黒い円を描画（外側の円）
        ctx.strokeStyle = 'black';
        ctx.lineWidth = 2;
        ctx.beginPath();
        ctx.arc(50, 50, 40, 0, 2 * Math.PI);
        ctx.stroke();
        
        // Base64に変換
        const imageData = imageCanvas.toDataURL().split(',')[1];
        const result = global.processGrimoireImage(imageData);
        
        assert(result && typeof result === 'object', 'Should handle complex image');
        assert(result.success === true, 'Should process complex image successfully');
        
        // debugInfoの構造を詳しくチェック
        if (result.debug) {
            const debugObj = JSON.parse(result.debug);
            assert(typeof debugObj.symbolCount === 'number', 'symbolCount should be a number');
            assert(Array.isArray(debugObj.symbols), 'symbols should be an array');
            
            // 各シンボルの構造をチェック
            debugObj.symbols.forEach((symbol, index) => {
                assert(typeof symbol.type === 'string', `Symbol ${index}: type should be string`);
                assert(typeof symbol.position === 'object', `Symbol ${index}: position should be object`);
                assert(typeof symbol.position.x === 'number', `Symbol ${index}: position.x should be number`);
                assert(typeof symbol.position.y === 'number', `Symbol ${index}: position.y should be number`);
                if (symbol.pattern !== undefined) {
                    assert(typeof symbol.pattern === 'string', `Symbol ${index}: pattern should be string`);
                }
            });
        }
        
    } catch (error) {
        // canvas モジュールがない場合はスキップ
        if (error.code === 'MODULE_NOT_FOUND') {
            console.log('  (Skipping complex image test - canvas module not available)');
        } else {
            console.error('Complex image test error:', error);
            testsFailed++;
            throw error;
        }
    }
    
    // 9. 過去のバグの再現テスト
    console.log('\nTesting past bug patterns...');
    
    try {
        // シンプルな画像でテスト（1x1の白い画像）
        function createGrimoireImage() {
            // 1x1の白い画像（validBase64と同じ）
            const simpleImage = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==';
            return simpleImage;
        }
        
        const complexImage = createGrimoireImage();
        const complexResult = global.processGrimoireImage(complexImage);
        
        // 結果の基本的な検証
        assert(complexResult && complexResult.success === true, 'Should process complex image');
        assert(typeof complexResult.code === 'string', 'Should return code string');
        
        // debugInfoの詳細な検証
        if (complexResult.debug) {
            const debugObj = JSON.parse(complexResult.debug);
            
            // シンボルが検出された場合の検証
            if (debugObj.symbolCount > 0) {
                assert(Array.isArray(debugObj.symbols), 'symbols should be an array');
                
                // 各シンボルでPatternフィールドの処理を確認（過去のバグ）
                debugObj.symbols.forEach((symbol, index) => {
                    assert(typeof symbol.type === 'string', `Symbol ${index}: type should be string`);
                    assert(typeof symbol.position === 'object', `Symbol ${index}: position should be object`);
                    assert(typeof symbol.position.x === 'number', `Symbol ${index}: x should be number`);
                    assert(typeof symbol.position.y === 'number', `Symbol ${index}: y should be number`);
                    
                    // Patternフィールドは空文字列の場合は含まれないはず
                    if (symbol.pattern !== undefined) {
                        assert(symbol.pattern !== '', `Symbol ${index}: pattern should not be empty string`);
                        assert(typeof symbol.pattern === 'string', `Symbol ${index}: pattern should be string`);
                    }
                });
            }
        }
        
        // 全体のJSONシリアライズ可能性をテスト
        try {
            const fullJSON = JSON.stringify(complexResult);
            assert(fullJSON.length > 0, 'Result should be JSON serializable');
        } catch (e) {
            assert(false, 'Result should be fully JSON serializable: ' + e.message);
        }
        
    } catch (error) {
        console.error('Past bug pattern test error:', error);
        testsFailed++;
        throw error;
    }
    
    // 10. WASM初期化パスのテスト（過去のバグ）
    console.log('\nTesting WASM initialization paths...');
    
    try {
        // app.jsで使用されるパスが正しいことを確認
        const appJsPath = path.join(__dirname, 'static', 'app.js');
        if (fs.existsSync(appJsPath)) {
            const appJsContent = fs.readFileSync(appJsPath, 'utf8');
            
            // WASMファイルのパスが正しいことを確認
            assert(appJsContent.includes('static/wasm/grimoire.wasm') || 
                   appJsContent.includes('./static/wasm/grimoire.wasm'),
                   'app.js should reference correct WASM path');
            
            // wasm_exec.jsのパスはindex.htmlで参照されているのでスキップ
            console.log('  (wasm_exec.js is referenced in index.html, not app.js)');
        }
        
        // build-wasm.shが正しいディレクトリに出力することを確認
        const buildScriptPath = path.join(__dirname, 'build-wasm.sh');
        if (fs.existsSync(buildScriptPath)) {
            const buildScriptContent = fs.readFileSync(buildScriptPath, 'utf8');
            assert(buildScriptContent.includes('static/wasm'),
                   'build-wasm.sh should output to static/wasm directory');
        }
        
    } catch (error) {
        console.error('Path test error:', error);
        testsFailed++;
        throw error;
    }
    
    // 11. メモリリークテスト（複数回実行）
    console.log('\nTesting memory stability...');
    
    try {
        const validBase64 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==';
        
        for (let i = 0; i < 10; i++) {
            const result = global.processGrimoireImage(validBase64);
            assert(result && result.success === true, `Iteration ${i + 1} should succeed`);
        }
        
    } catch (error) {
        console.error('Memory test error:', error);
        testsFailed++;
        throw error;
    }
    
    // 12. Go 1.21との互換性テスト（型変換の明示的チェック）
    console.log('\nTesting Go type conversions...');
    
    try {
        // 実際の画像処理をテストして、すべての数値が正しくfloat64に変換されているか確認
        const testImage = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==';
        const result = global.processGrimoireImage(testImage);
        
        if (result.debug) {
            const debugObj = JSON.parse(result.debug);
            
            // symbolCountが数値であることを確認
            assert(typeof debugObj.symbolCount === 'number', 'symbolCount should be a number, not a Go int');
            
            // シンボルがある場合、座標が正しく数値に変換されているか確認
            if (debugObj.symbols && debugObj.symbols.length > 0) {
                debugObj.symbols.forEach((symbol, i) => {
                    assert(typeof symbol.position.x === 'number', `Symbol ${i}: x should be number`);
                    assert(typeof symbol.position.y === 'number', `Symbol ${i}: y should be number`);
                    // 整数値でもJavaScriptでは浮動小数点として扱われることを確認
                    assert(Number.isFinite(symbol.position.x), `Symbol ${i}: x should be finite number`);
                    assert(Number.isFinite(symbol.position.y), `Symbol ${i}: y should be finite number`);
                });
            }
        }
        
    } catch (error) {
        console.error('Type conversion test error:', error);
        testsFailed++;
        throw error;
    }
    
    // 13. 実際のサンプル画像でのテスト
    console.log('\nTesting with actual sample images...');
    
    try {
        // hello-world.pngのBase64データを読み込む（実際のデータが必要）
        const fs = require('fs');
        const helloWorldPath = path.join(__dirname, 'static', 'samples', 'hello-world.png');
        
        if (fs.existsSync(helloWorldPath)) {
            const imageBuffer = fs.readFileSync(helloWorldPath);
            const base64Image = imageBuffer.toString('base64');
            
            console.log('  Processing hello-world.png...');
            const result = global.processGrimoireImage(base64Image);
            
            console.log('  Result:', JSON.stringify(result, null, 2));
            
            assert(result && result.success === true, 'Should process hello-world.png successfully');
            assert(typeof result.code === 'string', 'Should return Python code');
            
            // Hello Worldが含まれているか確認
            assert(result.code.includes('print') || result.code.includes('Hello'), 
                   'Hello world sample should generate print statement or Hello text');
            
            // while Falseループが含まれていないか確認
            assert(!result.code.includes('while False'), 
                   'Should not generate while False loops');
        } else {
            console.log('  (Skipping hello-world.png test - file not found)');
        }
        
    } catch (error) {
        console.error('Sample image test error:', error);
        testsFailed++;
        throw error;
    }

    // クリーンアップ
    dom.window.close();
}

// メイン実行
(async () => {
    try {
        await runTests();
        
        console.log('\n========================================');
        console.log(`Tests passed: ${testsPassed}`);
        console.log(`Tests failed: ${testsFailed}`);
        console.log('========================================\n');
        
        if (testsFailed > 0) {
            process.exitCode = 1;
        }
        
    } catch (error) {
        console.error('\nTest suite failed:', error);
        process.exitCode = 1;
    }
})();