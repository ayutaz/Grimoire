"""Grimoire CLI - コマンドラインインターフェース"""

import click
import sys
from pathlib import Path
from .mock_compiler import compile_grimoire, run_grimoire, debug_grimoire


@click.group()
def cli():
    """Grimoire - 手描きプログラミング言語"""
    pass


@cli.command()
@click.argument('image_file', type=click.Path(exists=True))
@click.option('-o', '--output', help='出力ファイル名')
def compile(image_file, output):
    """画像をコンパイルする"""
    try:
        result = compile_grimoire(image_file, output)
        if output:
            click.echo(f"コンパイル完了: {output}", err=True)
        else:
            click.echo(result)
    except Exception as e:
        click.echo(f"エラー: {e}", err=True)
        sys.exit(1)


@cli.command()
@click.argument('image_file', type=click.Path(exists=True))
def run(image_file):
    """画像を直接実行する"""
    try:
        result = run_grimoire(image_file)
        click.echo(result)
    except Exception as e:
        click.echo(f"エラー: {e}", err=True)
        sys.exit(1)


@cli.command()
@click.argument('image_file', type=click.Path(exists=True))
def debug(image_file):
    """デバッグモードで実行する"""
    try:
        debug_grimoire(image_file)
    except Exception as e:
        click.echo(f"エラー: {e}", err=True)
        sys.exit(1)


def main():
    """メインエントリーポイント"""
    cli()


if __name__ == "__main__":
    main()