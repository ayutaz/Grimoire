#!/usr/bin/env python3
# Create placeholder sample images
from PIL import Image, ImageDraw, ImageFont
import os

samples = {
    'hello-world': 'Hello World',
    'calculator': '電卓',
    'fibonacci': 'フィボナッチ',
    'loop': 'ループ'
}

# Create sample directory if not exists
os.makedirs('static/samples', exist_ok=True)

for filename, text in samples.items():
    # Create image
    img = Image.new('RGB', (300, 300), color='white')
    draw = ImageDraw.Draw(img)
    
    # Draw outer circle
    draw.ellipse([10, 10, 290, 290], outline='black', width=3)
    
    # Draw center double circle (entry point)
    draw.ellipse([130, 130, 170, 170], outline='black', width=2)
    draw.ellipse([135, 135, 165, 165], outline='black', width=2)
    
    # Draw some shapes
    # Square
    draw.rectangle([80, 80, 120, 120], outline='black', width=2)
    
    # Star (output)
    draw.polygon([(150, 200), (160, 220), (140, 220)], outline='black', width=2)
    
    # Connection lines
    draw.line([(150, 170), (150, 200)], fill='black', width=2)
    draw.line([(130, 150), (100, 100)], fill='black', width=2)
    
    # Add text label
    try:
        font = ImageFont.truetype("/System/Library/Fonts/Helvetica.ttc", 20)
    except:
        font = None
    draw.text((100, 250), text, fill='black', font=font)
    
    # Save
    img.save(f'static/samples/{filename}.png')
    print(f'Created static/samples/{filename}.png')

print('Sample images created successfully!')