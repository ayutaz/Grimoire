// Canvas management and rendering
class CanvasManager {
    constructor(canvasId) {
        this.canvas = document.getElementById(canvasId);
        this.ctx = this.canvas.getContext('2d');
        this.elements = [];
        this.connections = [];
        this.selectedElement = null;
        this.selectedConnection = null;
        this.gridSize = 20;
        this.showGrid = true;
        this.snapToGrid = true;
        this.zoom = 1;
        this.panX = 0;
        this.panY = 0;
        this.isDragging = false;
        this.dragStartX = 0;
        this.dragStartY = 0;
        this.elementIdCounter = 0;
        this.connectionIdCounter = 0;
        
        this.init();
    }
    
    init() {
        this.resize();
        window.addEventListener('resize', () => this.resize());
        this.setupMouseEvents();
        this.render();
    }
    
    resize() {
        const container = this.canvas.parentElement;
        this.canvas.width = container.clientWidth;
        this.canvas.height = container.clientHeight;
        this.render();
    }
    
    setupMouseEvents() {
        this.canvas.addEventListener('mousedown', (e) => this.handleMouseDown(e));
        this.canvas.addEventListener('mousemove', (e) => this.handleMouseMove(e));
        this.canvas.addEventListener('mouseup', (e) => this.handleMouseUp(e));
        this.canvas.addEventListener('wheel', (e) => this.handleWheel(e));
        this.canvas.addEventListener('contextmenu', (e) => e.preventDefault());
    }
    
    handleMouseDown(e) {
        const rect = this.canvas.getBoundingClientRect();
        const x = (e.clientX - rect.left) / this.zoom - this.panX;
        const y = (e.clientY - rect.top) / this.zoom - this.panY;
        
        if (e.button === 1 || (e.button === 0 && e.shiftKey)) {
            // Middle mouse or shift+left for panning
            this.isPanning = true;
            this.panStartX = e.clientX;
            this.panStartY = e.clientY;
            this.canvas.style.cursor = 'grabbing';
            return;
        }
        
        // Check if clicking on an element
        const clickedElement = this.getElementAt(x, y);
        if (clickedElement) {
            this.selectElement(clickedElement);
            this.isDragging = true;
            this.dragStartX = x - clickedElement.x;
            this.dragStartY = y - clickedElement.y;
        } else {
            // Check if clicking on a connection
            const clickedConnection = this.getConnectionAt(x, y);
            if (clickedConnection) {
                this.selectConnection(clickedConnection);
            } else {
                this.clearSelection();
            }
        }
        
        this.render();
    }
    
    handleMouseMove(e) {
        const rect = this.canvas.getBoundingClientRect();
        const x = (e.clientX - rect.left) / this.zoom - this.panX;
        const y = (e.clientY - rect.top) / this.zoom - this.panY;
        
        // Update status bar
        this.updateMousePosition(x, y);
        
        if (this.isPanning) {
            this.panX += (e.clientX - this.panStartX) / this.zoom;
            this.panY += (e.clientY - this.panStartY) / this.zoom;
            this.panStartX = e.clientX;
            this.panStartY = e.clientY;
            this.render();
            return;
        }
        
        if (this.isDragging && this.selectedElement) {
            let newX = x - this.dragStartX;
            let newY = y - this.dragStartY;
            
            if (this.snapToGrid) {
                newX = Math.round(newX / this.gridSize) * this.gridSize;
                newY = Math.round(newY / this.gridSize) * this.gridSize;
            }
            
            this.selectedElement.x = newX;
            this.selectedElement.y = newY;
            this.render();
        }
    }
    
    handleMouseUp(e) {
        this.isDragging = false;
        this.isPanning = false;
        this.canvas.style.cursor = 'crosshair';
    }
    
    handleWheel(e) {
        e.preventDefault();
        const rect = this.canvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        
        const delta = e.deltaY > 0 ? 0.9 : 1.1;
        const newZoom = Math.max(0.1, Math.min(5, this.zoom * delta));
        
        // Zoom towards mouse position
        this.panX = x / this.zoom - x / newZoom + this.panX;
        this.panY = y / this.zoom - y / newZoom + this.panY;
        this.zoom = newZoom;
        
        this.updateZoomLevel();
        this.render();
    }
    
    addElement(type, x, y) {
        const symbolInfo = window.symbolManager.getSymbol(type);
        if (!symbolInfo) return null;
        
        if (this.snapToGrid) {
            x = Math.round(x / this.gridSize) * this.gridSize;
            y = Math.round(y / this.gridSize) * this.gridSize;
        }
        
        const element = {
            id: `elem_${this.elementIdCounter++}`,
            type: type,
            symbol: symbolInfo.symbol,
            x: x,
            y: y,
            size: symbolInfo.defaultSize,
            rotation: 0,
            properties: {}
        };
        
        this.elements.push(element);
        this.selectElement(element);
        this.render();
        return element;
    }
    
    removeElement(element) {
        // Remove connections to this element
        this.connections = this.connections.filter(conn => 
            conn.from !== element.id && conn.to !== element.id
        );
        
        // Remove element
        const index = this.elements.indexOf(element);
        if (index > -1) {
            this.elements.splice(index, 1);
        }
        
        if (this.selectedElement === element) {
            this.selectedElement = null;
        }
        
        this.render();
    }
    
    addConnection(fromId, toId) {
        // Check if connection already exists
        const exists = this.connections.some(conn => 
            (conn.from === fromId && conn.to === toId) ||
            (conn.from === toId && conn.to === fromId)
        );
        
        if (exists) return null;
        
        const connection = {
            id: `conn_${this.connectionIdCounter++}`,
            from: fromId,
            to: toId,
            type: 'energy_flow',
            style: 'solid'
        };
        
        this.connections.push(connection);
        this.render();
        return connection;
    }
    
    removeConnection(connection) {
        const index = this.connections.indexOf(connection);
        if (index > -1) {
            this.connections.splice(index, 1);
        }
        
        if (this.selectedConnection === connection) {
            this.selectedConnection = null;
        }
        
        this.render();
    }
    
    getElementAt(x, y) {
        // Check in reverse order (top to bottom)
        for (let i = this.elements.length - 1; i >= 0; i--) {
            const elem = this.elements[i];
            const halfSize = elem.size / 2;
            
            if (x >= elem.x - halfSize && x <= elem.x + halfSize &&
                y >= elem.y - halfSize && y <= elem.y + halfSize) {
                return elem;
            }
        }
        return null;
    }
    
    getConnectionAt(x, y) {
        const threshold = 10; // pixels
        
        for (const conn of this.connections) {
            const from = this.getElementById(conn.from);
            const to = this.getElementById(conn.to);
            
            if (!from || !to) continue;
            
            const dist = this.pointToLineDistance(x, y, from.x, from.y, to.x, to.y);
            if (dist < threshold) {
                return conn;
            }
        }
        return null;
    }
    
    pointToLineDistance(px, py, x1, y1, x2, y2) {
        const A = px - x1;
        const B = py - y1;
        const C = x2 - x1;
        const D = y2 - y1;
        
        const dot = A * C + B * D;
        const lenSq = C * C + D * D;
        let param = -1;
        
        if (lenSq !== 0) {
            param = dot / lenSq;
        }
        
        let xx, yy;
        
        if (param < 0) {
            xx = x1;
            yy = y1;
        } else if (param > 1) {
            xx = x2;
            yy = y2;
        } else {
            xx = x1 + param * C;
            yy = y1 + param * D;
        }
        
        const dx = px - xx;
        const dy = py - yy;
        return Math.sqrt(dx * dx + dy * dy);
    }
    
    getElementById(id) {
        return this.elements.find(elem => elem.id === id);
    }
    
    selectElement(element) {
        this.selectedElement = element;
        this.selectedConnection = null;
    }
    
    selectConnection(connection) {
        this.selectedConnection = connection;
        this.selectedElement = null;
    }
    
    clearSelection() {
        this.selectedElement = null;
        this.selectedConnection = null;
    }
    
    render() {
        // Clear canvas
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
        
        // Save context
        this.ctx.save();
        
        // Apply zoom and pan
        this.ctx.translate(this.panX * this.zoom, this.panY * this.zoom);
        this.ctx.scale(this.zoom, this.zoom);
        
        // Draw grid
        if (this.showGrid) {
            this.drawGrid();
        }
        
        // Draw connections
        this.connections.forEach(conn => {
            this.drawConnection(conn);
        });
        
        // Draw elements
        this.elements.forEach(elem => {
            this.drawElement(elem);
        });
        
        // Restore context
        this.ctx.restore();
    }
    
    drawGrid() {
        this.ctx.strokeStyle = '#ddd';
        this.ctx.lineWidth = 0.5;
        
        const startX = Math.floor(-this.panX / this.gridSize) * this.gridSize;
        const startY = Math.floor(-this.panY / this.gridSize) * this.gridSize;
        const endX = startX + this.canvas.width / this.zoom + this.gridSize;
        const endY = startY + this.canvas.height / this.zoom + this.gridSize;
        
        // Vertical lines
        for (let x = startX; x < endX; x += this.gridSize) {
            this.ctx.beginPath();
            this.ctx.moveTo(x, startY);
            this.ctx.lineTo(x, endY);
            this.ctx.stroke();
        }
        
        // Horizontal lines
        for (let y = startY; y < endY; y += this.gridSize) {
            this.ctx.beginPath();
            this.ctx.moveTo(startX, y);
            this.ctx.lineTo(endX, y);
            this.ctx.stroke();
        }
    }
    
    drawElement(element) {
        const isSelected = element === this.selectedElement;
        
        this.ctx.save();
        this.ctx.translate(element.x, element.y);
        
        if (element.rotation) {
            this.ctx.rotate(element.rotation * Math.PI / 180);
        }
        
        // Draw background for circles
        if (window.symbolManager.isCircleType(element.type)) {
            this.ctx.strokeStyle = isSelected ? '#0066cc' : '#333';
            this.ctx.lineWidth = isSelected ? 3 : 2;
            this.ctx.beginPath();
            this.ctx.arc(0, 0, element.size / 2, 0, Math.PI * 2);
            this.ctx.stroke();
            
            if (element.type === 'double-circle') {
                this.ctx.beginPath();
                this.ctx.arc(0, 0, element.size / 2 - 10, 0, Math.PI * 2);
                this.ctx.stroke();
            }
        } else {
            // Draw symbol
            this.ctx.fillStyle = isSelected ? '#0066cc' : '#333';
            this.ctx.font = `${element.size}px Arial`;
            this.ctx.textAlign = 'center';
            this.ctx.textBaseline = 'middle';
            this.ctx.fillText(element.symbol, 0, 0);
        }
        
        // Draw selection indicator
        if (isSelected) {
            this.ctx.strokeStyle = '#0066cc';
            this.ctx.lineWidth = 2;
            this.ctx.setLineDash([5, 5]);
            this.ctx.strokeRect(
                -element.size / 2 - 5,
                -element.size / 2 - 5,
                element.size + 10,
                element.size + 10
            );
            this.ctx.setLineDash([]);
        }
        
        this.ctx.restore();
    }
    
    drawConnection(connection) {
        const from = this.getElementById(connection.from);
        const to = this.getElementById(connection.to);
        
        if (!from || !to) return;
        
        const isSelected = connection === this.selectedConnection;
        
        this.ctx.strokeStyle = isSelected ? '#0066cc' : '#333';
        this.ctx.lineWidth = isSelected ? 3 : 2;
        
        if (connection.style === 'dashed') {
            this.ctx.setLineDash([5, 5]);
        }
        
        this.ctx.beginPath();
        this.ctx.moveTo(from.x, from.y);
        
        // Draw curved line for better aesthetics
        const dx = to.x - from.x;
        const dy = to.y - from.y;
        const cx = from.x + dx / 2;
        const cy = from.y + dy / 2;
        
        this.ctx.quadraticCurveTo(cx, cy - 30, to.x, to.y);
        this.ctx.stroke();
        
        // Draw arrow
        const angle = Math.atan2(to.y - cy + 30, to.x - cx);
        const arrowSize = 10;
        
        this.ctx.save();
        this.ctx.translate(to.x, to.y);
        this.ctx.rotate(angle);
        
        this.ctx.beginPath();
        this.ctx.moveTo(0, 0);
        this.ctx.lineTo(-arrowSize, -arrowSize / 2);
        this.ctx.lineTo(-arrowSize, arrowSize / 2);
        this.ctx.closePath();
        this.ctx.fill();
        
        this.ctx.restore();
        this.ctx.setLineDash([]);
    }
    
    updateMousePosition(x, y) {
        const mousePos = document.getElementById('mousePos');
        if (mousePos) {
            mousePos.textContent = `X: ${Math.round(x)}, Y: ${Math.round(y)}`;
        }
    }
    
    updateZoomLevel() {
        const zoomLevel = document.getElementById('zoomLevel');
        if (zoomLevel) {
            zoomLevel.textContent = `Zoom: ${Math.round(this.zoom * 100)}%`;
        }
    }
    
    clear() {
        this.elements = [];
        this.connections = [];
        this.selectedElement = null;
        this.selectedConnection = null;
        this.render();
    }
    
    toJSON() {
        return {
            version: '1.0',
            canvas: {
                width: this.canvas.width,
                height: this.canvas.height,
                grid: this.showGrid,
                gridSize: this.gridSize
            },
            elements: this.elements,
            connections: this.connections
        };
    }
    
    fromJSON(data) {
        if (data.version !== '1.0') {
            console.warn('Unsupported version:', data.version);
        }
        
        this.clear();
        this.elements = data.elements || [];
        this.connections = data.connections || [];
        this.elementIdCounter = Math.max(...this.elements.map(e => 
            parseInt(e.id.split('_')[1]) || 0
        ), 0) + 1;
        this.connectionIdCounter = Math.max(...this.connections.map(c => 
            parseInt(c.id.split('_')[1]) || 0
        ), 0) + 1;
        
        this.render();
    }
}

// Create global instance
window.canvasManager = null; // Will be initialized in app.js