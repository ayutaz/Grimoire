// Main application entry point
document.addEventListener('DOMContentLoaded', () => {
    // Initialize canvas manager
    window.canvasManager = new CanvasManager('canvas');
    
    // Initialize tool manager
    window.toolManager = new ToolManager(window.canvasManager);
    
    // Initialize export manager
    const exportManager = new ExportManager(window.canvasManager);
    
    // Initialize drag and drop
    initSymbolDragDrop(window.canvasManager);
    
    // Initialize connection context menu
    initConnectionContextMenu(window.canvasManager);
    
    // Setup toolbar buttons
    setupToolbarButtons(exportManager);
    
    // Setup keyboard shortcuts
    setupKeyboardShortcuts();
    
    // Load example if requested
    const urlParams = new URLSearchParams(window.location.search);
    const example = urlParams.get('example');
    if (example) {
        loadExample(example);
    }
    
    // Update status bar
    updateStatusBar();
});

function setupToolbarButtons(exportManager) {
    // New button
    document.getElementById('newBtn').addEventListener('click', () => {
        if (confirm('Clear canvas and start new project?')) {
            window.canvasManager.clear();
        }
    });
    
    // Save button
    document.getElementById('saveBtn').addEventListener('click', () => {
        exportManager.saveProject();
    });
    
    // Load button
    const loadBtn = document.getElementById('loadBtn');
    const loadInput = document.getElementById('loadInput');
    
    loadBtn.addEventListener('click', () => {
        loadInput.click();
    });
    
    loadInput.addEventListener('change', (e) => {
        if (e.target.files.length > 0) {
            exportManager.loadProject(e.target.files[0]);
        }
    });
    
    // Export button
    document.getElementById('exportBtn').addEventListener('click', () => {
        exportManager.exportToPNG();
    });
    
    // Undo/Redo (placeholder)
    document.getElementById('undoBtn').addEventListener('click', () => {
        console.log('Undo not yet implemented');
    });
    
    document.getElementById('redoBtn').addEventListener('click', () => {
        console.log('Redo not yet implemented');
    });
    
    // Clear button
    document.getElementById('clearBtn').addEventListener('click', () => {
        if (confirm('Clear all elements?')) {
            window.canvasManager.clear();
        }
    });
    
    // Grid toggle
    document.getElementById('gridToggle').addEventListener('change', (e) => {
        window.canvasManager.showGrid = e.target.checked;
        window.canvasManager.render();
    });
    
    // Snap toggle
    document.getElementById('snapToggle').addEventListener('change', (e) => {
        window.canvasManager.snapToGrid = e.target.checked;
    });
}

function setupKeyboardShortcuts() {
    document.addEventListener('keydown', (e) => {
        // Ctrl/Cmd shortcuts
        if (e.ctrlKey || e.metaKey) {
            switch(e.key.toLowerCase()) {
                case 's':
                    e.preventDefault();
                    document.getElementById('saveBtn').click();
                    break;
                case 'o':
                    e.preventDefault();
                    document.getElementById('loadBtn').click();
                    break;
                case 'e':
                    e.preventDefault();
                    document.getElementById('exportBtn').click();
                    break;
                case 'z':
                    e.preventDefault();
                    document.getElementById('undoBtn').click();
                    break;
                case 'y':
                    e.preventDefault();
                    document.getElementById('redoBtn').click();
                    break;
            }
        }
        
        // Escape to deselect
        if (e.key === 'Escape') {
            window.canvasManager.clearSelection();
            window.canvasManager.render();
        }
    });
}

function updateStatusBar() {
    // Update grid size display
    const gridSizeElem = document.getElementById('gridSize');
    if (gridSizeElem) {
        gridSizeElem.textContent = `Grid: ${window.canvasManager.gridSize}px`;
    }
    
    // Initial zoom level
    window.canvasManager.updateZoomLevel();
}

function loadExample(name) {
    // Example magic circles
    const examples = {
        'hello-world': {
            elements: [
                {
                    id: 'elem_0',
                    type: 'outer-circle',
                    symbol: '⭕',
                    x: 400,
                    y: 300,
                    size: 300,
                    rotation: 0,
                    properties: {}
                },
                {
                    id: 'elem_1',
                    type: 'star',
                    symbol: '☆',
                    x: 400,
                    y: 300,
                    size: 60,
                    rotation: 0,
                    properties: {}
                }
            ],
            connections: []
        },
        'conditional': {
            elements: [
                {
                    id: 'elem_0',
                    type: 'outer-circle',
                    symbol: '⭕',
                    x: 400,
                    y: 300,
                    size: 300,
                    rotation: 0,
                    properties: {}
                },
                {
                    id: 'elem_1',
                    type: 'triangle',
                    symbol: '△',
                    x: 400,
                    y: 250,
                    size: 60,
                    rotation: 0,
                    properties: {}
                },
                {
                    id: 'elem_2',
                    type: 'sun',
                    symbol: '☀',
                    x: 350,
                    y: 320,
                    size: 50,
                    rotation: 0,
                    properties: {}
                },
                {
                    id: 'elem_3',
                    type: 'moon',
                    symbol: '☾',
                    x: 450,
                    y: 320,
                    size: 50,
                    rotation: 0,
                    properties: {}
                }
            ],
            connections: [
                {
                    id: 'conn_0',
                    from: 'elem_1',
                    to: 'elem_2',
                    type: 'conditional',
                    style: 'dashed'
                },
                {
                    id: 'conn_1',
                    from: 'elem_1',
                    to: 'elem_3',
                    type: 'conditional',
                    style: 'dashed'
                }
            ]
        },
        'function': {
            elements: [
                {
                    id: 'elem_0',
                    type: 'outer-circle',
                    symbol: '⭕',
                    x: 400,
                    y: 300,
                    size: 300,
                    rotation: 0,
                    properties: {}
                },
                {
                    id: 'elem_1',
                    type: 'inner-circle',
                    symbol: '○',
                    x: 400,
                    y: 250,
                    size: 100,
                    rotation: 0,
                    properties: {}
                },
                {
                    id: 'elem_2',
                    type: 'data-node',
                    symbol: '□•',
                    x: 350,
                    y: 250,
                    size: 40,
                    rotation: 0,
                    properties: {}
                },
                {
                    id: 'elem_3',
                    type: 'data-node',
                    symbol: '□•',
                    x: 450,
                    y: 250,
                    size: 40,
                    rotation: 0,
                    properties: {}
                },
                {
                    id: 'elem_4',
                    type: 'plus',
                    symbol: '+',
                    x: 400,
                    y: 300,
                    size: 40,
                    rotation: 0,
                    properties: {}
                },
                {
                    id: 'elem_5',
                    type: 'star',
                    symbol: '☆',
                    x: 400,
                    y: 350,
                    size: 50,
                    rotation: 0,
                    properties: {}
                }
            ],
            connections: [
                {
                    id: 'conn_0',
                    from: 'elem_2',
                    to: 'elem_4',
                    type: 'energy_flow',
                    style: 'solid'
                },
                {
                    id: 'conn_1',
                    from: 'elem_3',
                    to: 'elem_4',
                    type: 'energy_flow',
                    style: 'solid'
                },
                {
                    id: 'conn_2',
                    from: 'elem_4',
                    to: 'elem_5',
                    type: 'energy_flow',
                    style: 'solid'
                }
            ]
        }
    };
    
    const example = examples[name];
    if (example) {
        window.canvasManager.fromJSON({
            version: '1.0',
            canvas: {
                width: 800,
                height: 600,
                grid: true,
                gridSize: 20
            },
            elements: example.elements,
            connections: example.connections
        });
    }
}

// Helper function for creating preset patterns
window.createPattern = function(pattern) {
    const cm = window.canvasManager;
    const centerX = 400;
    const centerY = 300;
    
    switch(pattern) {
        case 'pentagram':
            // Create pentagram pattern
            cm.clear();
            cm.addElement('outer-circle', centerX, centerY);
            
            const pentaRadius = 100;
            for (let i = 0; i < 5; i++) {
                const angle = (i * 72 - 90) * Math.PI / 180;
                const x = centerX + Math.cos(angle) * pentaRadius;
                const y = centerY + Math.sin(angle) * pentaRadius;
                cm.addElement('star', x, y);
            }
            break;
            
        case 'hexagram':
            // Create hexagram pattern
            cm.clear();
            cm.addElement('outer-circle', centerX, centerY);
            cm.addElement('hexagram', centerX, centerY);
            
            const hexRadius = 80;
            for (let i = 0; i < 6; i++) {
                const angle = (i * 60) * Math.PI / 180;
                const x = centerX + Math.cos(angle) * hexRadius;
                const y = centerY + Math.sin(angle) * hexRadius;
                cm.addElement('hex-crystal', x, y);
            }
            break;
            
        case 'trinity':
            // Create trinity pattern
            cm.clear();
            cm.addElement('outer-circle', centerX, centerY);
            cm.addElement('triangle', centerX, centerY);
            
            const triRadius = 80;
            for (let i = 0; i < 3; i++) {
                const angle = (i * 120 - 90) * Math.PI / 180;
                const x = centerX + Math.cos(angle) * triRadius;
                const y = centerY + Math.sin(angle) * triRadius;
                cm.addElement('inner-circle', x, y);
            }
            break;
    }
};

// Debug helper
window.debugCanvas = function() {
    console.log('Elements:', window.canvasManager.elements);
    console.log('Connections:', window.canvasManager.connections);
    console.log('Canvas state:', window.canvasManager.toJSON());
};