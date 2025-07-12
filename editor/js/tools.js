// Tool management
class ToolManager {
    constructor(canvasManager) {
        this.canvas = canvasManager;
        this.currentTool = 'select';
        this.tempConnection = null;
        this.connectStart = null;
        
        this.initTools();
    }
    
    initTools() {
        // Tool buttons
        document.querySelectorAll('.tool').forEach(btn => {
            btn.addEventListener('click', (e) => {
                this.setTool(btn.dataset.tool);
            });
        });
        
        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            switch(e.key.toLowerCase()) {
                case 'v':
                    this.setTool('select');
                    break;
                case 'c':
                    this.setTool('circle');
                    break;
                case 'l':
                    this.setTool('connect');
                    break;
                case 't':
                    this.setTool('text');
                    break;
                case 'delete':
                case 'backspace':
                    if (this.currentTool === 'select') {
                        this.deleteSelected();
                    }
                    break;
            }
        });
        
        // Canvas tool events
        this.canvas.canvas.addEventListener('click', (e) => this.handleCanvasClick(e));
    }
    
    setTool(tool) {
        this.currentTool = tool;
        
        // Update UI
        document.querySelectorAll('.tool').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.tool === tool);
        });
        
        // Update cursor
        switch(tool) {
            case 'select':
                this.canvas.canvas.style.cursor = 'default';
                break;
            case 'circle':
                this.canvas.canvas.style.cursor = 'crosshair';
                break;
            case 'connect':
                this.canvas.canvas.style.cursor = 'crosshair';
                break;
            case 'text':
                this.canvas.canvas.style.cursor = 'text';
                break;
        }
        
        // Reset connection tool state
        if (tool !== 'connect') {
            this.connectStart = null;
        }
    }
    
    handleCanvasClick(e) {
        if (e.defaultPrevented) return;
        
        const rect = this.canvas.canvas.getBoundingClientRect();
        const x = (e.clientX - rect.left) / this.canvas.zoom - this.canvas.panX;
        const y = (e.clientY - rect.top) / this.canvas.zoom - this.canvas.panY;
        
        switch(this.currentTool) {
            case 'circle':
                this.createCircle(x, y);
                break;
            case 'connect':
                this.handleConnect(x, y);
                break;
            case 'text':
                this.createText(x, y);
                break;
        }
    }
    
    createCircle(x, y) {
        // Simple circle creation - creates an outer circle
        this.canvas.addElement('outer-circle', x, y);
    }
    
    handleConnect(x, y) {
        const element = this.canvas.getElementAt(x, y);
        
        if (!element) {
            // Clicked on empty space, cancel connection
            this.connectStart = null;
            return;
        }
        
        if (!window.symbolManager.isConnectionPoint(element.type)) {
            // Can't connect to this element type
            return;
        }
        
        if (!this.connectStart) {
            // Start connection
            this.connectStart = element;
        } else if (this.connectStart !== element) {
            // Complete connection
            this.canvas.addConnection(this.connectStart.id, element.id);
            this.connectStart = null;
        }
    }
    
    createText(x, y) {
        const text = prompt('Enter text:');
        if (text) {
            // For now, create a custom element with text
            // In a full implementation, this would be a proper text element
            const element = this.canvas.addElement('square', x, y);
            if (element) {
                element.text = text;
                this.canvas.render();
            }
        }
    }
    
    deleteSelected() {
        if (this.canvas.selectedElement) {
            this.canvas.removeElement(this.canvas.selectedElement);
        } else if (this.canvas.selectedConnection) {
            this.canvas.removeConnection(this.canvas.selectedConnection);
        }
    }
}

// Symbol palette drag and drop
function initSymbolDragDrop(canvasManager) {
    const symbols = document.querySelectorAll('.symbol');
    let dragGhost = null;
    
    symbols.forEach(symbol => {
        symbol.addEventListener('dragstart', (e) => {
            e.dataTransfer.effectAllowed = 'copy';
            e.dataTransfer.setData('symbol-type', symbol.dataset.type);
            
            // Create custom drag ghost
            dragGhost = document.createElement('div');
            dragGhost.className = 'drag-ghost';
            dragGhost.textContent = symbol.dataset.symbol;
            dragGhost.style.left = '-1000px';
            document.body.appendChild(dragGhost);
            e.dataTransfer.setDragImage(dragGhost, 20, 20);
        });
        
        symbol.addEventListener('dragend', () => {
            if (dragGhost) {
                dragGhost.remove();
                dragGhost = null;
            }
        });
        
        // Alternative: click to select, then click on canvas
        symbol.addEventListener('click', () => {
            // Visual feedback
            symbols.forEach(s => s.classList.remove('selected'));
            symbol.classList.add('selected');
            
            // Set up canvas for placement
            const handleCanvasClick = (e) => {
                const rect = canvasManager.canvas.getBoundingClientRect();
                const x = (e.clientX - rect.left) / canvasManager.zoom - canvasManager.panX;
                const y = (e.clientY - rect.top) / canvasManager.zoom - canvasManager.panY;
                
                canvasManager.addElement(symbol.dataset.type, x, y);
                
                // Clean up
                symbol.classList.remove('selected');
                canvasManager.canvas.removeEventListener('click', handleCanvasClick);
            };
            
            canvasManager.canvas.addEventListener('click', handleCanvasClick, { once: true });
        });
    });
    
    // Canvas drop handling
    const canvas = canvasManager.canvas;
    
    canvas.addEventListener('dragover', (e) => {
        e.preventDefault();
        e.dataTransfer.dropEffect = 'copy';
    });
    
    canvas.addEventListener('drop', (e) => {
        e.preventDefault();
        
        const type = e.dataTransfer.getData('symbol-type');
        if (!type) return;
        
        const rect = canvas.getBoundingClientRect();
        const x = (e.clientX - rect.left) / canvasManager.zoom - canvasManager.panX;
        const y = (e.clientY - rect.top) / canvasManager.zoom - canvasManager.panY;
        
        canvasManager.addElement(type, x, y);
    });
}

// Create global instance
window.toolManager = null; // Will be initialized in app.js