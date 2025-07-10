"""Grimoire CLI - Command Line Interface"""

import click
import sys
from pathlib import Path
from .compiler import compile_grimoire, run_grimoire, debug_grimoire


@click.group()
def cli():
    """Grimoire - Visual Programming Language"""
    pass


@cli.command()
@click.argument('image_file', type=click.Path(exists=True))
@click.option('-o', '--output', help='Output file name')
def compile(image_file, output):
    """Compile a magic circle image"""
    try:
        result = compile_grimoire(image_file, output)
        if output:
            click.echo(f"Compilation complete: {output}", err=True)
        else:
            click.echo(result)
    except Exception as e:
        click.echo(f"Error: {e}", err=True)
        sys.exit(1)


@cli.command()
@click.argument('image_file', type=click.Path(exists=True))
def run(image_file):
    """Run a magic circle image directly"""
    try:
        result = run_grimoire(image_file)
        click.echo(result)
    except Exception as e:
        click.echo(f"Error: {e}", err=True)
        sys.exit(1)


@cli.command()
@click.argument('image_file', type=click.Path(exists=True))
def debug(image_file):
    """Run in debug mode"""
    try:
        debug_grimoire(image_file)
    except Exception as e:
        click.echo(f"Error: {e}", err=True)
        sys.exit(1)


def main():
    """Main entry point"""
    cli()


if __name__ == "__main__":
    main()