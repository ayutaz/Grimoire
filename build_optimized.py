#!/usr/bin/env python3
"""最適化されたGrimoireバイナリをビルドするスクリプト"""

import os
import sys
import subprocess
import shutil

def build_optimized():
    """最適化されたバイナリをビルド"""
    
    print("🔧 最適化されたGrimoireバイナリをビルドします...")
    
    # 古いビルドを削除
    if os.path.exists('dist'):
        shutil.rmtree('dist')
    if os.path.exists('build'):
        shutil.rmtree('build')
    
    # PyInstallerでビルド
    cmd = [
        sys.executable, '-m', 'PyInstaller',
        '--clean',
        '--noconfirm',
        'grimoire_optimized.spec'
    ]
    
    print("📦 PyInstallerでビルド中...")
    result = subprocess.run(cmd, capture_output=True, text=True)
    
    if result.returncode != 0:
        print("❌ ビルドに失敗しました:")
        print(result.stderr)
        return False
    
    print("✅ ビルド完了!")
    
    # バイナリの場所を表示
    binary_path = os.path.join('dist', 'grimoire')
    if os.path.exists(binary_path):
        size_mb = os.path.getsize(binary_path) / (1024 * 1024)
        print(f"📍 バイナリ: {binary_path} ({size_mb:.1f} MB)")
        
        # artifactsディレクトリにコピー
        artifacts_dir = 'artifacts'
        if not os.path.exists(artifacts_dir):
            os.makedirs(artifacts_dir)
        
        dest_path = os.path.join(artifacts_dir, 'grimoire_optimized')
        shutil.copy2(binary_path, dest_path)
        os.chmod(dest_path, 0o755)
        print(f"📋 コピー先: {dest_path}")
        
        print("\n🚀 最適化されたバイナリをテストするには:")
        print(f"   ./artifacts/grimoire_optimized run examples/images/hello_world.png")
        print(f"   time ./artifacts/grimoire_optimized run examples/images/hello_world.png")
        
    return True

if __name__ == '__main__':
    success = build_optimized()
    sys.exit(0 if success else 1)