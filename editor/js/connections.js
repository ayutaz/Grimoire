// Connection management and rendering
class ConnectionManager {
    constructor(canvasManager) {
        this.canvas = canvasManager;
        this.connectionStyles = {
            'energy_flow': {
                name: 'Energy Flow',
                strokeStyle: '#333',
                lineWidth: 2,
                dashArray: []
            },
            'conditional': {
                name: 'Conditional',
                strokeStyle: '#666',
                lineWidth: 2,
                dashArray: [5, 5]
            },
            'spirit_flow': {
                name: 'Spirit Flow',
                strokeStyle: '#9370DB',
                lineWidth: 2,
                dashArray: [2, 4]
            },
            'bidirectional': {
                name: 'Bidirectional',
                strokeStyle: '#333',
                lineWidth: 3,
                dashArray: [],
                arrows: 'both'
            },
            'loop_back': {
                name: 'Loop Back',
                strokeStyle: '#FF6347',
                lineWidth: 2,
                dashArray: [3, 3],
                curved: true
            }
        };
    }
    
    getStyle(type) {
        return this.connectionStyles[type] || this.connectionStyles['energy_flow'];
    }
    
    drawConnection(ctx, connection, from, to, isSelected = false) {
        const style = this.getStyle(connection.type);
        
        ctx.save();
        
        // Set style
        ctx.strokeStyle = isSelected ? '#0066cc' : style.strokeStyle;
        ctx.lineWidth = isSelected ? style.lineWidth + 1 : style.lineWidth;
        ctx.setLineDash(style.dashArray);
        
        // Calculate connection points on element boundaries
        const fromPoint = this.getConnectionPoint(from, to);
        const toPoint = this.getConnectionPoint(to, from);
        
        ctx.beginPath();
        ctx.moveTo(fromPoint.x, fromPoint.y);
        
        if (style.curved || connection.type === 'loop_back') {
            // Draw curved connection
            const dx = toPoint.x - fromPoint.x;
            const dy = toPoint.y - fromPoint.y;
            const dist = Math.sqrt(dx * dx + dy * dy);
            const curve = Math.min(50, dist * 0.3);
            
            // Control points for bezier curve
            const angle = Math.atan2(dy, dx) + Math.PI / 2;
            const cx1 = fromPoint.x + dx * 0.25 + Math.cos(angle) * curve;
            const cy1 = fromPoint.y + dy * 0.25 + Math.sin(angle) * curve;
            const cx2 = fromPoint.x + dx * 0.75 + Math.cos(angle) * curve;
            const cy2 = fromPoint.y + dy * 0.75 + Math.sin(angle) * curve;
            
            ctx.bezierCurveTo(cx1, cy1, cx2, cy2, toPoint.x, toPoint.y);
        } else {
            // Draw straight connection
            ctx.lineTo(toPoint.x, toPoint.y);
        }
        
        ctx.stroke();
        
        // Draw arrows
        if (style.arrows === 'both') {
            this.drawArrow(ctx, fromPoint, toPoint, true);
            this.drawArrow(ctx, toPoint, fromPoint, true);
        } else {
            this.drawArrow(ctx, fromPoint, toPoint, false);
        }
        
        ctx.restore();
    }
    
    getConnectionPoint(from, to) {
        // Get the point on the boundary of 'from' element closest to 'to' element
        const dx = to.x - from.x;
        const dy = to.y - from.y;
        const angle = Math.atan2(dy, dx);
        
        // For circle elements
        if (window.symbolManager.isCircleType(from.type)) {
            const radius = from.size / 2;
            return {
                x: from.x + Math.cos(angle) * radius,
                y: from.y + Math.sin(angle) * radius
            };
        }
        
        // For rectangular elements (simplified)
        const halfSize = from.size / 2;
        const corners = [
            { x: from.x - halfSize, y: from.y - halfSize },
            { x: from.x + halfSize, y: from.y - halfSize },
            { x: from.x + halfSize, y: from.y + halfSize },
            { x: from.x - halfSize, y: from.y + halfSize }
        ];
        
        // Find intersection with rectangle edges
        // Simplified: just use center for now
        return { x: from.x, y: from.y };
    }
    
    drawArrow(ctx, from, to, reverse = false) {
        const angle = Math.atan2(to.y - from.y, to.x - from.x);
        const arrowSize = 10;
        
        ctx.save();
        
        if (reverse) {
            ctx.translate(from.x, from.y);
            ctx.rotate(angle + Math.PI);
        } else {
            ctx.translate(to.x, to.y);
            ctx.rotate(angle);
        }
        
        ctx.fillStyle = ctx.strokeStyle;
        ctx.beginPath();
        ctx.moveTo(0, 0);
        ctx.lineTo(-arrowSize, -arrowSize / 2);
        ctx.lineTo(-arrowSize, arrowSize / 2);
        ctx.closePath();
        ctx.fill();
        
        ctx.restore();
    }
    
    createAutoConnection(elements) {
        // Auto-connect elements based on proximity and type
        const connections = [];
        
        for (let i = 0; i < elements.length - 1; i++) {
            for (let j = i + 1; j < elements.length; j++) {
                const elem1 = elements[i];
                const elem2 = elements[j];
                
                // Skip if either is an outer circle
                if (!window.symbolManager.isConnectionPoint(elem1.type) ||
                    !window.symbolManager.isConnectionPoint(elem2.type)) {
                    continue;
                }
                
                // Check distance
                const dx = elem2.x - elem1.x;
                const dy = elem2.y - elem1.y;
                const dist = Math.sqrt(dx * dx + dy * dy);
                
                // Connect if close enough
                if (dist < 150) {
                    connections.push({
                        from: elem1.id,
                        to: elem2.id,
                        type: 'energy_flow'
                    });
                }
            }
        }
        
        return connections;
    }
    
    validateConnection(from, to) {
        // Validation rules for connections
        if (from === to) return false;
        
        const fromElem = this.canvas.getElementById(from);
        const toElem = this.canvas.getElementById(to);
        
        if (!fromElem || !toElem) return false;
        
        // Can't connect to outer circles
        if (!window.symbolManager.isConnectionPoint(fromElem.type) ||
            !window.symbolManager.isConnectionPoint(toElem.type)) {
            return false;
        }
        
        // Check for duplicate connections
        const exists = this.canvas.connections.some(conn =>
            (conn.from === from && conn.to === to) ||
            (conn.from === to && conn.to === from)
        );
        
        return !exists;
    }
}

// Connection context menu
function initConnectionContextMenu(canvasManager) {
    let contextMenu = null;
    
    canvasManager.canvas.addEventListener('contextmenu', (e) => {
        e.preventDefault();
        
        const rect = canvasManager.canvas.getBoundingClientRect();
        const x = (e.clientX - rect.left) / canvasManager.zoom - canvasManager.panX;
        const y = (e.clientY - rect.top) / canvasManager.zoom - canvasManager.panY;
        
        const connection = canvasManager.getConnectionAt(x, y);
        if (!connection) return;
        
        // Remove existing menu
        if (contextMenu) {
            contextMenu.remove();
        }
        
        // Create context menu
        contextMenu = document.createElement('div');
        contextMenu.className = 'context-menu';
        contextMenu.style.left = e.clientX + 'px';
        contextMenu.style.top = e.clientY + 'px';
        
        const connectionManager = new ConnectionManager(canvasManager);
        
        // Menu items
        Object.entries(connectionManager.connectionStyles).forEach(([type, style]) => {
            const item = document.createElement('div');
            item.className = 'context-menu-item';
            item.textContent = style.name;
            if (connection.type === type) {
                item.style.fontWeight = 'bold';
            }
            
            item.addEventListener('click', () => {
                connection.type = type;
                canvasManager.render();
                contextMenu.remove();
            });
            
            contextMenu.appendChild(item);
        });
        
        // Delete option
        const deleteItem = document.createElement('div');
        deleteItem.className = 'context-menu-item';
        deleteItem.textContent = 'Delete';
        deleteItem.style.color = '#ff4444';
        
        deleteItem.addEventListener('click', () => {
            canvasManager.removeConnection(connection);
            contextMenu.remove();
        });
        
        contextMenu.appendChild(deleteItem);
        
        document.body.appendChild(contextMenu);
        
        // Remove menu on click outside
        const removeMenu = (e) => {
            if (!contextMenu.contains(e.target)) {
                contextMenu.remove();
                document.removeEventListener('click', removeMenu);
            }
        };
        
        setTimeout(() => {
            document.addEventListener('click', removeMenu);
        }, 0);
    });
}

// Export for use in other modules
window.ConnectionManager = ConnectionManager;
window.initConnectionContextMenu = initConnectionContextMenu;