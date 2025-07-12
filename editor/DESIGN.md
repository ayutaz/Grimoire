# Grimoire Visual Editor Design Document

## Overview

The Grimoire Visual Editor is a web-based tool that allows users to create magic circles visually using drag-and-drop functionality and drawing tools. The editor generates magic circle diagrams that can be exported as PNG images and processed by the Grimoire compiler.

## Architecture

### Components

1. **Frontend (Client-side)**
   - HTML5 Canvas for drawing
   - JavaScript for interactivity
   - CSS for styling
   - No external dependencies (pure vanilla JS)

2. **Backend (Server)**
   - Simple HTTP server (Go)
   - Serves static files
   - No database required

3. **Core Modules**
   - **Canvas Manager**: Handles drawing operations
   - **Symbol Library**: Contains all Grimoire symbols
   - **Tool Manager**: Manages drawing tools
   - **Export Manager**: Handles PNG export
   - **Connection Manager**: Manages connections between symbols

### Directory Structure
```
editor/
├── DESIGN.md           # This document
├── README.md           # User documentation
├── server.go           # HTTP server
├── index.html          # Main HTML file
├── css/
│   └── style.css      # Styling
├── js/
│   ├── app.js         # Main application
│   ├── canvas.js      # Canvas management
│   ├── symbols.js     # Symbol definitions
│   ├── tools.js       # Drawing tools
│   ├── connections.js # Connection management
│   └── export.js      # Export functionality
└── examples/          # Example magic circles
```

## Features

### Core Features
1. **Symbol Palette**
   - All Grimoire symbols organized by category
   - Drag-and-drop to canvas
   - Visual preview

2. **Drawing Tools**
   - Selection tool (move, resize, delete)
   - Circle tool (for outer circles)
   - Connection tool (for energy flows)
   - Text tool (for labels)

3. **Canvas Operations**
   - Grid snapping
   - Zoom and pan
   - Undo/redo
   - Clear canvas

4. **Export Options**
   - Export to PNG
   - Save/load projects (JSON format)
   - Copy to clipboard

### Symbol Categories

Based on the Grimoire language specification:

1. **Structure Elements**
   - Outer Circle (⭕)
   - Inner Circle (○)
   - Double Circle (◎)
   - Pentagram (⭐)
   - Hexagram (✡)
   - Octagram (✦)
   - Triangle (△)
   - Square (□)

2. **Mystical Operators**
   - Fusion (⟐) - Addition
   - Separation (⟑) - Subtraction
   - Amplification (✦) - Multiplication
   - Division (⟠) - Division
   - Transfer (⟷) - Assignment
   - Seal (⊗) - Constant
   - Cycle (⟳) - Loop

3. **Comparison Symbols**
   - Balance (⟨⟩) - Equality
   - Imbalance (⟨≠⟩) - Inequality
   - Descent (⟨<⟩) - Less than
   - Ascent (⟨>⟩) - Greater than
   - Earth (⟨≤⟩) - Less or equal
   - Heaven (⟨≥⟩) - Greater or equal

4. **Logic Symbols**
   - Light Union (⊕) - AND
   - Light Choice (⊖) - OR
   - Light Inversion (⊗) - NOT
   - Light Exclusion (⊙) - XOR

5. **Energy Nodes**
   - Hexagonal Crystal (⬢) - Branch point
   - Square Crystal (◈) - Aggregation point
   - Pentagon Crystal (⬟) - Transformation point
   - Star Crystal (✧) - Amplification point

6. **Special Symbols**
   - Sun (☀) - Start/True
   - Moon (☾) - False/Alternative
   - Star (☆) - Output/Result
   - Double Circle (○○) - Function call
   - Musical Note (♪) - Sound/Event
   - Envelope (✉) - Message/Communication
   - Checkmark (✓) - Success/Complete

## User Interface Design

### Layout
```
+--------------------------------------------------+
| Toolbar (File, Edit, View, Help)                |
+--------------------------------------------------+
| Tool    |                           | Symbol     |
| Panel   |     Canvas Area           | Palette    |
|         |                           |            |
| - Select|                           | Structures |
| - Circle|                           | Operators  |
| - Line  |                           | Logic      |
| - Text  |                           | Nodes      |
|         |                           | Special    |
+--------------------------------------------------+
| Status Bar (Zoom, Grid, Coordinates)             |
+--------------------------------------------------+
```

### Interaction Patterns

1. **Adding Symbols**
   - Click symbol in palette
   - Click on canvas to place
   - Or drag from palette to canvas

2. **Creating Connections**
   - Select connection tool
   - Click first symbol
   - Click second symbol
   - Connection automatically drawn

3. **Editing Elements**
   - Select tool to move/resize
   - Double-click to edit properties
   - Delete key to remove

4. **Canvas Navigation**
   - Mouse wheel to zoom
   - Middle mouse to pan
   - Spacebar + drag to pan

## Technical Implementation

### Canvas Rendering
- Use 2D context for drawing
- Implement layered rendering:
  1. Grid layer
  2. Connection layer
  3. Symbol layer
  4. Selection layer

### Data Model
```javascript
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
      "id": "elem1",
      "type": "circle",
      "symbol": "⭕",
      "x": 400,
      "y": 300,
      "width": 300,
      "height": 300,
      "properties": {}
    }
  ],
  "connections": [
    {
      "id": "conn1",
      "from": "elem1",
      "to": "elem2",
      "type": "energy_flow",
      "style": "solid"
    }
  ]
}
```

### Symbol Rendering
- Use Unicode characters for symbols
- Scale based on element size
- Support for custom fonts
- Fallback to image sprites if needed

## Extension Points

1. **Custom Symbols**
   - Plugin system for adding new symbols
   - Symbol definition format

2. **Export Formats**
   - Plugin for different export formats
   - Integration with Grimoire compiler

3. **Themes**
   - Support for different visual themes
   - Dark mode support

4. **Validation**
   - Basic structure validation
   - Integration with compiler for syntax checking

## Performance Considerations

1. **Rendering Optimization**
   - Only redraw changed areas
   - Use requestAnimationFrame
   - Implement viewport culling

2. **Memory Management**
   - Limit undo history
   - Efficient symbol caching
   - Clean up unused resources

## Future Enhancements

1. **Collaboration**
   - Real-time collaborative editing
   - Share via URL

2. **Animation**
   - Animate energy flows
   - Step-through execution

3. **Templates**
   - Pre-built magic circle templates
   - Template library

4. **Mobile Support**
   - Touch gestures
   - Responsive design