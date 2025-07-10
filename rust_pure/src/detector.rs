use anyhow::Result;
use image::{DynamicImage, GrayImage, Luma, ImageBuffer};
use imageproc::contours::{find_contours, Contour, BorderType};
use imageproc::contrast::{threshold, otsu_level};
use imageproc::distance_transform::distance_transform;
use imageproc::morphology::{close, open};
use imageproc::stats::histogram;
use nalgebra::{Point2, Vector2};
use std::f32::consts::PI;

#[derive(Debug, Clone, PartialEq)]
pub enum SymbolType {
    OuterCircle,
    Circle,
    DoubleCircle,
    Square,
    Triangle,
    Pentagon,
    Hexagon,
    Star,
    // 演算子
    Convergence,   // 加算
    Divergence,    // 減算
    Amplification, // 乗算
    Distribution,  // 除算
}

#[derive(Debug, Clone)]
pub struct Symbol {
    pub symbol_type: SymbolType,
    pub position: Point2<f32>,
    pub size: f32,
    pub confidence: f32,
    pub pattern: Option<String>, // 内部パターン（ドット、線など）
}

pub struct MagicCircleDetector {
    min_area: f32,
    circularity_threshold: f32,
}

impl MagicCircleDetector {
    pub fn new() -> Self {
        Self {
            min_area: 100.0,
            circularity_threshold: 0.7,
        }
    }
    
    pub fn detect_symbols(&self, image: &DynamicImage) -> Result<Vec<Symbol>> {
        // グレースケール変換
        let gray = image.to_luma8();
        
        // 適応的二値化
        let binary = self.adaptive_threshold(&gray);
        
        // 輪郭検出
        let contours = find_contours::<u8>(&binary);
        
        // シンボルを検出
        let mut symbols = Vec::new();
        
        // 1. 外円を検出
        let outer_circle = self.find_outer_circle(&contours)?;
        symbols.push(outer_circle.clone());
        
        // 2. 外円内のシンボルを検出
        for contour in &contours {
            if let Some(symbol) = self.classify_contour(contour, &binary) {
                // 外円内かチェック
                if self.is_inside_circle(&symbol.position, &outer_circle) {
                    symbols.push(symbol);
                }
            }
        }
        
        // 3. 内部パターンを検出
        for symbol in &mut symbols {
            if matches!(symbol.symbol_type, SymbolType::Square | SymbolType::Circle) {
                symbol.pattern = self.detect_internal_pattern(&binary, symbol);
            }
        }
        
        Ok(symbols)
    }
    
    fn adaptive_threshold(&self, gray: &GrayImage) -> GrayImage {
        // Otsuの手法で閾値を決定
        let level = otsu_level(&gray);
        let mut binary = threshold(&gray, level);
        
        // ノイズ除去
        binary = open(&binary, imageproc::morphology::disk(2));
        binary = close(&binary, imageproc::morphology::disk(2));
        
        binary
    }
    
    fn find_outer_circle(&self, contours: &[Contour<u8>]) -> Result<Symbol> {
        // 最大の輪郭を探す
        let largest = contours
            .iter()
            .max_by_key(|c| c.points.len())
            .ok_or_else(|| anyhow::anyhow!("No contours found"))?;
        
        // 円形度をチェック
        let (center, radius, circularity) = self.fit_circle(&largest.points);
        
        if circularity < self.circularity_threshold {
            return Err(anyhow::anyhow!("Outer circle not circular enough"));
        }
        
        Ok(Symbol {
            symbol_type: SymbolType::OuterCircle,
            position: center,
            size: radius * 2.0,
            confidence: circularity,
            pattern: None,
        })
    }
    
    fn classify_contour(&self, contour: &Contour<u8>, binary: &GrayImage) -> Option<Symbol> {
        if contour.points.len() < 5 {
            return None;
        }
        
        // 面積チェック
        let area = self.contour_area(&contour.points);
        if area < self.min_area {
            return None;
        }
        
        // 重心を計算
        let center = self.contour_center(&contour.points);
        
        // 多角形近似
        let simplified = self.approx_poly_dp(&contour.points, 0.04);
        let vertices = simplified.len();
        
        // 円形度を計算
        let (_, radius, circularity) = self.fit_circle(&contour.points);
        
        // 形状を分類
        let symbol_type = if circularity > 0.85 && vertices > 6 {
            // 円形チェック
            if self.is_double_circle(binary, center, radius) {
                SymbolType::DoubleCircle
            } else {
                SymbolType::Circle
            }
        } else {
            match vertices {
                3 => SymbolType::Triangle,
                4 => {
                    // 正方形か長方形かチェック
                    if self.is_square(&simplified) {
                        SymbolType::Square
                    } else {
                        return None;
                    }
                }
                5 => SymbolType::Pentagon,
                6 => SymbolType::Hexagon,
                n if n >= 8 && n <= 12 => {
                    // 星形かチェック
                    if self.is_star_shape(&contour.points) {
                        SymbolType::Star
                    } else {
                        SymbolType::Circle
                    }
                }
                _ => return None,
            }
        };
        
        Some(Symbol {
            symbol_type,
            position: center,
            size: (area as f32).sqrt(),
            confidence: 0.8,
            pattern: None,
        })
    }
    
    fn fit_circle(&self, points: &[imageproc::point::Point<u32>]) -> (Point2<f32>, f32, f32) {
        // 最小二乗法で円をフィッティング
        let n = points.len() as f32;
        let mut sum_x = 0.0;
        let mut sum_y = 0.0;
        
        for p in points {
            sum_x += p.x as f32;
            sum_y += p.y as f32;
        }
        
        let center_x = sum_x / n;
        let center_y = sum_y / n;
        let center = Point2::new(center_x, center_y);
        
        // 半径を計算
        let mut sum_r = 0.0;
        for p in points {
            let dx = p.x as f32 - center_x;
            let dy = p.y as f32 - center_y;
            sum_r += (dx * dx + dy * dy).sqrt();
        }
        let radius = sum_r / n;
        
        // 円形度を計算（面積比）
        let area = self.contour_area(points);
        let circle_area = PI * radius * radius;
        let circularity = (4.0 * PI * area) / (self.contour_perimeter(points).powi(2));
        
        (center, radius, circularity)
    }
    
    fn contour_area(&self, points: &[imageproc::point::Point<u32>]) -> f32 {
        // Shoelace formulaで面積を計算
        let mut area = 0.0;
        let n = points.len();
        
        for i in 0..n {
            let j = (i + 1) % n;
            area += (points[i].x * points[j].y) as f32;
            area -= (points[j].x * points[i].y) as f32;
        }
        
        area.abs() / 2.0
    }
    
    fn contour_perimeter(&self, points: &[imageproc::point::Point<u32>]) -> f32 {
        let mut perimeter = 0.0;
        let n = points.len();
        
        for i in 0..n {
            let j = (i + 1) % n;
            let dx = (points[j].x as f32) - (points[i].x as f32);
            let dy = (points[j].y as f32) - (points[i].y as f32);
            perimeter += (dx * dx + dy * dy).sqrt();
        }
        
        perimeter
    }
    
    fn contour_center(&self, points: &[imageproc::point::Point<u32>]) -> Point2<f32> {
        let n = points.len() as f32;
        let sum_x: f32 = points.iter().map(|p| p.x as f32).sum();
        let sum_y: f32 = points.iter().map(|p| p.y as f32).sum();
        Point2::new(sum_x / n, sum_y / n)
    }
    
    fn approx_poly_dp(&self, points: &[imageproc::point::Point<u32>], epsilon_ratio: f32) -> Vec<imageproc::point::Point<u32>> {
        // Douglas-Peucker algorithm（簡易版）
        let perimeter = self.contour_perimeter(points);
        let epsilon = epsilon_ratio * perimeter;
        
        // TODO: 実際のDouglas-Peucker実装
        // ここでは簡易的に元の点を返す
        points.to_vec()
    }
    
    fn is_double_circle(&self, binary: &GrayImage, center: Point2<f32>, radius: f32) -> bool {
        // 中心付近をチェックして二重円かどうか判定
        let inner_radius = radius * 0.7;
        let sample_points = 8;
        
        let mut inner_white_count = 0;
        for i in 0..sample_points {
            let angle = 2.0 * PI * (i as f32) / (sample_points as f32);
            let x = (center.x + inner_radius * angle.cos()) as u32;
            let y = (center.y + inner_radius * angle.sin()) as u32;
            
            if x < binary.width() && y < binary.height() {
                if binary.get_pixel(x, y).0[0] > 128 {
                    inner_white_count += 1;
                }
            }
        }
        
        inner_white_count >= sample_points / 2
    }
    
    fn is_square(&self, points: &[imageproc::point::Point<u32>]) -> bool {
        if points.len() != 4 {
            return false;
        }
        
        // 4つの角度が約90度かチェック
        for i in 0..4 {
            let p1 = points[i];
            let p2 = points[(i + 1) % 4];
            let p3 = points[(i + 2) % 4];
            
            let v1 = Vector2::new(
                (p1.x as f32) - (p2.x as f32),
                (p1.y as f32) - (p2.y as f32),
            );
            let v2 = Vector2::new(
                (p3.x as f32) - (p2.x as f32),
                (p3.y as f32) - (p2.y as f32),
            );
            
            let angle = v1.angle(&v2);
            if (angle - PI / 2.0).abs() > 0.3 {
                return false;
            }
        }
        
        true
    }
    
    fn is_star_shape(&self, points: &[imageproc::point::Point<u32>]) -> bool {
        // 星形の特徴：中心からの距離が交互に変化
        let center = self.contour_center(points);
        let mut distances: Vec<f32> = points
            .iter()
            .map(|p| {
                let dx = p.x as f32 - center.x;
                let dy = p.y as f32 - center.y;
                (dx * dx + dy * dy).sqrt()
            })
            .collect();
        
        // 距離の変動が大きければ星形
        let mean_dist = distances.iter().sum::<f32>() / distances.len() as f32;
        let variance = distances
            .iter()
            .map(|d| (d - mean_dist).powi(2))
            .sum::<f32>() / distances.len() as f32;
        
        variance > mean_dist * 0.2
    }
    
    fn is_inside_circle(&self, point: &Point2<f32>, circle: &Symbol) -> bool {
        let dx = point.x - circle.position.x;
        let dy = point.y - circle.position.y;
        let distance = (dx * dx + dy * dy).sqrt();
        distance < circle.size / 2.0
    }
    
    fn detect_internal_pattern(&self, binary: &GrayImage, symbol: &Symbol) -> Option<String> {
        // シンボル内部のパターンを検出
        let roi_size = (symbol.size * 0.8) as u32;
        let x = (symbol.position.x - roi_size as f32 / 2.0) as u32;
        let y = (symbol.position.y - roi_size as f32 / 2.0) as u32;
        
        if x + roi_size >= binary.width() || y + roi_size >= binary.height() {
            return None;
        }
        
        // ROI内の白ピクセル数をカウント
        let mut white_count = 0;
        for dy in 0..roi_size {
            for dx in 0..roi_size {
                if binary.get_pixel(x + dx, y + dy).0[0] > 128 {
                    white_count += 1;
                }
            }
        }
        
        let fill_ratio = white_count as f32 / (roi_size * roi_size) as f32;
        
        // パターンを分類
        match fill_ratio {
            r if r < 0.1 => Some("empty".to_string()),
            r if r < 0.2 => Some("dot".to_string()),
            r if r < 0.3 => Some("double_dot".to_string()),
            r if r < 0.5 => Some("lines".to_string()),
            r if r < 0.7 => Some("cross".to_string()),
            _ => Some("filled".to_string()),
        }
    }
}