#!/usr/bin/env python3
"""
Grimoireのクロスプラットフォームビルドスクリプト
"""

import os
import sys
import platform
import subprocess
import shutil
from pathlib import Path

def get_platform_info():
    """プラットフォーム情報を取得"""
    system = platform.system()
    machine = platform.machine()
    python_version = sys.version_info
    
    print(f"Platform: {system}")
    print(f"Architecture: {machine}")
    print(f"Python: {python_version.major}.{python_version.minor}.{python_version.micro}")
    
    return system, machine

def check_dependencies():
    """依存関係をチェック"""
    required = ['pyinstaller', 'pillow', 'opencv-python', 'numpy', 'click']
    missing = []
    
    for package in required:
        try:
            __import__(package.replace('-', '_'))
        except ImportError:
            missing.append(package)
    
    if missing:
        print(f"Missing packages: {', '.join(missing)}")
        print("Installing missing packages...")
        subprocess.run([sys.executable, '-m', 'pip', 'install'] + missing)
    
    return True

def build_binary(system):
    """プラットフォーム別のバイナリをビルド"""
    print("\nBuilding binary...")
    
    # PyInstallerのオプション
    pyinstaller_args = [
        'pyinstaller',
        '--clean',
        '--noconfirm',
        'grimoire.spec'
    ]
    
    # プラットフォーム別の追加オプション
    if system == 'Darwin':  # macOS
        # macOS用の追加設定
        pyinstaller_args.extend(['--osx-bundle-identifier', 'com.grimoire.app'])
    elif system == 'Windows':
        # Windows用の追加設定（アイコンがあれば）
        icon_path = Path('assets/grimoire.ico')
        if icon_path.exists():
            pyinstaller_args.extend(['--icon', str(icon_path)])
    
    # ビルド実行
    result = subprocess.run(pyinstaller_args, capture_output=True, text=True)
    
    if result.returncode != 0:
        print(f"Build failed:\n{result.stderr}")
        return False
    
    print("Build successful!")
    return True

def create_distribution(system):
    """配布用パッケージを作成"""
    dist_dir = Path('dist')
    release_dir = Path('release')
    release_dir.mkdir(exist_ok=True)
    
    # プラットフォーム別のパッケージ名
    if system == 'Windows':
        package_name = f'grimoire-windows-{platform.machine().lower()}.zip'
        binary_name = 'grimoire.exe'
    elif system == 'Darwin':
        package_name = f'grimoire-macos-{platform.machine().lower()}.tar.gz'
        binary_name = 'grimoire'
    else:  # Linux
        package_name = f'grimoire-linux-{platform.machine().lower()}.tar.gz'
        binary_name = 'grimoire'
    
    # バイナリの存在確認
    binary_path = dist_dir / binary_name
    if not binary_path.exists():
        print(f"Binary not found: {binary_path}")
        return False
    
    # READMEとライセンスをコピー
    for file in ['README.md', 'LICENSE']:
        if Path(file).exists():
            shutil.copy(file, dist_dir)
    
    # パッケージ作成
    os.chdir(dist_dir)
    if system == 'Windows':
        # ZIP作成
        subprocess.run(['powershell', 'Compress-Archive', '-Path', '*', '-DestinationPath', f'../{release_dir}/{package_name}'])
    else:
        # TAR.GZ作成
        subprocess.run(['tar', 'czf', f'../{release_dir}/{package_name}', '.'])
    
    os.chdir('..')
    print(f"Package created: {release_dir}/{package_name}")
    return True

def main():
    """メイン処理"""
    print("=== Grimoire Build Script ===\n")
    
    # プラットフォーム情報取得
    system, machine = get_platform_info()
    
    # 依存関係チェック
    if not check_dependencies():
        print("Failed to install dependencies")
        return 1
    
    # バイナリビルド
    if not build_binary(system):
        print("Build failed")
        return 1
    
    # 配布パッケージ作成
    if not create_distribution(system):
        print("Package creation failed")
        return 1
    
    print("\n=== Build completed successfully! ===")
    return 0

if __name__ == '__main__':
    sys.exit(main())