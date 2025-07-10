use anyhow::Result;
use clap::{Parser, Subcommand};
use image::{DynamicImage, GrayImage, Luma};
use imageproc::contours::{find_contours, Contour};
use imageproc::contrast::threshold;
use std::path::PathBuf;
use std::time::Instant;

mod detector;
mod parser;
mod compiler;

use detector::{MagicCircleDetector, Symbol};
use parser::MagicCircleParser;
use compiler::PythonCompiler;

#[derive(Parser)]
#[command(name = "grimoire")]
#[command(about = "A visual programming language using magic circles")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Run a Grimoire program
    Run {
        /// Path to the image file
        path: PathBuf,
    },
    /// Compile a Grimoire program to Python
    Compile {
        /// Path to the image file
        path: PathBuf,
        /// Output file path
        #[arg(short, long)]
        output: Option<PathBuf>,
    },
    /// Debug mode - show detected symbols
    Debug {
        /// Path to the image file
        path: PathBuf,
    },
}

fn main() -> Result<()> {
    let cli = Cli::parse();
    let start = Instant::now();
    
    match cli.command {
        Commands::Run { path } => {
            let (ast, _) = process_image(&path)?;
            let code = PythonCompiler::compile(&ast)?;
            
            // Pythonコードを実行
            let output = std::process::Command::new("python3")
                .arg("-c")
                .arg(&code)
                .output()?;
            
            print!("{}", String::from_utf8_lossy(&output.stdout));
            eprintln!("{}", String::from_utf8_lossy(&output.stderr));
        }
        Commands::Compile { path, output } => {
            let (ast, _) = process_image(&path)?;
            let code = PythonCompiler::compile(&ast)?;
            
            if let Some(output_path) = output {
                std::fs::write(output_path, code)?;
            } else {
                println!("{}", code);
            }
        }
        Commands::Debug { path } => {
            let (_, symbols) = process_image(&path)?;
            println!("Detected {} symbols:", symbols.len());
            for symbol in symbols {
                println!("  {:?}", symbol);
            }
        }
    }
    
    eprintln!("Execution time: {:?}", start.elapsed());
    Ok(())
}

fn process_image(path: &PathBuf) -> Result<(parser::AST, Vec<Symbol>)> {
    // 画像を読み込み
    let img = image::open(path)?;
    
    // シンボルを検出
    let detector = MagicCircleDetector::new();
    let symbols = detector.detect_symbols(&img)?;
    
    // ASTに変換
    let parser = MagicCircleParser::new();
    let ast = parser.parse(&symbols)?;
    
    Ok((ast, symbols))
}