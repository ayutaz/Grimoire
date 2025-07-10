use std::env;

fn main() {
    // プラットフォーム別のOpenCV設定
    let target_os = env::var("CARGO_CFG_TARGET_OS").unwrap();
    
    match target_os.as_str() {
        "macos" => {
            // macOSの場合
            println!("cargo:rustc-link-search=/opt/homebrew/lib");
            println!("cargo:rustc-link-search=/usr/local/lib");
        }
        "linux" => {
            // Linuxの場合
            println!("cargo:rustc-link-search=/usr/lib/x86_64-linux-gnu");
            println!("cargo:rustc-link-search=/usr/local/lib");
        }
        "windows" => {
            // Windowsの場合
            println!("cargo:rustc-link-search=C:\\tools\\opencv\\build\\x64\\vc15\\lib");
        }
        _ => {}
    }
}