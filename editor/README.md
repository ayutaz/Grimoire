# Grimoire Visual Editor

A web-based visual editor for creating Grimoire magic circles. Draw your magical programs visually and export them as PNG images or Grimoire code.

## Features

- **Visual Drawing**: Create magic circles using drag-and-drop symbols
- **Symbol Library**: Complete set of Grimoire symbols organized by category
- **Connection Tool**: Draw energy flows between symbols
- **Export Options**: Save as PNG image or JSON project file
- **Grid Snapping**: Align elements perfectly with adjustable grid
- **Zoom & Pan**: Navigate large magic circles easily
- **Keyboard Shortcuts**: Efficient workflow with hotkeys

## Quick Start

### Running the Editor

1. Navigate to the editor directory:
   ```bash
   cd editor
   ```

2. Start the server:
   ```bash
   go run server.go
   ```

3. Open your browser to: `http://localhost:8080`

### Alternative: Python Server

If you don't have Go installed, you can use Python:

```bash
# Python 3
python -m http.server 8080

# Python 2
python -m SimpleHTTPServer 8080
```

## User Guide

### Tools

- **Select Tool (V)**: Select, move, and resize elements
- **Circle Tool (C)**: Draw circles and boundaries
- **Connect Tool (L)**: Create connections between symbols
- **Text Tool (T)**: Add text labels

### Adding Symbols

1. **Drag & Drop**: Drag symbols from the palette to the canvas
2. **Click to Place**: Click a symbol, then click on the canvas

### Creating Connections

1. Select the Connect tool (L)
2. Click on the first symbol
3. Click on the second symbol
4. Right-click connections to change their style

### Keyboard Shortcuts

- **V**: Select tool
- **C**: Circle tool
- **L**: Connect tool
- **T**: Text tool
- **Delete**: Delete selected element
- **Ctrl/Cmd + S**: Save project
- **Ctrl/Cmd + O**: Open project
- **Ctrl/Cmd + E**: Export to PNG
- **Escape**: Deselect all

### Mouse Controls

- **Left Click**: Select/place elements
- **Middle Click** or **Shift + Left Click**: Pan canvas
- **Mouse Wheel**: Zoom in/out
- **Right Click**: Context menu (on connections)

## Symbol Reference

### Structure Elements
- **⭕ Outer Circle**: Magic circle boundary
- **○ Inner Circle**: Internal scope
- **◎ Double Circle**: Main entry point
- **⭐ Pentagram**: Function definition
- **✡ Hexagram**: Parallel processing
- **✦ Octagram**: Class definition
- **△ Triangle**: Conditional branch
- **□ Square**: Data storage

### Mystical Operators
- **⟐ Fusion**: Addition
- **⟑ Separation**: Subtraction
- **✦ Amplification**: Multiplication
- **⟠ Division**: Division
- **⟷ Transfer**: Assignment
- **⊗ Seal**: Constant
- **⟳ Cycle**: Loop

### Comparison Symbols
- **= Balance**: Equality
- **≠ Imbalance**: Inequality
- **< Descent**: Less than
- **> Ascent**: Greater than
- **≤ Earth**: Less or equal
- **≥ Heaven**: Greater or equal

### Logic Symbols
- **⊕ Light Union**: AND
- **⊖ Light Choice**: OR
- **⊗ Light Inversion**: NOT
- **⊙ Light Exclusion**: XOR

### Energy Nodes
- **⬢ Hexagonal Crystal**: Branch point
- **◈ Square Crystal**: Aggregation point
- **⬟ Pentagon Crystal**: Transformation point
- **✧ Star Crystal**: Amplification point

### Special Symbols
- **☀ Sun**: Start/True
- **☾ Moon**: False/Alternative
- **☆ Star**: Output/Result
- **○○ Double Circle**: Function call
- **♪ Musical Note**: Sound/Event
- **✉ Envelope**: Message/Communication
- **✓ Checkmark**: Success/Complete

## Examples

### Hello World
```
1. Add an Outer Circle (⭕)
2. Place a Star (☆) in the center
3. Export to PNG
```

### Conditional Branch
```
1. Add an Outer Circle (⭕)
2. Place a Triangle (△) at the top
3. Add Sun (☀) on the left branch
4. Add Moon (☾) on the right branch
5. Connect Triangle to both symbols
```

### Function Definition
```
1. Add an Outer Circle (⭕)
2. Add an Inner Circle (○) for function scope
3. Place Data Nodes (□•) for parameters
4. Add operator symbol (e.g., +)
5. Connect parameters to operator
6. Add Star (☆) for output
7. Connect operator to output
```

## Extending the Editor

### Adding New Symbols

Edit `js/symbols.js` to add new symbols:

```javascript
'new-symbol': {
    symbol: '🔮',
    name: 'Crystal Ball',
    category: 'special',
    defaultSize: 60
}
```

### Custom Connection Types

Edit `js/connections.js` to add connection styles:

```javascript
'magic_flow': {
    name: 'Magic Flow',
    strokeStyle: '#9400D3',
    lineWidth: 3,
    dashArray: [10, 5],
    arrows: 'both'
}
```

### Themes

Modify `css/style.css` to create custom themes. The editor supports both light and dark modes.

## Project File Format

Projects are saved as JSON with the following structure:

```json
{
  "version": "1.0",
  "canvas": {
    "width": 800,
    "height": 600,
    "grid": true,
    "gridSize": 20
  },
  "elements": [
    {
      "id": "elem_0",
      "type": "outer-circle",
      "symbol": "⭕",
      "x": 400,
      "y": 300,
      "size": 300,
      "rotation": 0,
      "properties": {}
    }
  ],
  "connections": [
    {
      "id": "conn_0",
      "from": "elem_0",
      "to": "elem_1",
      "type": "energy_flow",
      "style": "solid"
    }
  ]
}
```

## Troubleshooting

### Symbols Not Displaying
- Ensure your browser supports Unicode symbols
- Try a different font in the CSS

### Cannot Save/Load Files
- Check browser permissions for file access
- Try a different browser if issues persist

### Performance Issues
- Reduce the number of elements
- Disable grid display
- Use a modern browser with hardware acceleration

## Future Enhancements

- [ ] Undo/Redo functionality
- [ ] Copy/Paste elements
- [ ] Alignment tools
- [ ] Layers support
- [ ] Animation preview
- [ ] Direct Grimoire code generation
- [ ] Collaborative editing
- [ ] Mobile/tablet support
- [ ] SVG export
- [ ] Integration with Grimoire compiler

## License

This editor is part of the Grimoire project and shares the same license.