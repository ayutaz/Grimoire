use opencv::{
    prelude::*,
    core,
    imgcodecs,
    imgproc,
    types::VectorOfVectorOfPoint,
};
use anyhow::Result;
use clap::{Parser, Subcommand};
use std::path::PathBuf;
use std::time::Instant;

#[derive(Parser)]
#[command(name = "grimoire")]
#[command(about = "A visual programming language using magic circles")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    Run {
        /// Path to the image file
        path: PathBuf,
    },
    Compile {
        /// Path to the image file
        path: PathBuf,
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
    min_contour_area: f64,
    circle_threshold: f64,
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
        
        // OpenCVで画像を読み込み
        let img = imgcodecs::imread(
            image_path.to_str().unwrap(),
            imgcodecs::IMREAD_COLOR,
        )?;
        
        println!("Image loaded in {:?}", start.elapsed());
        
        // グレースケール変換
        let mut gray = Mat::default();
        imgproc::cvt_color(&img, &mut gray, imgproc::COLOR_BGR2GRAY, 0)?;
        
        // 二値化
        let mut binary = Mat::default();
        imgproc::threshold(
            &gray,
            &mut binary,
            0.0,
            255.0,
            imgproc::THRESH_BINARY_INV | imgproc::THRESH_OTSU,
        )?;
        
        // 輪郭検出
        let mut contours = VectorOfVectorOfPoint::new();
        let mut hierarchy = Mat::default();
        imgproc::find_contours(
            &binary,
            &mut contours,
            &mut hierarchy,
            imgproc::RETR_EXTERNAL,
            imgproc::CHAIN_APPROX_SIMPLE,
            core::Point::new(0, 0),
        )?;
        
        let mut symbols = Vec::new();
        
        // 外円を検出
        if let Some(outer_circle) = self.find_outer_circle(&contours)? {
            symbols.push(outer_circle);
            
            // 他のシンボルを検出
            self.detect_other_symbols(&binary, &contours, &mut symbols)?;
        } else {
            return Err(anyhow::anyhow!("No outer circle detected"));
        }
        
        println!("Total detection time: {:?}", start.elapsed());
        
        Ok(symbols)
    }
    
    fn find_outer_circle(&self, contours: &VectorOfVectorOfPoint) -> Result<Option<Symbol>> {
        let mut max_area = 0.0;
        let mut max_idx = None;
        
        // 最大の輪郭を見つける
        for i in 0..contours.len() {
            let contour = contours.get(i)?;
            let area = imgproc::contour_area(&contour, false)?;
            
            if area > max_area && area > self.min_contour_area * 10.0 {
                max_area = area;
                max_idx = Some(i);
            }
        }
        
        if let Some(idx) = max_idx {
            let contour = contours.get(idx)?;
            
            // 円形度をチェック
            let mut center = core::Point2f::default();
            let mut radius = 0.0f32;
            imgproc::min_enclosing_circle(&contour, &mut center, &mut radius)?;
            
            let circle_area = std::f64::consts::PI * (radius as f64) * (radius as f64);
            let circularity = max_area / circle_area;
            
            if circularity > self.circle_threshold {
                return Ok(Some(Symbol {
                    symbol_type: SymbolType::OuterCircle,
                    position: (center.x, center.y),
                    size: radius * 2.0,
                    confidence: circularity as f32,
                }));
            }
        }
        
        Ok(None)
    }
    
    fn detect_other_symbols(
        &self,
        binary: &Mat,
        contours: &VectorOfVectorOfPoint,
        symbols: &mut Vec<Symbol>,
    ) -> Result<()> {
        // 各輪郭を解析
        for i in 0..contours.len() {
            let contour = contours.get(i)?;
            let area = imgproc::contour_area(&contour, false)?;
            
            if area < self.min_contour_area {
                continue;
            }
            
            // 多角形近似
            let mut approx = Mat::default();
            let epsilon = 0.04 * imgproc::arc_length(&contour, true)?;
            imgproc::approx_poly_dp(&contour, &mut approx, epsilon, true)?;
            
            let vertices = approx.rows();
            
            // モーメントから重心を計算
            let moments = imgproc::moments(&contour, false)?;
            let cx = (moments.m10 / moments.m00) as f32;
            let cy = (moments.m01 / moments.m00) as f32;
            
            // 頂点数に基づいて形状を分類
            let symbol_type = match vertices {
                3 => Some(SymbolType::Triangle),
                4 => Some(SymbolType::Square),
                n if n >= 8 => Some(SymbolType::Star),
                _ => None,
            };
            
            if let Some(st) = symbol_type {
                symbols.push(Symbol {
                    symbol_type: st,
                    position: (cx, cy),
                    size: (area as f32).sqrt(),
                    confidence: 0.8,
                });
            }
        }
        
        Ok(())
    }
}

fn run_program(path: &PathBuf) -> Result<()> {
    let detector = MagicCircleDetector::new();
    let symbols = detector.detect_symbols(path)?;
    
    // シンプルなHello World判定
    if symbols.iter().any(|s| matches!(s.symbol_type, SymbolType::OuterCircle)) {
        println!("Hello, World!");
    }
    
    Ok(())
}

fn compile_program(path: &PathBuf, output: Option<PathBuf>) -> Result<()> {
    let detector = MagicCircleDetector::new();
    let _symbols = detector.detect_symbols(path)?;
    
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