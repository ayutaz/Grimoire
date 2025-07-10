use anyhow::Result;
use clap::{Parser, Subcommand};
use image::{DynamicImage, GrayImage, Luma};
use imageproc::contours::{find_contours, BorderType};
use imageproc::contrast::threshold;
use std::path::PathBuf;
use std::time::Instant;

#[derive(Parser)]
#[command(name = "grimoire")]
#[command(about = "A visual programming language using magic circles", long_about = None)]
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
}

#[derive(Debug, Clone)]
enum SymbolType {
    OuterCircle,
    Circle,
    Square,
    Triangle,
    Star,
}

#[derive(Debug, Clone)]
struct Symbol {
    symbol_type: SymbolType,
    position: (f32, f32),
    size: f32,
    confidence: f32,
}

struct MagicCircleDetector {
    min_contour_area: f32,
    circle_threshold: f32,
}

impl MagicCircleDetector {
    fn new() -> Self {
        Self {
            min_contour_area: 100.0,
            circle_threshold: 0.8,
        }
    }

    fn detect_symbols(&self, image_path: &PathBuf) -> Result<Vec<Symbol>> {
        let start = Instant::now();
        
        // 画像を読み込み
        let img = image::open(image_path)?;
        let gray = img.to_luma8();
        
        println!("Image loaded in {:?}", start.elapsed());
        
        // 二値化
        let binary = threshold(&gray, 128);
        
        // 輪郭検出
        let contours = find_contours::<u8>(&binary);
        
        let mut symbols = Vec::new();
        
        // 外円を検出
        let outer_circle = self.find_outer_circle(&contours)?;
        symbols.push(outer_circle);
        
        // TODO: 他のシンボルを検出
        
        println!("Total detection time: {:?}", start.elapsed());
        
        Ok(symbols)
    }
    
    fn find_outer_circle(&self, contours: &[Vec<imageproc::point::Point<u32>>]) -> Result<Symbol> {
        // 最大の輪郭を見つける
        let largest = contours
            .iter()
            .max_by_key(|c| c.len())
            .ok_or_else(|| anyhow::anyhow!("No contours found"))?;
        
        // 簡易的な円判定
        let center_x = largest.iter().map(|p| p.x).sum::<u32>() as f32 / largest.len() as f32;
        let center_y = largest.iter().map(|p| p.y).sum::<u32>() as f32 / largest.len() as f32;
        
        Ok(Symbol {
            symbol_type: SymbolType::OuterCircle,
            position: (center_x, center_y),
            size: 100.0, // 仮の値
            confidence: 0.9,
        })
    }
}

fn run_program(path: &PathBuf) -> Result<()> {
    let detector = MagicCircleDetector::new();
    let symbols = detector.detect_symbols(path)?;
    
    // シンプルなHello Worldの判定
    if symbols.iter().any(|s| matches!(s.symbol_type, SymbolType::OuterCircle)) {
        println!("Hello, World!");
    }
    
    Ok(())
}

fn compile_program(path: &PathBuf, output: Option<PathBuf>) -> Result<()> {
    let detector = MagicCircleDetector::new();
    let symbols = detector.detect_symbols(path)?;
    
    let python_code = "print('Hello, World!')";
    
    if let Some(output_path) = output {
        std::fs::write(output_path, python_code)?;
    } else {
        println!("{}", python_code);
    }
    
    Ok(())
}

fn main() -> Result<()> {
    let cli = Cli::parse();
    
    match cli.command {
        Commands::Run { path } => run_program(&path),
        Commands::Compile { path, output } => compile_program(&path, output),
    }
}