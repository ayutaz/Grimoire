// Export functionality
class ExportManager {
    constructor(canvasManager) {
        this.canvas = canvasManager;
    }
    
    exportToPNG() {
        // Create a temporary canvas for export
        const exportCanvas = document.createElement('canvas');
        const exportCtx = exportCanvas.getContext('2d');
        
        // Calculate bounds of all elements
        const bounds = this.calculateBounds();
        if (!bounds) {
            alert('No elements to export');
            return;
        }
        
        // Add padding
        const padding = 50;
        const width = bounds.maxX - bounds.minX + padding * 2;
        const height = bounds.maxY - bounds.minY + padding * 2;
        
        exportCanvas.width = width;
        exportCanvas.height = height;
        
        // White background
        exportCtx.fillStyle = '#ffffff';
        exportCtx.fillRect(0, 0, width, height);
        
        // Translate to center content
        exportCtx.translate(-bounds.minX + padding, -bounds.minY + padding);
        
        // Temporarily modify canvas manager for export
        const originalCanvas = this.canvas.canvas;
        const originalCtx = this.canvas.ctx;
        const originalShowGrid = this.canvas.showGrid;
        
        this.canvas.canvas = exportCanvas;
        this.canvas.ctx = exportCtx;
        this.canvas.showGrid = false;
        
        // Render without grid and selection
        const originalSelected = this.canvas.selectedElement;
        const originalSelectedConn = this.canvas.selectedConnection;
        this.canvas.selectedElement = null;
        this.canvas.selectedConnection = null;
        
        // Draw connections
        this.canvas.connections.forEach(conn => {
            this.canvas.drawConnection(conn);
        });
        
        // Draw elements
        this.canvas.elements.forEach(elem => {
            this.canvas.drawElement(elem);
        });
        
        // Restore original state
        this.canvas.canvas = originalCanvas;
        this.canvas.ctx = originalCtx;
        this.canvas.showGrid = originalShowGrid;
        this.canvas.selectedElement = originalSelected;
        this.canvas.selectedConnection = originalSelectedConn;
        
        // Convert to PNG and download
        exportCanvas.toBlob((blob) => {
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `magic_circle_${Date.now()}.png`;
            a.click();
            URL.revokeObjectURL(url);
        }, 'image/png');
    }
    
    calculateBounds() {
        if (this.canvas.elements.length === 0) return null;
        
        let minX = Infinity, minY = Infinity;
        let maxX = -Infinity, maxY = -Infinity;
        
        this.canvas.elements.forEach(elem => {
            const halfSize = elem.size / 2;
            minX = Math.min(minX, elem.x - halfSize);
            minY = Math.min(minY, elem.y - halfSize);
            maxX = Math.max(maxX, elem.x + halfSize);
            maxY = Math.max(maxY, elem.y + halfSize);
        });
        
        return { minX, minY, maxX, maxY };
    }
    
    saveProject() {
        const data = this.canvas.toJSON();
        const json = JSON.stringify(data, null, 2);
        const blob = new Blob([json], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        
        const a = document.createElement('a');
        a.href = url;
        a.download = `grimoire_project_${Date.now()}.json`;
        a.click();
        URL.revokeObjectURL(url);
    }
    
    loadProject(file) {
        const reader = new FileReader();
        
        reader.onload = (e) => {
            try {
                const data = JSON.parse(e.target.result);
                this.canvas.fromJSON(data);
            } catch (error) {
                alert('Error loading project: ' + error.message);
            }
        };
        
        reader.readAsText(file);
    }
    
    copyToClipboard() {
        this.exportToPNG();
        // Note: Direct clipboard API for images requires additional implementation
        // For now, we just export to file
    }
    
    exportToGrimoire() {
        // Convert visual representation to Grimoire code
        // This is a placeholder for future implementation
        const code = this.generateGrimoireCode();
        
        const blob = new Blob([code], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        
        const a = document.createElement('a');
        a.href = url;
        a.download = `magic_circle_${Date.now()}.grim`;
        a.click();
        URL.revokeObjectURL(url);
    }
    
    generateGrimoireCode() {
        // Simple code generation based on visual elements
        let code = '# Generated Grimoire Magic Circle\n\n';
        
        // Find outer circle
        const outerCircle = this.canvas.elements.find(e => e.type === 'outer-circle');
        if (!outerCircle) {
            code += '# Warning: No outer circle found\n\n';
        }
        
        // Group elements by position
        const center = this.findCenter();
        const centerElements = [];
        const peripheralElements = [];
        
        this.canvas.elements.forEach(elem => {
            if (elem.type === 'outer-circle') return;
            
            const dist = Math.sqrt(
                Math.pow(elem.x - center.x, 2) + 
                Math.pow(elem.y - center.y, 2)
            );
            
            if (dist < 100) {
                centerElements.push(elem);
            } else {
                peripheralElements.push(elem);
            }
        });
        
        // Generate structure
        code += '⭕[\n';
        
        if (centerElements.length > 0) {
            code += '  # Central elements\n';
            centerElements.forEach(elem => {
                code += `  ${elem.symbol}`;
                if (elem.text) code += ` "${elem.text}"`;
                code += '\n';
            });
        }
        
        if (peripheralElements.length > 0) {
            code += '  \n  # Peripheral elements\n';
            peripheralElements.forEach(elem => {
                code += `  ${elem.symbol}`;
                if (elem.text) code += ` "${elem.text}"`;
                code += '\n';
            });
        }
        
        code += ']\n';
        
        return code;
    }
    
    findCenter() {
        if (this.canvas.elements.length === 0) {
            return { x: 400, y: 300 };
        }
        
        let sumX = 0, sumY = 0;
        this.canvas.elements.forEach(elem => {
            sumX += elem.x;
            sumY += elem.y;
        });
        
        return {
            x: sumX / this.canvas.elements.length,
            y: sumY / this.canvas.elements.length
        };
    }
}

// Create global instance
window.ExportManager = ExportManager;